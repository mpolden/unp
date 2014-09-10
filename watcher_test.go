package main

import (
	"code.google.com/p/go.exp/inotify"
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

func TestEventDir(t *testing.T) {
	e := Event{
		Mask: inotify.IN_ISDIR,
		Name: "/foo/bar",
	}
	if dir := e.Dir(); dir != "/foo/bar" {
		t.Fatal("Expected '/foo/bar', got '%s'", dir)
	}
	e = Event{
		Mask: inotify.IN_CLOSE,
		Name: "/foo/bar",
	}
	if dir := e.Dir(); dir != "/foo" {
		t.Fatal("Expected '/foo', got '%s'", dir)
	}
}

func TestEventBase(t *testing.T) {
	e := Event{
		Mask: inotify.IN_CLOSE,
		Name: "/foo/bar",
	}
	if base := e.Base(); base != "bar" {
		t.Fatal("Expected 'bar', got '%s'", base)
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
}

func TestParentPath(t *testing.T) {
	paths := []Path{
		Path{Name: "/foo/bar"},
	}
	w := Worker{Paths: paths}
	parent, ok := w.parentPath("/foo/bar/baz/bax")
	if !ok {
		t.Fatal("Expected true")
	}
	if parent.Name != paths[0].Name {
		t.Fatalf("Expected %s, got %s", paths[0].Name, parent.Name)
	}

	parent, ok = w.parentPath("/eggs/spam")
	if ok {
		t.Fatal("Expected false")
	}
	empty := Path{}
	if parent.Name != empty.Name {
		t.Fatalf("Expected nil, got %+v", parent)
	}
}
