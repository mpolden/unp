package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
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
}

type CommandValues struct {
	Name string
	Dir  string
	Base string
}

func PathDepth(name string) int {
	name = filepath.Clean(name)
	return strings.Count(name, string(os.PathSeparator))
}

func IsHidden(name string) bool {
	return strings.HasPrefix(filepath.Base(name), ".")
}

func DirBase(name string) string {
	return filepath.Join(filepath.Base(filepath.Dir(name)),
		filepath.Base(name))
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

func (p *Path) ArchiveExtWithDot() string {
	if strings.HasPrefix(p.ArchiveExt, ".") {
		return p.ArchiveExt
	}
	return "." + p.ArchiveExt
}

func (p *Path) NewUnpackCommand(v CommandValues) (*exec.Cmd, error) {
	t, err := template.New("cmd").Parse(p.UnpackCommand)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, v); err != nil {
		return nil, err
	}
	argv := strings.Split(b.String(), " ")
	if len(argv) == 0 {
		return nil, fmt.Errorf("template compiled to empty command")
	}
	cmd := exec.Command(argv[0])
	cmd.Dir = v.Dir
	if len(argv) > 1 {
		cmd.Args = argv[1:]
	}
	return cmd, nil
}

func (p *Path) ValidDirDepth(depth int) bool {
	return depth < p.MaxDepth
}

func (p *Path) ValidDepth(depth int) bool {
	return depth >= p.MinDepth && depth <= p.MaxDepth
}
