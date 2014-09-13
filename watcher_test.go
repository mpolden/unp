package main

import (
	"testing"
)

func TestFindPath(t *testing.T) {
	paths := []Path{
		Path{Name: "/foo/bar"},
	}
	w := Worker{Paths: paths}
	parent, ok := w.findPath("/foo/bar/baz/bax")
	if !ok {
		t.Fatal("Expected true")
	}
	if parent.Name != paths[0].Name {
		t.Fatalf("Expected %s, got %s", paths[0].Name, parent.Name)
	}

	parent, ok = w.findPath("/eggs/spam")
	if ok {
		t.Fatal("Expected false")
	}
	empty := Path{}
	if parent.Name != empty.Name {
		t.Fatalf("Expected nil, got %+v", parent)
	}
}
