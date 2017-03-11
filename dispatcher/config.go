package dispatcher

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	Default    Path
	BufferSize int
	Paths      []Path
	filename   string
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
	return cfg, nil
}

func ReadConfig(name string) (Config, error) {
	if name == "~/.gounpackrc" {
		home := os.Getenv("HOME")
		name = filepath.Join(home, ".gounpackrc")
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
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	// Set a default buffer size
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 100
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

func (c *Config) validate() error {
	for _, p := range c.Paths {
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
		if !strings.HasPrefix(p.ArchiveExt, ".") {
			return fmt.Errorf("file extension missing dot prefix: %s", p.ArchiveExt)
		}
		if _, err := p.match("foo.bar"); err != nil {
			return err
		}
		if err := isExecutable(p.UnpackCommand); err != nil {
			return err
		}
		if err := isExecutable(p.PostCommand); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) findPath(name string) (Path, bool) {
	for _, p := range c.Paths {
		if strings.HasPrefix(name, p.Name) {
			return p, true
		}
	}
	return Path{}, false
}
