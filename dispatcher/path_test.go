package dispatcher

import "testing"

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
