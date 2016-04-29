package dispatcher

import (
	"testing"

	"github.com/rjeczalik/notify"
)

func (e *mockEvent) Path() string { return e.name }

func (e *mockEvent) Event() notify.Event { return e.mask }

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
		{newEvent("/foo", IsDir), true},
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
		{newEvent("/foo", notify.Create), true},
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
		{newEvent("/foo", IsCloseWrite), true},
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
		{newEvent("/foo/bar", IsDir), "/foo/bar"},
		{newEvent("/foo/bar", IsCloseWrite), "/foo"},
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
		{newEvent("/foo/.bar", IsDir), true},
		{newEvent("/foo/.bar", 0), false},
	}
	for _, tt := range tests {
		if rv := tt.in.IsParentHidden(); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}
