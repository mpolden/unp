package main

import (
	"fmt"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"

	"github.com/mpolden/unp/rar"
	"github.com/mpolden/unp/watcher"
)

func main() {
	var opts struct {
		Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.unprc"`
		Test   bool   `short:"t" long:"test" description:"Test and print config"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	cfg, err := watcher.ReadConfig(opts.Config)
	if err != nil {
		log.Fatal(err)
	}

	if opts.Test {
		json, err := cfg.JSON()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
		return
	}

	log := log.New(os.Stderr, "unp: ", 0)
	w := watcher.New(cfg, rar.Unpack, log)
	w.Start()
}
