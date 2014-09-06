package main

import (
	"fmt"
	"github.com/martinp/gosfv"
	"io/ioutil"
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

func findSFV(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sfv" {
			return filepath.Join(dir, f.Name()), nil
		}
	}
	return "", nil
}

func unpack(e *Event, p *Path) {
	log.Printf("event %s", e)
	log.Printf("path %s", p)
	log.Printf("dir %s", filepath.Dir(e.Name))
	dir := filepath.Dir(e.Name)
	sfvPath, err := findSFV(dir)
	if err != nil {
		log.Print(err)
		return
	}
	if sfvPath == "" {
		log.Print("No SFV found")
	} else {
		log.Printf("sfv: %s", sfvPath)
	}
	sfvFile, err := sfv.Read(sfvPath)
	if err != nil {
		log.Print(err)
		return
	}
	if sfvFile.IsExist() {
		log.Printf("all files in sfv exist!")
		log.Printf("verifying sfv")
		ok, err := sfvFile.Verify()
		if err != nil {
			log.Print(err)
			return
		}
		if ok {
			log.Printf("unpack!")
		}
	}
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
	w.OnFile = unpack
	w.Serve()

}
