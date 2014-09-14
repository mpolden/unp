package main

import (
	"code.google.com/p/go.exp/inotify"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Worker struct {
	Config
	OnFile  func(e *Event, path *Path)
	watcher *inotify.Watcher
	watched map[string]bool
}

func (w *Worker) handleCreateDir(e *Event) error {
	if !e.IsDir() {
		return nil
	}
	if !e.IsCreate() && !e.IsCloseNoWrite() {
		return nil
	}
	if _, exists := w.watched[e.Name]; exists {
		return nil
	}
	p, ok := w.FindPath(e.Name)
	if !ok {
		return fmt.Errorf("no configured path found for %s", e.Name)
	}
	if p.SkipHidden && e.IsHidden() {
		return fmt.Errorf("hidden directory: %s", e.Name)
	}
	if !p.ValidDirDepth(e.Depth()) {
		return fmt.Errorf("invalid dir depth %s (max: %d)", e,
			p.MaxDepth-1)
	}
	log.Printf("Watching path: %s", e.Name)
	if err := w.watcher.Watch(e.Name); err != nil {
		return err
	}
	w.watched[e.Name] = true
	return nil
}

func (w *Worker) handleCloseFile(e *Event) error {
	if !e.IsClose() && !e.IsCloseWrite() {
		return nil
	}
	p, ok := w.FindPath(e.Name)
	if !ok {
		return fmt.Errorf("no configured path found for %s", e.Name)
	}
	if p.SkipHidden && e.IsHidden() {
		return fmt.Errorf("hidden file: %s", e.Name)
	}
	if !p.ValidDepth(e.Depth()) {
		return fmt.Errorf("invalid depth %s (min: %d, max: %d)", e,
			p.MinDepth, p.MaxDepth)
	}
	if match, err := p.Match(e.Base()); !match {
		if err != nil {
			return err
		}
		return fmt.Errorf("no match found for %s", e.Name)
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
		p, ok := w.FindPath(path)
		if !ok {
			return fmt.Errorf("%s is not configured", path)
		}
		if !p.ValidDirDepth(PathDepth(path)) {
			return nil
		}
		log.Printf("Watching path: %s", path)
		if err := w.watcher.Watch(path); err != nil {
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
		case ev := <-w.watcher.Event:
			e := Event(*ev)
			if err := w.handleCreateDir(&e); err != nil {
				log.Printf("Skipping event: %s", err)
			}
			if err := w.handleCloseFile(&e); err != nil {
				log.Printf("Skipping event: %s", err)
			}
		case err := <-w.watcher.Error:
			log.Print(err)
		}
	}
}

func New() (*Worker, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watched := make(map[string]bool)
	return &Worker{
		watcher: watcher,
		watched: watched,
	}, nil
}
