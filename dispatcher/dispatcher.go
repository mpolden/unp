package dispatcher

import (
	"fmt"

	"path/filepath"

	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.InCloseWrite}

type Dispatcher struct {
	Config
	OnFile    func(e Event, path Path, message chan<- string)
	watcher   chan notify.EventInfo
	message   chan string
	fileEvent chan Event
}

func (d *Dispatcher) closeFile(e *Event) error {
	p, ok := d.FindPath(e.Path())
	if !ok {
		return fmt.Errorf("no configured path found: %s", e.Path())
	}
	if p.SkipHidden && (e.IsHidden() || e.IsParentHidden()) {
		return fmt.Errorf("hidden parent dir or file: %s", e.Path())
	}
	if !p.ValidDepth(e.Depth()) {
		return fmt.Errorf("incorrect depth: %s depth:%d min:%d max:%d",
			e.Path(), e.Depth(), p.MinDepth, p.MaxDepth)
	}
	if match, err := p.Match(e.Base()); !match {
		if err != nil {
			return err
		}
		return fmt.Errorf("no match found: %s", e.Path())
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
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, d.watcher, flags...); err != nil {
			d.message <- err.Error()
		}
		d.message <- fmt.Sprintf("Watching recursively: %s", path.Name)
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

func (d *Dispatcher) readEvent() {
	for {
		select {
		case ev := <-d.watcher:
			e := Event{ev}
			if e.IsCloseWrite() {
				d.fileEvent <- e
			}
		}
	}
}

func (d *Dispatcher) Serve() <-chan string {
	d.watch()
	go d.readFileEvent()
	go d.readEvent()
	return d.message
}

func New(cfg Config, bufferSize int) (*Dispatcher, error) {
	// Buffer events so that we don't miss any
	watcher := make(chan notify.EventInfo, bufferSize)
	message := make(chan string, bufferSize)
	fileEvent := make(chan Event, bufferSize)
	return &Dispatcher{
		Config:    cfg,
		watcher:   watcher,
		message:   message,
		fileEvent: fileEvent,
	}, nil
}
