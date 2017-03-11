package dispatcher

import "testing"

func TestPathMatch(t *testing.T) {
	var tests = []struct {
		p      Path
		in     string
		out    bool
		nilErr bool
	}{
		{Path{Patterns: []string{"*.txt"}}, "foo.txt", true, true},
		{Path{Patterns: []string{"*.txt"}}, "foo", false, true},
		{Path{Patterns: []string{"[bad pattern"}}, "foo", false, false},
	}

	for _, tt := range tests {
		rv, err := tt.p.match(tt.in)
		nilErr := err == nil
		if nilErr != tt.nilErr {
			t.Fatalf("Expected %t, got %t", tt.nilErr, nilErr)
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
