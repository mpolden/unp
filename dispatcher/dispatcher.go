package dispatcher

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.InCreate, notify.InCloseWrite}

type OnFile func(event Event, path Path, message chan<- string)

type Dispatcher struct {
	Config
	onFile  OnFile
	watcher chan notify.EventInfo
	message chan string
	signal  chan os.Signal
}

func (d *Dispatcher) createDir(e Event) error {
	return filepath.Walk(e.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		d.message <- fmt.Sprintf("New directory: %s", path)
		return nil
	})
}

func (d *Dispatcher) processFile(e Event) error {
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
	if d.Async {
		go d.onFile(e, p, d.message)
	} else {
		d.onFile(e, p, d.message)
	}
	return nil
}

func (d *Dispatcher) watch() {
	for _, path := range d.Paths {
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, d.watcher, flags...); err != nil {
			d.message <- err.Error()
		} else {
			d.message <- fmt.Sprintf("Watching recursively: %s", path.Name)
		}
	}
}

func (d *Dispatcher) rewatch() {
	for {
		s := <-d.signal
		d.message <- fmt.Sprintf("Received %s, rewatching directories", s)
		d.watch()
	}
}

func (d *Dispatcher) readEvents() {
	for {
		select {
		case ev := <-d.watcher:
			e := Event{ev}
			if e.IsCreate() && e.IsDir() {
				if err := d.createDir(e); err != nil {
					d.message <- fmt.Sprintf("Skipping event: %s", err)
				}
			} else if e.IsCloseWrite() {
				if err := d.processFile(e); err != nil {
					d.message <- fmt.Sprintf("Skipping event: %s", err)
				}
			}
		}
	}
}

func (d *Dispatcher) Serve() <-chan string {
	d.watch()
	go d.rewatch()
	go d.readEvents()
	return d.message
}

func New(cfg Config, bufferSize int, handler OnFile) *Dispatcher {
	// Buffer events so that we don't miss any
	watcher := make(chan notify.EventInfo, bufferSize)
	message := make(chan string, bufferSize)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	return &Dispatcher{
		Config:  cfg,
		watcher: watcher,
		message: message,
		onFile:  handler,
		signal:  sig,
	}
}
