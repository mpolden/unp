package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mpolden/unp/watcher"
)

func init() {
	log.SetPrefix("unp: ")
	log.SetFlags(0)
}

func main() {
	var test bool
	var configFile string
	flag.StringVar(&configFile, "f", "~/.unprc", "Path to config file")
	flag.BoolVar(&test, "t", false, "Test and print config")
	flag.Parse()

	cfg, err := watcher.ReadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if test {
		json, err := cfg.JSON()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
		return
	}

	w := watcher.New(cfg)
	w.Start()
}
