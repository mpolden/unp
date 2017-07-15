package pathutil

import "testing"

func TestDepth(t *testing.T) {
	var tests = []struct {
		in  string
		out int
	}{
		{"/foo", 1},
		{"/foo/", 1},
		{"/foo/bar/baz", 3},
		{"/foo/bar/baz/", 3},
	}
	for _, tt := range tests {
		if got := Depth(tt.in); got != tt.out {
			t.Errorf("want %d, got %d for %s", tt.out, got, tt.in)
		}
	}
}

func TestIsHidden(t *testing.T) {
	var tests = []struct {
		in  string
		out bool
	}{
		{"/foo/.bar", true},
		{"/foo/bar", false},
		{"/foo/.bar/baz", false},
	}
	for _, tt := range tests {
		if got := IsHidden(tt.in); got != tt.out {
			t.Errorf("want %t, got %t for %s", tt.out, got, tt.in)
		}
	}
}

func TestIsParentHidden(t *testing.T) {
	var tests = []struct {
		in  string
		out bool
	}{
		{".bar", false},
		{"/foo/.bar", false},
		{"/foo/.bar/baz", true},
		{"/foo/.bar/baz/foo", true},
		{"/foo/.bar/baz/foo/bar", true},
	}
	for _, tt := range tests {
		if got := IsParentHidden(tt.in); got != tt.out {
			t.Errorf("want %t, got %t for %s", tt.out, got, tt.in)
		}
	}
}
