package dispatcher

import (
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"log"

	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
)

type OnFile func(Event, Path) error

type dispatcher struct {
	config Config
	onFile OnFile
	events chan notify.EventInfo
	signal chan os.Signal
	log    *log.Logger
}

func (d *dispatcher) dispatch(e Event) error {
	p, ok := d.config.findPath(e.Name())
	if !ok {
		return errors.Errorf("no configured path found: %s", e.Name())
	}
	if p.SkipHidden && (e.IsHidden() || e.IsParentHidden()) {
		return errors.Errorf("hidden parent dir or file: %s", e.Name())
	}
	if !p.validDepth(e.Depth()) {
		return errors.Errorf("incorrect depth: %s depth=%d min=%d max=%d",
			e.Name(), e.Depth(), p.MinDepth, p.MaxDepth)
	}
	if match, err := p.match(e.Base()); !match {
		if err != nil {
			return err
		}
		return errors.Errorf("no match found: %s", e.Name())
	}
	return d.onFile(e, p)
}

func (d *dispatcher) watch() {
	for _, path := range d.config.Paths {
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, d.events, flags...); err != nil {
			d.log.Printf("Failed to watch %s: %s", recursivePath, err)
		} else {
			d.log.Printf("Watching %s recursively", path.Name)
		}
	}
}

func (d *dispatcher) reload() {
	for {
		s := <-d.signal
		d.log.Printf("Received %s, reloading configuration", s)
		cfg, err := ReadConfig(d.config.filename)
		if err == nil {
			notify.Stop(d.events)
			d.config = cfg
			d.watch()
		} else {
			d.log.Printf("Failed to read config: %s", err)
		}
	}
}

func (d *dispatcher) readEvents() {
	for ev := range d.events {
		e := Event{ev}
		if e.IsCloseWrite() {
			if err := d.dispatch(e); err != nil {
				d.log.Printf("Skipping event: %s", err)
			}
		}
	}
}

func (d *dispatcher) Serve() {
	d.watch()
	go d.reload()
	d.readEvents()
}

func New(cfg Config, onFile OnFile, log *log.Logger) *dispatcher {
	// Buffer events so that we don't miss any
	events := make(chan notify.EventInfo, cfg.BufferSize)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	return &dispatcher{
		config: cfg,
		events: events,
		log:    log,
		onFile: onFile,
		signal: sig,
	}
}
