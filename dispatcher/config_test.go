package dispatcher

import (
	"testing"
)

func TestFindPath(t *testing.T) {
	c := Config{Paths: []Path{Path{Name: "/foo/bar"}}}
	var tests = []struct {
		in  string
		out string
		ok  bool
	}{
		{"/foo/bar/baz", "/foo/bar", true},
		{"/foo/bar/baz/bax", "/foo/bar", true},
		{"/foo", "", false},
		{"/eggs/spam", "", false},
	}
	for _, tt := range tests {
		rv, ok := c.FindPath(tt.in)
		if ok != tt.ok {
			t.Errorf("Expected %t, got %t", tt.ok, ok)
		}
		if rv.Name != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, rv)
		}
	}
}
