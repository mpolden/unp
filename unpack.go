package main

import (
	"fmt"
	"github.com/martinp/gosfv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Unpack struct {
	SFV      *sfv.SFV
	Path     string
	Patterns []string
	Remove   bool
}

func (u *Unpack) validEvent(e *Event) bool {
	name := filepath.Base(e.Name)
	for _, p := range u.Patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}

func (u *Unpack) findSFV() (string, error) {
	files, err := ioutil.ReadDir(u.Path)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sfv" {
			return filepath.Join(u.Path, f.Name()), nil
		}
	}
	return "", nil
}

func (u *Unpack) readSFV() error {
	sfvPath, err := u.findSFV()
	if err != nil {
		return err
	}
	if sfvPath == "" {
		return fmt.Errorf("no sfv found in %s", u.Path)
	}
	sfv, err := sfv.Read(sfvPath)
	if err != nil {
		return err
	}
	if !sfv.IsExist() {
		return fmt.Errorf("not all files exist")
	}
	u.SFV = sfv
	return nil
}

func (u *Unpack) findRAR() string {
	for _, c := range u.SFV.Checksums {
		if filepath.Ext(c.Path) == ".rar" {
			return c.Path
		}
	}
	return ""
}

func (u *Unpack) unpack() error {
	rar := u.findRAR()
	if rar == "" {
		return fmt.Errorf("No rar file found")
	}
	log.Printf("Unpacking %s", rar)
	cmd := exec.Command("unrar", "x", "-o-", rar, u.Path)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (u *Unpack) removeFiles() error {
	log.Printf("Removing RAR files")
	for _, c := range u.SFV.Checksums {
		log.Printf("Removing %s", c.Path)
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}

func (u *Unpack) onFile(e *Event, p *Path) {
	u.Path = filepath.Dir(e.Name)

	if !u.validEvent(e) {
		return
	}

	err := u.readSFV()
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("Verifying SFV: %s", u.SFV.Path)
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			log.Print(err)
			return
		}
		if !ok {
			log.Printf("Invalid checksum: %s", c.Path)
			return
		}

	}

	if err := u.unpack(); err != nil {
		log.Printf("Failed to unpack: %s", err)
	}

	if u.Remove {
		if err := u.removeFiles(); err != nil {
			log.Printf("Failed to delete files: %s", err)
		}
	}
}
