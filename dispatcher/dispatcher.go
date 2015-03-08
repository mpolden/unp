package dispatcher

import (
	"fmt"
	"golang.org/x/exp/inotify"
	"log"
	"os"
	"path/filepath"
)

const flags = inotify.IN_CREATE | inotify.IN_CLOSE | inotify.IN_CLOSE_WRITE

type Dispatcher struct {
	Config
	OnFile   func(e Event, path Path, messages chan<- string)
	watcher  *inotify.Watcher
	messages chan string
	events   chan Event
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
	if err := d.watcher.AddWatch(path, flags); err != nil {
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
		if d.Async {
			go d.OnFile(*e, p, d.messages)
		} else {
			d.OnFile(*e, p, d.messages)
		}
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
	go func() {
		for {
			select {
			case msg := <-d.messages:
				log.Print(msg)
			}
		}
	}()
	go func() {
		for {
			select {
			case e := <-d.events:
				if err := d.handleCreateDir(&e); err != nil {
					d.messages <- fmt.Sprintf("Skipping event: %s",
						err)
				}
				if err := d.handleCloseFile(&e); err != nil {
					d.messages <- fmt.Sprintf("Skipping event: %s",
						err)
				}
			}
		}
	}()
	for {
		select {
		case ev := <-d.watcher.Event:
			d.events <- Event(*ev)
		case err := <-d.watcher.Error:
			log.Print(err)
		}
	}
}

func New(cfg Config, bufferSize int) (*Dispatcher, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	messages := make(chan string)
	// Buffer events so that we don't miss any
	events := make(chan Event, bufferSize)
	return &Dispatcher{
		Config:   cfg,
		watcher:  watcher,
		messages: messages,
		events:   events,
	}, nil
}
