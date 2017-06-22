package dispatcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func mustTempDir() string {
	name, err := ioutil.TempDir("", "dispatcher")
	if err != nil {
		panic(err)
	}
	return name
}

func TestReadConfig(t *testing.T) {
	path1 := mustTempDir()
	defer os.RemoveAll(path1)
	path2 := mustTempDir()
	defer os.RemoveAll(path2)
	jsonConfig := fmt.Sprintf(`
{
  "Default": {
    "MaxDepth": 3
  },
  "Paths": [
    {
      "Name": "%s"
    },
    {
      "Name": "%s",
      "MaxDepth": 4
    }
  ]
}
`, path1, path2)
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

func TestPathMatch(t *testing.T) {
	var tests = []struct {
		p   Path
		in  string
		out bool
		err string
	}{
		{Path{Patterns: []string{"*.txt"}}, "foo.txt", true, ""},
		{Path{Patterns: []string{"*.txt"}}, "foo", false, ""},
		{Path{Patterns: []string{"[bad pattern"}}, "foo", false, "[bad pattern: syntax error in pattern"},
	}

	for _, tt := range tests {
		rv, err := tt.p.match(tt.in)
		if err != nil && err.Error() != tt.err {
			t.Fatalf("Expected error %q, got %q", tt.err, err.Error())
		}
		if rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}

func TestValidDepth(t *testing.T) {
	var tests = []struct {
		in  int
		out bool
	}{
		{3, false},
		{4, true},
		{5, true},
		{6, false},
	}
	p := Path{MinDepth: 4, MaxDepth: 5}
	for _, tt := range tests {
		if rv := p.validDepth(tt.in); rv != tt.out {
			t.Errorf("Expected %t, got %t", tt.out, rv)
		}
	}
}
