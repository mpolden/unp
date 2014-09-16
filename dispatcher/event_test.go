package dispatcher

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

func TestIsDir(t *testing.T) {
	e := Event{Mask: inotify.IN_ISDIR}
	if !e.IsDir() {
		t.Fatal("Expected true")
	}
	e = Event{Mask: 0}
	if e.IsDir() {
		t.Fatal("Expected false")
	}
}

func TestIsCreate(t *testing.T) {
	e := Event{Mask: inotify.IN_CREATE}
	if !e.IsCreate() {
		t.Fatal("Expected true")
	}
	e = Event{Mask: 0}
	if e.IsCreate() {
		t.Fatal("Expected false")
	}
}

func TestIsClose(t *testing.T) {
	e := Event{Mask: inotify.IN_CLOSE}
	if !e.IsClose() {
		t.Fatal("Expected true")
	}
	e = Event{Mask: 0}
	if e.IsClose() {
		t.Fatal("Expected false")
	}
}

func TestIsCloseWrite(t *testing.T) {
	e := Event{Mask: inotify.IN_CLOSE_WRITE}
	if !e.IsCloseWrite() {
		t.Fatal("Expected true")
	}
	e = Event{Mask: 0}
	if e.IsCloseWrite() {
		t.Fatal("Expected false")
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
