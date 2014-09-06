package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s path:level [path:level ...]\n",
		os.Args[0])
	os.Exit(1)
}

func parsePaths() ([]Path, error) {
	paths := make([]Path, 0, len(os.Args[1:]))
	for _, arg := range os.Args[1:] {
		s := strings.Split(arg, ":")
		if len(s) != 2 {
			return nil, fmt.Errorf("expected path:depth, got %s",
				arg)
		}
		path := filepath.Clean(s[0])
		depth, err := strconv.Atoi(s[1])
		if depth < 1 || err != nil {
			return nil, fmt.Errorf(
				"depth must be an positive integer")
		}
		p := Path{
			Name:     path,
			MaxDepth: depth,
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	if _, err := exec.LookPath("unrar"); err != nil {
		log.Fatal(err)
	}

	paths, err := parsePaths()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w, err := New()
	if err != nil {
		log.Fatal(err)
	}
	if err := w.AddWatch(paths); err != nil {
		log.Fatal(err)
	}
	u := Unpack{
		Patterns: []string{"*.r??", "*.sfv"},
	}
	w.OnFile = u.onFile
	w.Serve()
}
