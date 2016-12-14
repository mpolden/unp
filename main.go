package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"

	"github.com/martinp/gounpack/dispatcher"
	"github.com/martinp/gounpack/unpack"
)

func main() {
	var opts struct {
		Config   string `short:"f" long:"config" description:"Config file" value-name:"FILE" default:"~/.gounpackrc"`
		Test     bool   `short:"t" long:"test" description:"Test and print config"`
		LogLevel string `short:"L" long:"log-level" description:"Log level to use" default:"info" choice:"debug" choice:"info" choice:"warn" choice:"error" choice:"fatal" choice:"panic"`
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

	log := logrus.New()
	level, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.Level = level

	dispatcher.New(cfg, unpack.OnFile, log).Serve()
}
