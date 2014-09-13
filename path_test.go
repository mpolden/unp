package main

import (
	"os/exec"
	"testing"
)

func TestPathDepth(t *testing.T) {
	if d := PathDepth("/foo"); d != 1 {
		t.Fatalf("Expected 1, got %d", d)
	}
	if d := PathDepth("/foo/"); d != 1 {
		t.Fatalf("Expected 1, got %d", d)
	}
	if d := PathDepth("/foo/bar/baz"); d != 3 {
		t.Fatalf("Expected 3, got %d", d)
	}
	if d := PathDepth("/foo/bar/baz/"); d != 3 {
		t.Fatalf("Expected 3, got %d", d)
	}
}

func TestPathMatch(t *testing.T) {
	p := Path{
		Patterns: []string{"*.txt"},
	}
	match, err := p.Match("foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Fatal("Expected true")
	}
	match, err = p.Match("foo")
	if err != nil {
		t.Fatal(err)
	}
	if match {
		t.Fatal("Expected false")
	}
	p = Path{
		Patterns: []string{"[bad pattern"},
	}
	_, err = p.Match("foo")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestArchiveExtWithDot(t *testing.T) {
	p := Path{
		ArchiveExt: "rar",
	}
	if ext := p.ArchiveExtWithDot(); ext != ".rar" {
		t.Fatalf("Expected '.rar', got '%s'", ext)
	}
	p = Path{
		ArchiveExt: ".rar",
	}
	if ext := p.ArchiveExtWithDot(); ext != ".rar" {
		t.Fatalf("Expected '.rar', got '%s'", ext)
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
	if cmd.Args[0] != "-xf" {
		t.Fatalf("Expected '-xf', got '%s'", cmd.Args[0])
	}
	if cmd.Args[1] != values.Name {
		t.Fatalf("Expected '%s', got '%s'", values.Name, cmd.Args[1])
	}
	if cmd.Args[2] != values.Base {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[2])
	}
	if cmd.Args[3] != values.Dir {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[3])
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
	p := Path{
		MinDepth: 4,
		MaxDepth: 5,
	}
	if p.ValidDepth(3) {
		t.Fatal("Expected false")
	}
	if !p.ValidDepth(4) {
		t.Fatal("Expected true")
	}
	if !p.ValidDepth(5) {
		t.Fatal("Expected true")
	}
	if p.ValidDepth(6) {
		t.Fatal("Expected false")
	}
}

func TestValidDirDepth(t *testing.T) {
	p := Path{
		MaxDepth: 5,
	}
	if !p.ValidDirDepth(4) {
		t.Fatal("Expected true")
	}
	if p.ValidDirDepth(5) {
		t.Fatal("Expected false")
	}
	if p.ValidDirDepth(6) {
		t.Fatal("Expected false")
	}
}
