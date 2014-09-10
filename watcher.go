package main

import (
	"bytes"
	"code.google.com/p/go.exp/inotify"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Worker struct {
	Watcher *inotify.Watcher
	Paths   []Path
	OnFile  func(e *Event, path *Path)
}

type Path struct {
	Name          string
	MaxDepth      int
	Patterns      []string
	Remove        bool
	UnpackCommand string
	Command       string
}

type templateValues struct {
	Path string
	Dir  string
	File string
}

func (p *Path) createCommand(v templateValues, tmpl string) (*exec.Cmd, error) {
	t := template.Must(template.New("command").Parse(tmpl))
	var b bytes.Buffer
	if err := t.Execute(&b, v); err != nil {
		return nil, err
	}
	argv := strings.Split(b.String(), " ")
	if len(argv) == 0 {
		return nil, fmt.Errorf("template compiled to empty command")
	}
	if len(argv) == 1 {
		return exec.Command(argv[0]), nil
	}
	return exec.Command(argv[0], argv[1:]...), nil
}

func (p *Path) unpackCommand(v templateValues) (*exec.Cmd, error) {
	return p.createCommand(v, p.UnpackCommand)
}

func (p *Path) command(v templateValues) (*exec.Cmd, error) {
	return p.createCommand(v, p.Command)
}

func PathDepth(name string) int {
	name = filepath.Clean(name)
	return strings.Count(name, string(os.PathSeparator))
}

type Event inotify.Event

func (e *Event) Depth() int {
	return PathDepth(e.Name)
}

func (e *Event) String() string {
	return fmt.Sprintf("%s (depth: %d)", e.Name, e.Depth())
}

func (e *Event) IsDir() bool {
	return e.Mask&inotify.IN_ISDIR == inotify.IN_ISDIR
}

func (e *Event) IsCreate() bool {
	return e.Mask&inotify.IN_CREATE == inotify.IN_CREATE
}

func (e *Event) IsClose() bool {
	return e.Mask&inotify.IN_CLOSE == inotify.IN_CLOSE
}

func (e *Event) IsCloseWrite() bool {
	return e.Mask&inotify.IN_CLOSE_WRITE == inotify.IN_CLOSE_WRITE
}

func (e *Event) Dir() string {
	if e.IsDir() {
		return e.Name
	}
	return filepath.Dir(e.Name)
}

func (e *Event) Base() string {
	return filepath.Base(e.Name)
}

func (p *Path) Match(name string) (bool, error) {
	for _, pattern := range p.Patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (w *Worker) parentPath(name string) (Path, bool) {
	for _, p := range w.Paths {
		if strings.HasPrefix(name, p.Name) {
			return p, true
		}
	}
	return Path{}, false
}

func (w *Worker) handleCreateDir(e *Event) error {
	if !e.IsCreate() || !e.IsDir() {
		return nil
	}
	p, ok := w.parentPath(e.Name)
	if ok && e.Depth() <= p.MaxDepth {
		log.Printf("Watching path: %s", e)
		err := w.Watcher.AddWatch(e.Name, inotify.IN_ALL_EVENTS)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) handleCloseFile(e *Event) error {
	if !e.IsClose() && !e.IsCloseWrite() {
		return nil
	}
	p, ok := w.parentPath(e.Name)
	if !ok {
		return nil
	}
	if e.Depth() < p.MaxDepth {
		log.Printf("Not processing files at this level: %s", e)
		return nil
	}
	if match, err := p.Match(e.Base()); !match {
		if err != nil {
			log.Printf("Invalid pattern: %s", err)
		}
		log.Printf("Skipping %s", e)
		return nil
	}
	if w.OnFile != nil {
		w.OnFile(e, &p)
	}
	return nil
}

func (w *Worker) AddWatch(path Path) error {
	w.Paths = append(w.Paths, path)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		p, ok := w.parentPath(path)
		if !ok {
			return fmt.Errorf("%s is not configured", path)
		}
		if PathDepth(path) > p.MaxDepth {
			return nil
		}
		log.Printf("Watching path: %s", path)
		err = w.Watcher.AddWatch(path, inotify.IN_ALL_EVENTS)
		if err != nil {
			return err
		}
		return nil
	}
	if err := filepath.Walk(path.Name, walkFn); err != nil {
		return err
	}
	return nil
}

func (w *Worker) Serve() {
	for {
		select {
		case ev := <-w.Watcher.Event:
			e := Event(*ev)
			if err := w.handleCreateDir(&e); err != nil {
				log.Printf("failed to process event: %s", err)
			}
			if err := w.handleCloseFile(&e); err != nil {
				log.Printf("failed to process event: %s", err)
			}
		case err := <-w.Watcher.Error:
			log.Printf("error: %s", err)
		}
	}
}

func New() (*Worker, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Worker{
		Watcher: watcher,
		Paths:   []Path{},
	}, nil
}
