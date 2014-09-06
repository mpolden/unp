package main

import (
	"github.com/martinp/gosfv"
	"io/ioutil"
	"log"
	"path/filepath"
)

func findSFV(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sfv" {
			return filepath.Join(dir, f.Name()), nil
		}
	}
	return "", nil
}

func unpack(e *Event, p *Path) {
	log.Printf("event %s", e)
	log.Printf("path %s", p)
	log.Printf("dir %s", filepath.Dir(e.Name))
	dir := filepath.Dir(e.Name)
	sfvPath, err := findSFV(dir)
	if err != nil {
		log.Print(err)
		return
	}
	if sfvPath == "" {
		log.Print("No SFV found")
	} else {
		log.Printf("sfv: %s", sfvPath)
	}
	sfvFile, err := sfv.Read(sfvPath)
	if err != nil {
		log.Print(err)
		return
	}
	if sfvFile.IsExist() {
		log.Printf("all files in sfv exist!")
		log.Printf("verifying sfv")
		ok, err := sfvFile.Verify()
		if err != nil {
			log.Print(err)
			return
		}
		if ok {
			log.Printf("unpack!")
		}
	}
}
