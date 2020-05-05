package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mpolden/unp/rar"
	"github.com/mpolden/unp/watcher"
)

func main() {
	configFile := flag.String("f", "~/.unprc", "Path to config file")
	test := flag.Bool("t", false, "Test and print config")
	flag.Parse()

	cfg, err := watcher.ReadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	if *test {
		json, err := cfg.JSON()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", json)
		return
	}

	log := log.New(os.Stderr, "unp: ", 0)
	w := watcher.New(cfg, rar.NewHandler(), log)
	w.Start()
}
