package dispatcher

import (
	"testing"

	"syscall"

	"github.com/rjeczalik/notify"
)

type mockEvent struct {
	name         string
	mask         notify.Event
	inotifyEvent *syscall.InotifyEvent
}

// Satisfy notify.EventInfo interface
func (e *mockEvent) Path() string { return e.name }

func (e *mockEvent) Event() notify.Event { return e.mask }

func (e *mockEvent) Sys() interface{} { return e.inotifyEvent }

func newEvent(name string, mask notify.Event) Event {
	e := mockEvent{
		name:         name,
		mask:         mask,
		inotifyEvent: &syscall.InotifyEvent{Mask: uint32(mask)},
	}
	return Event{&e}
}

func TestDepth(t *testing.T) {
	var tests = []struct {
		in  Event
		out int
	}{
		{newEvent("/foo", 0), 1},
		{newEvent("/foo/", 0), 1},
		{newEvent("/foo/bar/baz", 0), 3},
		{newEvent("/foo/bar/baz/", 0), 3},
	}
	for _, tt := range tests {
		if depth := tt.in.Depth(); depth != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, depth)
		}
	}
}

func TestIsDir(t *testing.T) {
	var tests = []struct {
		in  Event
		out bool
	}{
		{newEvent("/foo", syscall.IN_ISDIR), true},
		{newEvent("/foo", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsDir(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestIsCreate(t *testing.T) {
	var tests = []struct {
		in  Event
		out bool
	}{
		{newEvent("/foo", notify.InCreate), true},
		{newEvent("/foo", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsCreate(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestIsCloseWrite(t *testing.T) {
	var tests = []struct {
		in  Event
		out bool
	}{
		{newEvent("/foo", notify.InCloseWrite), true},
		{newEvent("/foo", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsCloseWrite(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestDir(t *testing.T) {
	var tests = []struct {
		in  Event
		out string
	}{
		{newEvent("/foo/bar", syscall.IN_ISDIR), "/foo/bar"},
		{newEvent("/foo/bar", syscall.IN_CLOSE), "/foo"},
	}
	for _, tt := range tests {
		if rv := tt.in.Dir(); rv != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, rv)
		}
	}
}

func TestBase(t *testing.T) {
	e := newEvent("/foo/bar", 0)
	expected := "bar"
	if base := e.Base(); base != expected {
		t.Errorf("Expected %q, got %q", expected, base)
	}
}

func TestIsHidden(t *testing.T) {
	var tests = []struct {
		in  Event
		out bool
	}{
		{newEvent("/foo/.bar", 0), true},
		{newEvent("/foo/bar", 0), false},
		{newEvent("/foo/.bar/baz", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsHidden(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestIsParentHidden(t *testing.T) {
	var tests = []struct {
		in  Event
		out bool
	}{
		{newEvent("/foo/.bar/baz", 0), true},
		{newEvent("/foo/.bar/baz/foo", 0), true},
		{newEvent("/foo/.bar/baz/foo/bar", 0), true},
		{newEvent("/foo/.bar", syscall.IN_ISDIR), true},
		{newEvent("/foo/.bar", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsParentHidden(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}
