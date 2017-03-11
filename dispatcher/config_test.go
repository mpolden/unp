package dispatcher

import (
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	jsonConfig := `
{
  "Default": {
    "MaxDepth": 3
  },
  "Paths": [
    {
      "Name": "/tmp/foo"
    },
    {
      "Name": "/tmp/bar",
      "MaxDepth": 4
    }
  ]
}
`
	cfg, err := readConfig(strings.NewReader(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	// Test that defaults are applied and can be overridden
	var tests = []struct {
		i   int
		out int
	}{
		{0, 3},
		{1, 4},
	}
	for _, tt := range tests {
		path := cfg.Paths[tt.i]
		if got := path.MaxDepth; got != tt.out {
			t.Errorf("Expected MaxDepth=%d, got Parser=%d for Name=%q", tt.out, got, path.Name)
		}
	}
}

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
		rv, ok := c.findPath(tt.in)
		if ok != tt.ok {
			t.Errorf("Expected %t, got %t", tt.ok, ok)
		}
		if rv.Name != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, rv.Name)
		}
	}
}
