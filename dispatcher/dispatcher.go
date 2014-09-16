package dispatcher

import (
	"code.google.com/p/go.exp/inotify"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Dispatcher struct {
	Config
	OnFile  func(e *Event, path *Path)
	watcher *inotify.Watcher
}

func (d *Dispatcher) watchDir(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
	p, ok := d.FindPath(path)
	if !ok {
		return fmt.Errorf("not configured: %s", path)
	}
	if p.SkipHidden && IsHidden(path) {
		return fmt.Errorf("hidden directory: %s", path)
	}
	if depth := PathDepth(path); !p.ValidDirDepth(depth) {
		return fmt.Errorf("incorrect depth: %s depth:%d max:%d",
			path, depth, p.MaxDepth-1)
	}
	log.Printf("Watching path: %s", path)
	if err := d.watcher.Watch(path); err != nil {
		return err
	}
	return nil
}

func (d *Dispatcher) handleCreateDir(e *Event) error {
	if !e.IsCreate() || !e.IsDir() {
		return nil
	}
	if err := filepath.Walk(e.Name, d.watchDir); err != nil {
		return err
	}
	return nil
}

func (d *Dispatcher) handleCloseFile(e *Event) error {
	if !e.IsClose() && !e.IsCloseWrite() {
		return nil
	}
	p, ok := d.FindPath(e.Name)
	if !ok {
		return fmt.Errorf("no configured path found: %s", e.Name)
	}
	if p.SkipHidden && e.IsHidden() {
		return fmt.Errorf("hidden file: %s", e.Name)
	}
	if !p.ValidDepth(e.Depth()) {
		return fmt.Errorf("incorrect depth: %s depth:%d min:%d max:%d",
			e.Name, e.Depth(), p.MinDepth, p.MaxDepth)
	}
	if match, err := p.Match(e.Base()); !match {
		if err != nil {
			return err
		}
		return fmt.Errorf("no match found: %s", e.Name)
	}
	if d.OnFile != nil {
		d.OnFile(e, &p)
	}
	return nil
}

func (d *Dispatcher) Watch() error {
	for _, path := range d.Paths {
		if err := filepath.Walk(path.Name, d.watchDir); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Serve() {
	for {
		select {
		case ev := <-d.watcher.Event:
			e := Event(*ev)
			if err := d.handleCreateDir(&e); err != nil {
				log.Printf("Skipping event: %s", err)
			}
			if err := d.handleCloseFile(&e); err != nil {
				log.Printf("Skipping event: %s", err)
			}
		case err := <-d.watcher.Error:
			log.Print(err)
		}
	}
}

func New(cfg Config) (*Dispatcher, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Dispatcher{
		Config:  cfg,
		watcher: watcher,
	}, nil
}
