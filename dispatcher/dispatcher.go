package dispatcher

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/exp/inotify"
)

const flags = inotify.IN_CREATE | inotify.IN_CLOSE | inotify.IN_CLOSE_WRITE

type Dispatcher struct {
	Config
	OnFile    func(e Event, path Path, message chan<- string)
	watcher   *inotify.Watcher
	message   chan string
	dirEvent  chan Event
	fileEvent chan Event
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
	d.message <- fmt.Sprintf("Watching path: %s", path)
	if err := d.watcher.AddWatch(path, flags); err != nil {
		return err
	}
	return nil
}

func (d *Dispatcher) createDir(name string) error {
	if err := filepath.Walk(name, d.watchDir); err != nil {
		return err
	}
	return nil
}

func (d *Dispatcher) closeFile(e *Event) error {
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
			go d.OnFile(*e, p, d.message)
		} else {
			d.OnFile(*e, p, d.message)
		}
	}
	return nil
}

func (d *Dispatcher) watch() {
	for _, path := range d.Paths {
		if err := d.createDir(path.Name); err != nil {
			d.message <- err.Error()
		}
	}
}

func (d *Dispatcher) readDirEvent() {
	for {
		select {
		case e := <-d.dirEvent:
			if err := d.createDir(e.Name); err != nil {
				d.message <- fmt.Sprintf("Skipping event: %s", err)
			}
		}
	}
}

func (d *Dispatcher) readFileEvent() {
	for {
		select {
		case e := <-d.fileEvent:
			if err := d.closeFile(&e); err != nil {
				d.message <- fmt.Sprintf("Skipping event: %s", err)
			}
		}
	}
}

func (d *Dispatcher) Serve() <-chan string {
	go d.watch()
	go d.readDirEvent()
	go d.readFileEvent()
	go func() {
		for {
			select {
			case ev := <-d.watcher.Event:
				e := Event(*ev)
				if e.IsCreate() && e.IsDir() {
					d.dirEvent <- e
				} else if e.IsClose() || e.IsCloseWrite() {
					d.fileEvent <- e
				}
			case err := <-d.watcher.Error:
				d.message <- err.Error()
			}
		}
	}()
	return d.message
}

func New(cfg Config, bufferSize int) (*Dispatcher, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	// Buffer events so that we don't miss any
	message := make(chan string, bufferSize)
	dirEvent := make(chan Event, bufferSize)
	fileEvent := make(chan Event, bufferSize)
	return &Dispatcher{
		Config:    cfg,
		watcher:   watcher,
		message:   message,
		dirEvent:  dirEvent,
		fileEvent: fileEvent,
	}, nil
}
