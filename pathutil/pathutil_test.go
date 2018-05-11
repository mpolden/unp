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

func TestContainsHidden(t *testing.T) {
	var tests = []struct {
		in  string
		out bool
	}{
		{"foo", false},
		{"/foo/bar", false},
		{".bar", true},
		{"/foo/.bar", true},
		{"/foo/.bar/baz", true},
		{"/.foo/bar/baz", true},
	}
	for _, tt := range tests {
		if got := ContainsHidden(tt.in); got != tt.out {
			t.Errorf("want %t, got %t for %s", tt.out, got, tt.in)
		}
	}
}
