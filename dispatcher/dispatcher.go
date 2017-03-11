package dispatcher

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"log"

	"github.com/rjeczalik/notify"
)

type OnFile func(Event, Path) error

type Dispatcher struct {
	config  Config
	onFile  OnFile
	watcher chan notify.EventInfo
	signal  chan os.Signal
	log     *log.Logger
}

func (d *Dispatcher) createDir(e Event) error {
	return filepath.Walk(e.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		return nil
	})
}

func (d *Dispatcher) processFile(e Event) error {
	p, ok := d.config.FindPath(e.Path())
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
	return d.onFile(e, p)
}

func (d *Dispatcher) watch() {
	for _, path := range d.config.Paths {
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, d.watcher, flags...); err != nil {
			d.log.Printf("Failed to watch %s: %s", recursivePath, err)
		} else {
			d.log.Printf("Watching %s recursively", path.Name)
		}
	}
}

func (d *Dispatcher) reload() {
	for {
		s := <-d.signal
		d.log.Printf("Received %s, reloading configuration", s)
		cfg, err := ReadConfig(d.config.filename)
		if err == nil {
			d.log.Print("Removing all watches")
			notify.Stop(d.watcher)
			d.config = cfg
			d.watch()
		} else {
			d.log.Printf("Failed to read config: %s", err)
		}
	}
}

func (d *Dispatcher) readEvents() {
	for ev := range d.watcher {
		e := Event{ev}
		if e.IsCreate() && e.IsDir() {
			if err := d.createDir(e); err != nil {
				d.log.Printf("Skipping event: %s", err)
			}
		} else if e.IsCloseWrite() {
			if err := d.processFile(e); err != nil {
				d.log.Printf("Skipping event: %s", err)
			}
		}
	}
}

func (d *Dispatcher) Serve() {
	d.watch()
	go d.reload()
	d.readEvents()
}

func New(cfg Config, handler OnFile, log *log.Logger) *Dispatcher {
	// Buffer events so that we don't miss any
	watcher := make(chan notify.EventInfo, cfg.BufferSize)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	return &Dispatcher{
		config:  cfg,
		watcher: watcher,
		log:     log,
		onFile:  handler,
		signal:  sig,
	}
}
