package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func parseArgs(args []string) ([]Path, error) {
	paths := make([]Path, 0, len(args))
	for _, arg := range args {
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
	var opts struct {
		Remove   bool     `short:"r" long:"remove" description:"Remove files after extraction"`
		Patterns []string `short:"p" long:"patterns" description:"File patterns to process"`
	}

	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	if _, err := exec.LookPath("unrar"); err != nil {
		log.Fatal(err)
	}

	paths, err := parseArgs(args[1:])
	if err != nil {
		fmt.Println(err)
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
		Patterns: opts.Patterns,
		Remove:   opts.Remove,
	}
	w.OnFile = u.onFile
	w.Serve()
}
