package dispatcher

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
	PostCommand   string
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

func (p *Path) newCmd(tmpl string, v CommandValues) (*exec.Cmd, error) {
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

func (p *Path) NewUnpackCommand(v CommandValues) (*exec.Cmd, error) {
	return p.newCmd(p.UnpackCommand, v)
}

func (p *Path) NewPostCommand(v CommandValues) (*exec.Cmd, error) {
	return p.newCmd(p.PostCommand, v)
}

func (p *Path) ValidDepth(depth int) bool {
	return depth >= p.MinDepth && depth <= p.MaxDepth
}
