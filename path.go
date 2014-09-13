package main

import (
	"os"
	"path/filepath"
	"strings"
)

type Path struct {
	Name          string
	MaxDepth      int
	SkipHidden    bool
	Patterns      []string
	Remove        bool
	ArchiveExt    string
	UnpackCommand string
}

func PathDepth(name string) int {
	name = filepath.Clean(name)
	return strings.Count(name, string(os.PathSeparator))
}

func (p *Path) Match(name string) (bool, error) {
	for _, pattern := range p.Patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}
