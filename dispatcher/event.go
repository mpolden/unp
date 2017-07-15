package dispatcher

import (
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"log"

	"github.com/mpolden/unpacker/pathutil"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
)

type OnFile func(string, Path) error

type dispatcher struct {
	config Config
	onFile OnFile
	events chan notify.EventInfo
	signal chan os.Signal
	log    *log.Logger
}

func (d *dispatcher) dispatch(name string) error {
	p, ok := d.config.findPath(name)
	if !ok {
		return errors.Errorf("no configured path found: %s", name)
	}
	if p.SkipHidden && (pathutil.IsHidden(name) || pathutil.IsParentHidden(name)) {
		return errors.Errorf("hidden parent dir or file: %s", name)
	}
	depth := pathutil.Depth(name)
	if !p.validDepth(depth) {
		return errors.Errorf("incorrect depth: %s depth=%d min=%d max=%d",
			name, depth, p.MinDepth, p.MaxDepth)
	}
	if match, err := p.match(filepath.Base(name)); !match {
		if err != nil {
			return err
		}
		return errors.Errorf("no match found: %s", name)
	}
	return d.onFile(name, p)
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
		if isCloseWrite(ev.Event()) {
			if err := d.dispatch(ev.Path()); err != nil {
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
