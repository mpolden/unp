package executil

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"strings"
)

type CommandData struct {
	Base string
	Dir  string
	Name string
}

func compileCommand(tmpl string, data CommandData) (*exec.Cmd, error) {
	t, err := template.New("cmd").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, err
	}
	argv := strings.Split(b.String(), " ")
	if len(argv) == 0 {
		return nil, fmt.Errorf("template compiled to empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = data.Dir
	return cmd, nil
}

func Run(command string, data CommandData) error {
	if command == "" {
		return nil
	}
	cmd, err := compileCommand(command, data)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stderr: %q: %w", stderr.String(), err)
	}
	return nil
}
