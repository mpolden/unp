package dispatcher

import (
	"bytes"
	"fmt"
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
	PostCommand   string
}

type CmdValues struct {
	Name string
	Dir  string
	Base string
}

func (p *Path) Match(name string) (bool, error) {
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

func (p *Path) newCmd(tmpl string, v CmdValues) (*exec.Cmd, error) {
	t, err := template.New("cmd").Parse(tmpl)
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
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = v.Dir
	return cmd, nil
}

func (p *Path) NewUnpackCmd(v CmdValues) (*exec.Cmd, error) {
	return p.newCmd(p.UnpackCommand, v)
}

func (p *Path) NewPostCmd(v CmdValues) (*exec.Cmd, error) {
	return p.newCmd(p.PostCommand, v)
}

func (p *Path) ValidDepth(depth int) bool {
	return depth >= p.MinDepth && depth <= p.MaxDepth
}
