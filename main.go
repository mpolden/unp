package main

import (
	"encoding/json"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Paths []Path
}

func readConfig(name string) (*Config, error) {
	if name == "" {
		name = "config.json"
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	var opts struct {
		Config string `short:"c" long:"config" description:"Config file"`
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
