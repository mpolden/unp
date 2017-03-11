package dispatcher

import (
	"fmt"
	"path/filepath"
)

type Path struct {
	Name          string
	MaxDepth      int
	MinDepth      int
	SkipHidden    bool
	Patterns      []string
	Remove        bool
	ArchiveExt    string
	UnpackCommand string
	PostCommand   string
}

func (p *Path) match(name string) (bool, error) {
	for _, pattern := range p.Patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, fmt.Errorf("%s: %s", err, pattern)
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
