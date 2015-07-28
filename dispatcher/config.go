package dispatcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	Async bool
	Paths []Path
}

func ReadConfig(name string) (Config, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func isExecutable(s string) error {
	if s == "" {
		return nil
	}
	args := strings.Split(s, " ")
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func (c *Config) Validate() error {
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
		for _, pattern := range p.Patterns {
			if _, err := filepath.Match(pattern, "foo.bar"); err != nil {
				return fmt.Errorf("%s: %s", err, pattern)
			}
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

func (c *Config) FindPath(name string) (Path, bool) {
	for _, p := range c.Paths {
		if strings.HasPrefix(name, p.Name) {
			return p, true
		}
	}
	return Path{}, false
}
