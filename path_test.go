package main

import (
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
