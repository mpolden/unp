package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

func main() {
	var opts struct {
		Config string `short:"c" long:"config" description:"Config file" value-name:"FILE" default:"config.json"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	cfg, err := readConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}

	w, err := New()
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range cfg.Paths {
		if err := w.AddWatch(path); err != nil {
			log.Fatal(err)
		}
	}

	w.OnFile = onFile
	w.Serve()
}
