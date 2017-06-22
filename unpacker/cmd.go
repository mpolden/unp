package unpacker

import (
	"bytes"

	"github.com/pkg/errors"

	"os/exec"
	"strings"
	"text/template"
)

type cmdValues struct {
	Name string
	Dir  string
	Base string
}

func newCmd(tmpl string, v cmdValues) (*exec.Cmd, error) {
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
		return nil, errors.New("template compiled to empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = v.Dir
	return cmd, nil
}
