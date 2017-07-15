package event

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

type watcher struct {
	config Config
	onFile OnFile
	events chan notify.EventInfo
	signal chan os.Signal
	log    *log.Logger
}

func (w *watcher) handle(name string) error {
	p, ok := w.config.findPath(name)
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
	return w.onFile(name, p)
}

func (w *watcher) watch() {
	for _, path := range w.config.Paths {
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, w.events, flags...); err != nil {
			w.log.Printf("Failed to watch %s: %s", recursivePath, err)
		} else {
			w.log.Printf("Watching %s recursively", path.Name)
		}
	}
}

func (w *watcher) reload() {
	for {
		s := <-w.signal
		w.log.Printf("Received %s, reloading configuration", s)
		cfg, err := ReadConfig(w.config.filename)
		if err == nil {
			notify.Stop(w.events)
			w.config = cfg
			w.watch()
		} else {
			w.log.Printf("Failed to read config: %s", err)
		}
	}
}

func (w *watcher) readEvents() {
	for ev := range w.events {
		if isCloseWrite(ev.Event()) {
			if err := w.handle(ev.Path()); err != nil {
				w.log.Printf("Skipping event: %s", err)
			}
		}
	}
}

func (w *watcher) Serve() {
	w.watch()
	go w.reload()
	w.readEvents()
}

func NewWatcher(cfg Config, onFile OnFile, log *log.Logger) *watcher {
	// Buffer events so that we don't miss any
	events := make(chan notify.EventInfo, cfg.BufferSize)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	return &watcher{
		config: cfg,
		events: events,
		log:    log,
		onFile: onFile,
		signal: sig,
	}
}
