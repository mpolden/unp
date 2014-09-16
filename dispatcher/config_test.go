package dispatcher

import (
	"testing"
)

func TestFindPath(t *testing.T) {
	c := Config{
		Paths: []Path{Path{Name: "/foo/bar"}},
	}
	parent, ok := c.FindPath("/foo/bar/baz/bax")
	if !ok {
		t.Fatal("Expected true")
	}
	if parent.Name != c.Paths[0].Name {
		t.Fatalf("Expected %s, got %s", c.Paths[0].Name, parent.Name)
	}

	parent, ok = c.FindPath("/eggs/spam")
	if ok {
		t.Fatal("Expected false")
	}
	empty := Path{}
	if parent.Name != empty.Name {
		t.Fatalf("Expected nil, got %+v", parent)
	}
}
