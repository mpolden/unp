package dispatcher

import (
	"testing"

	"github.com/rjeczalik/notify"
)

func (e *mockEvent) Path() string { return e.name }

func (e *mockEvent) Event() notify.Event { return e.mask }

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
