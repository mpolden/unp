package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/mpolden/unp/logutil"
	"github.com/mpolden/unp/watcher"
)

func init() {
	var out io.Writer
	if stderr := os.Stderr; isatty.IsTerminal(stderr.Fd()) {
		out = logutil.NewUniqueWriter(stderr)
	} else {
		out = stderr
	}
	log.SetPrefix("unp: ")
	log.SetFlags(0)
	log.SetOutput(out)
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
