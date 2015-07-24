package dispatcher

import (
	"os/exec"
	"testing"
)

func TestPathMatch(t *testing.T) {
	var tests = []struct {
		p      Path
		in     string
		out    bool
		nilErr bool
	}{
		{Path{Patterns: []string{"*.txt"}}, "foo.txt", true, true},
		{Path{Patterns: []string{"*.txt"}}, "foo", false, true},
		{Path{Patterns: []string{"[bad pattern"}}, "foo", false, false},
	}

	for _, tt := range tests {
		rv, err := tt.p.Match(tt.in)
		nilErr := err == nil
		if nilErr != tt.nilErr {
			t.Fatalf("Expected %t, got %t", tt.nilErr, nilErr)
		}
		if rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestArchiveExtWithDot(t *testing.T) {
	var tests = []struct {
		in  Path
		out string
	}{
		{Path{ArchiveExt: "rar"}, ".rar"},
		{Path{ArchiveExt: ".rar"}, ".rar"},
	}
	for _, tt := range tests {
		if rv := tt.in.ArchiveExtWithDot(); rv != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, rv)
		}
	}
}

func TestParseUnpackCommand(t *testing.T) {
	cmdPath, err := exec.LookPath("tar")
	if err != nil {
		t.Fatal(err)
	}

	p := Path{
		UnpackCommand: "tar -xf {{.Name}} {{.Base}} {{.Dir}}",
	}
	values := CommandValues{
		Name: "/foo/bar/baz.rar",
		Base: "baz.rar",
		Dir:  "/foo/bar",
	}

	cmd, err := p.NewUnpackCommand(values)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != values.Dir {
		t.Fatalf("Expected %s, got %s", values.Dir, cmd.Dir)
	}
	if cmd.Path != cmdPath {
		t.Fatalf("Expected %s, got %s", cmdPath, cmd.Path)
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

	p = Path{
		UnpackCommand: "tar -xf {{.Bar}}",
	}
	_, err = p.NewUnpackCommand(values)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestValidDepth(t *testing.T) {
	var tests = []struct {
		in  int
		out bool
	}{
		{3, false},
		{4, true},
		{5, true},
		{6, false},
	}
	p := Path{MinDepth: 4, MaxDepth: 5}
	for _, tt := range tests {
		if rv := p.ValidDepth(tt.in); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}
