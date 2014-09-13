package main

import (
	"code.google.com/p/go.exp/inotify"
	"testing"
)

func TestDepth(t *testing.T) {
	e := Event{Name: "/foo"}
	if d := e.Depth(); d != 1 {
		t.Fatalf("Expected 1, got %d", d)
	}
	e = Event{Name: "/foo/"}
	if d := e.Depth(); d != 1 {
		t.Fatalf("Expected 1, got %d", d)
	}
	e = Event{Name: "/foo/bar/baz"}
	if d := e.Depth(); d != 3 {
		t.Fatalf("Expected 3, got %d", d)
	}
	e = Event{Name: "/foo/bar/baz/"}
	if d := e.Depth(); d != 3 {
		t.Fatalf("Expected 3, got %d", d)
	}
}

func TestDir(t *testing.T) {
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

func TestBase(t *testing.T) {
	e := Event{
		Mask: inotify.IN_CLOSE,
		Name: "/foo/bar",
	}
	if base := e.Base(); base != "bar" {
		t.Fatal("Expected 'bar', got '%s'", base)
	}
}

func TestIsHidden(t *testing.T) {
	e := Event{
		Name: "/foo/.bar",
	}
	if !e.IsHidden() {
		t.Fatal("Expected true")
	}
	e = Event{
		Name: "/foo/bar",
	}
	if e.IsHidden() {
		t.Fatal("Expected false")
	}
}
