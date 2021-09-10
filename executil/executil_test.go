package executil

import (
	"os"
	"strings"
	"testing"
)

func TestCompileCommand(t *testing.T) {
	tmpl := "tar -xf {{.Name}} {{.Base}} {{.Dir}}"
	data := CommandData{
		Name: "/foo/bar/baz.rar",
		Base: "baz.rar",
		Dir:  "/foo/bar",
	}
	cmd, err := compileCommand(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != data.Dir {
		t.Fatalf("want %q, got %q", data.Dir, cmd.Dir)
	}
	if !strings.Contains(cmd.Path, string(os.PathSeparator)) {
		t.Fatalf("want %q to contain a path separator", cmd.Path)
	}
	if cmd.Args[0] != "tar" {
		t.Fatalf("want %q, got %q", "tar", cmd.Args[0])
	}
	if cmd.Args[1] != "-xf" {
		t.Fatalf("want %q, got %q", "-xf", cmd.Args[1])
	}
	if cmd.Args[2] != data.Name {
		t.Fatalf("want %q, got %q", data.Name, cmd.Args[2])
	}
	if cmd.Args[3] != data.Base {
		t.Fatalf("want %q, got %q", data.Base, cmd.Args[3])
	}
	if cmd.Args[4] != data.Dir {
		t.Fatalf("want %q, got %q", data.Base, cmd.Args[4])
	}
	if _, err := compileCommand("tar -xf {{.Bar}}", data); err == nil {
		t.Fatal("want error")
	}
}
