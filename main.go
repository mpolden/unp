package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"

	"github.com/martinp/gounpack/dispatcher"
	"github.com/martinp/gounpack/unpack"
)

func main() {
	var opts struct {
		BufferSize int    `short:"b" long:"buffer-size" description:"Number of events to buffer" value-name:"COUNT" default:"100"`
		Config     string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.gounpackrc"`
		Test       bool   `short:"t" long:"test" description:"Test and print config"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	if opts.Config == "~/.gounpackrc" {
		home := os.Getenv("HOME")
		opts.Config = filepath.Join(home, ".gounpackrc")
	}
	cfg, err := dispatcher.ReadConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}
	if opts.Test {
		fmt.Printf("%+v\n", cfg)
		return
	}

	d, err := dispatcher.New(cfg, opts.BufferSize)
	if err != nil {
		log.Fatal(err)
	}

	d.OnFile = unpack.OnFile
	msgs := d.Serve()
	for {
		select {
		case msg := <-msgs:
			log.Print(msg)
		}
	}
}
