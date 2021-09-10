package watcher

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mpolden/unp/rar"
)

type Config struct {
	Default    Path
	BufferSize int
	Paths      []Path
	filename   string
}

type Path struct {
	Name        string
	Handler     string
	handler     Handler
	MaxDepth    int
	MinDepth    int
	SkipHidden  bool
	Patterns    []string
	Remove      bool
	PostCommand string
}

func (p *Path) match(name string) (bool, error) {
	for _, pattern := range p.Patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, fmt.Errorf("%s: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (p *Path) validDepth(depth int) bool {
	return depth >= p.MinDepth && depth <= p.MaxDepth
}

func readConfig(r io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	// Unmarshal config and replace every path with the default one
	var defaults Config
	if err := json.Unmarshal(data, &defaults); err != nil {
		return Config{}, err
	}
	for i := range defaults.Paths {
		defaults.Paths[i] = defaults.Default
	}
	// Unmarshal config again, letting individual paths override the defaults
	cfg := defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	// Set a default buffer size
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1024
	}
	if err := cfg.load(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ReadConfig(name string) (Config, error) {
	if name == "~/.unprc" {
		home := os.Getenv("HOME")
		name = filepath.Join(home, ".unprc")
	}
	f, err := os.Open(name)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	cfg, err := readConfig(f)
	if err != nil {
		return Config{}, err
	}

	cfg.filename = name
	return cfg, nil
}

func isExecutable(s string) error {
	if s == "" {
		return nil
	}
	args := strings.Split(s, " ")
	_, err := exec.LookPath(args[0])
	return err
}

func (c *Config) JSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func (c *Config) load() error {
	for i, p := range c.Paths {
		fi, err := os.Stat(p.Name)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("not a directory: %s", p.Name)
		}
		if p.MinDepth > p.MaxDepth {
			return fmt.Errorf("min depth must be <= max depth")
		}
		if _, err := p.match("foo.bar"); err != nil {
			return err
		}
		if err := isExecutable(p.PostCommand); err != nil {
			return err
		}
		switch p.Handler {
		case "rar", "":
			c.Paths[i].handler = rar.NewHandler()
		case "script":
			c.Paths[i].handler = &scriptHandler{}
		default:
			return fmt.Errorf("invalid handler: %q", p.Handler)
		}
	}
	return nil
}

func (c *Config) findPath(prefix string) (Path, bool) {
	for _, p := range c.Paths {
		if strings.HasPrefix(prefix, p.Name) {
			return p, true
		}
	}
	return Path{}, false
}
