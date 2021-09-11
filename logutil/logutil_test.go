package logutil

import (
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	var sb strings.Builder
	w := NewUniqueWriter(&sb)

	w.Write([]byte("l1\n"))
	w.Write([]byte("l1\n"))
	w.Write([]byte("l2\n"))
	w.Write([]byte("\n"))
	w.Write([]byte("\n"))

	want := "l1\n" +
		"\033[1A\033[Kl1 [repeated 2 times]\n" +
		"l2\n" +
		"\n\033[1A\033[K[repeated 2 times]\n"
	if got := sb.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
