package logutil

import (
	"bytes"
	"fmt"
	"io"
)

const (
	cursorUp = "\033[1A"
	eraseEOL = "\033[K"
)

type UniqueWriter struct {
	w     io.Writer
	prev  []byte
	count int64
}

func NewUniqueWriter(w io.Writer) *UniqueWriter { return &UniqueWriter{w: w} }

func (uw *UniqueWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if bytes.Equal(uw.prev, p) {
		uw.count++
		if p[len(p)-1] == '\n' {
			p = p[:len(p)-1]
		}
		space := ""
		if len(p) > 0 {
			space = " "
		}
		return fmt.Fprintf(uw.w, "%s%s%s%s[repeated %d times]\n", cursorUp, eraseEOL, p, space, uw.count)
	}
	uw.prev = p
	uw.count = 1
	return uw.w.Write(p)
}
