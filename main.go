package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/martinp/gounpack/dispatcher"
	"github.com/martinp/gounpack/unpack"
	"log"
	"os"
)

func main() {
	var opts struct {
		Config string `short:"c" long:"config" description:"Config file" value-name:"FILE" default:"config.json"`
		Colors bool   `short:"p" long:"colors" description:"Use colors in log output"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	unpack.Colorize.Disable = !opts.Colors
	cfg, err := dispatcher.ReadConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}

	d, err := dispatcher.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := d.Watch(); err != nil {
		log.Print(err)
	}

	d.OnFile = unpack.OnFile
	d.Serve()
}
