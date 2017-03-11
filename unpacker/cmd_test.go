package unpacker

import (
	"os"
	"strings"
	"testing"
)

func TestNewCmd(t *testing.T) {
	tmpl := "tar -xf {{.Name}} {{.Base}} {{.Dir}}"
	values := cmdValues{
		Name: "/foo/bar/baz.rar",
		Base: "baz.rar",
		Dir:  "/foo/bar",
	}

	cmd, err := newCmd(tmpl, values)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != values.Dir {
		t.Fatalf("Expected %s, got %s", values.Dir, cmd.Dir)
	}
	if !strings.Contains(cmd.Path, string(os.PathSeparator)) {
		t.Fatalf("Expected %s to contain a path separator", cmd.Path)
	}
	if cmd.Args[0] != "tar" {
		t.Fatalf("Expected 'tar', got '%s'", cmd.Args[0])
	}
	if cmd.Args[1] != "-xf" {
		t.Fatalf("Expected '-xf', got '%s'", cmd.Args[1])
	}
	if cmd.Args[2] != values.Name {
		t.Fatalf("Expected '%s', got '%s'", values.Name, cmd.Args[2])
	}
	if cmd.Args[3] != values.Base {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[3])
	}
	if cmd.Args[4] != values.Dir {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[4])
	}

	if _, err := newCmd("tar -xf {{.Bar}}", values); err == nil {
		t.Fatal("Expected error")
	}
}
