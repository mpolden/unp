package main

import (
	"github.com/jessevdk/go-flags"
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

	colorize.Disable = !opts.Colors
	cfg, err := readConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}

	dispatcher, err := New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := dispatcher.Watch(); err != nil {
		log.Print(err)
	}

	dispatcher.OnFile = onFile
	dispatcher.Serve()
}
