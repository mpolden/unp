package dispatcher

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/rjeczalik/notify"
)

type OnFile func(Event, Path, *logrus.Logger)

type Dispatcher struct {
	Config
	onFile  OnFile
	watcher chan notify.EventInfo
	log     *logrus.Logger
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
		d.log.WithFields(logrus.Fields{"path": path}).Info("New directory")
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
		go d.onFile(e, p, d.log)
	} else {
		d.onFile(e, p, d.log)
	}
	return nil
}

func (d *Dispatcher) watch() {
	for _, path := range d.Paths {
		recursivePath := filepath.Join(path.Name, "...")
		if err := notify.Watch(recursivePath, d.watcher, flags...); err != nil {
			d.log.Error(err)
		} else {
			d.log.WithFields(logrus.Fields{"path": path.Name}).Info("Watching recursively")
		}
	}
}

func (d *Dispatcher) reload() {
	for {
		s := <-d.signal
		d.log.Infof("Received %s, reloading configuration", s)
		cfg, err := ReadConfig(d.Config.filename)
		if err == nil {
			d.log.Info("Removing all watches")
			notify.Stop(d.watcher)
			d.Config = cfg
			d.watch()
		} else {
			d.log.WithError(err).Errorf("Failed to read config")
		}
	}
}

func (d *Dispatcher) readEvents() {
	for {
		select {
		case ev := <-d.watcher:
			e := Event{ev}
			if e.IsCreate() && e.IsDir() {
				if err := d.createDir(e); err != nil {
					d.log.WithError(err).Info("Skipping event")
				}
			} else if e.IsCloseWrite() {
				if err := d.processFile(e); err != nil {
					d.log.WithError(err).Info("Skipping event")
				}
			}
		}
	}
}

func (d *Dispatcher) Serve() {
	d.watch()
	go d.reload()
	d.readEvents()
}

func New(cfg Config, bufferSize int, handler OnFile, log *logrus.Logger) *Dispatcher {
	// Buffer events so that we don't miss any
	watcher := make(chan notify.EventInfo, bufferSize)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	return &Dispatcher{
		Config:  cfg,
		watcher: watcher,
		log:     log,
		onFile:  handler,
		signal:  sig,
	}
}
