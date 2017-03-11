package main

import (
	"fmt"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"

	"github.com/martinp/unpacker/dispatcher"
	"github.com/martinp/unpacker/unpacker"
)

func main() {
	var opts struct {
		Config string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.unpackerrc"`
		Test   bool   `short:"t" long:"test" description:"Test and print config"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	cfg, err := dispatcher.ReadConfig(opts.Config)
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

	log := log.New(os.Stderr, "", log.LstdFlags)
	dispatcher.New(cfg, unpacker.OnFile, log).Serve()
}
