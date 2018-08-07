package watcher

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"path/filepath"

	"log"

	"github.com/mpolden/unp/pathutil"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
)

type Handler interface {
	Handle(filename, postCommand string, remove bool) error
	Stop()
}

type Watcher struct {
	config  Config
	handler Handler
	events  chan notify.EventInfo
	signal  chan os.Signal
	done    chan bool
	log     *log.Logger
	mu      sync.Mutex
	wg      sync.WaitGroup
}

func (w *Watcher) handle(name string) error {
	p, ok := w.config.findPath(name)
	if !ok {
		return errors.Errorf("no configured path found: %s", name)
	}
	if p.SkipHidden && pathutil.ContainsHidden(name) {
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
	return w.handler.Handle(name, p.PostCommand, p.Remove)
}

func (w *Watcher) watch() {
	for _, path := range w.config.Paths {
		rpath := filepath.Join(path.Name, "...")
		if err := notify.Watch(rpath, w.events, writeFlag); err != nil {
			w.log.Printf("failed to watch %s: %s", rpath, err)
		} else {
			w.log.Printf("watching %s recursively", path.Name)
		}
	}
}

func (w *Watcher) reload() {
	cfg, err := ReadConfig(w.config.filename)
	if err == nil {
		notify.Stop(w.events)
		w.config = cfg
		w.watch()
	} else {
		w.log.Printf("failed to read config: %s", err)
	}
}

func (w *Watcher) rescan() {
	for _, p := range w.config.Paths {
		err := filepath.Walk(p.Name, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info == nil || !info.Mode().IsRegular() {
				return nil
			}
			if err := w.handle(path); err != nil {
				w.log.Print(err)
			}
			return nil
		})
		if err != nil {
			w.log.Printf("failed to rescan %s: %s", p.Name, err)
		}
	}
}

func (w *Watcher) readSignal() {
	for {
		select {
		case <-w.done:
			return
		case s := <-w.signal:
			w.mu.Lock()
			switch s {
			case syscall.SIGUSR1:
				w.log.Printf("received %s: rescanning watched directories", s)
				w.rescan()
			case syscall.SIGUSR2:
				w.log.Printf("received %s: reloading configuration", s)
				w.reload()
			case syscall.SIGTERM, syscall.SIGINT:
				w.log.Printf("received %s: shutting down", s)
				w.Stop()
			}
			w.mu.Unlock()
		}
	}
}

func (w *Watcher) readEvent() {
	for {
		select {
		case <-w.done:
			return
		case ev := <-w.events:
			w.mu.Lock()
			if err := w.handle(ev.Path()); err != nil {
				w.log.Print(err)
			}
			w.mu.Unlock()
		}
	}
}

func (w *Watcher) goServe() {
	w.wg.Add(2)
	go func() {
		defer w.wg.Done()
		w.readSignal()
	}()
	go func() {
		defer w.wg.Done()
		w.readEvent()
	}()
}

func (w *Watcher) Start() {
	w.goServe()
	w.watch()
	w.wg.Wait()
}

func (w *Watcher) Stop() {
	w.handler.Stop()
	notify.Stop(w.events)
	w.done <- true
	w.done <- true
}

func New(cfg Config, handler Handler, log *log.Logger) *Watcher {
	// Buffer events so that we don't miss any
	events := make(chan notify.EventInfo, cfg.BufferSize)
	sig := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sig)
	return &Watcher{
		config:  cfg,
		events:  events,
		log:     log,
		handler: handler,
		signal:  sig,
		done:    done,
	}
}
