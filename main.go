package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/martinp/gounpack/dispatcher"
	"github.com/martinp/gounpack/unpack"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var opts struct {
		BufferSize int    `short:"b" long:"buffer-size" description:"Number of events to buffer" value-name:"COUNT" default:"100"`
		Config     string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.gounpackrc"`
		Colors     bool   `short:"c" long:"colors" description:"Use colors in log output"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	unpack.Colorize.Disable = !opts.Colors
	if opts.Config == "~/.gounpackrc" {
		home := os.Getenv("HOME")
		opts.Config = filepath.Join(home, ".gounpackrc")
	}
	cfg, err := dispatcher.ReadConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}

	d, err := dispatcher.New(cfg, opts.BufferSize)
	if err != nil {
		log.Fatal(err)
	}

	if err := d.Watch(); err != nil {
		log.Print(err)
	}

	d.OnFile = unpack.OnFile
	d.Serve()
}
