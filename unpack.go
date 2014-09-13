package main

import (
	"fmt"
	"github.com/martinp/gosfv"
	"github.com/mitchellh/colorstring"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Unpack struct {
	SFV   *sfv.SFV
	Event *Event
	Path  *Path
}

func findSFV(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sfv" {
			return filepath.Join(path, f.Name()), nil
		}
	}
	return "", nil
}

func readSFV(path string) (*sfv.SFV, error) {
	sfvPath, err := findSFV(path)
	if err != nil {
		return nil, err
	}
	if sfvPath == "" {
		return nil, fmt.Errorf("no sfv found in %s", path)
	}
	sfv, err := sfv.Read(sfvPath)
	if err != nil {
		return nil, err
	}
	return sfv, nil
}

func (u *Unpack) findArchive() (string, error) {
	for _, c := range u.SFV.Checksums {
		if filepath.Ext(c.Path) == u.Path.ArchiveExtWithDot() {
			return c.Path, nil
		}
	}
	return "", fmt.Errorf("no archive file found in %s", u.SFV.Path)
}

func (u *Unpack) Run() error {
	archive, err := u.findArchive()
	if err != nil {
		return err
	}
	values := CommandValues{
		Name: archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewUnpackCommand(values)
	if err != nil {
		log.Printf("Failed to create command: %s", err)
		return nil
	}
	log.Printf(colorstring.Color("[yellow]Unpacking: %s[reset]"), archive)
	if err := cmd.Run(); err != nil {
		return err
	}
	log.Print(colorstring.Color("[green]File unpacked[reset]"))
	return nil
}

func (u *Unpack) RemoveFiles() error {
	if !u.Path.Remove {
		return nil
	}
	log.Print(colorstring.Color(
		"[yellow]Removing archive files and SFV[reset]"))
	for _, c := range u.SFV.Checksums {
		log.Printf(colorstring.Color("[yellow]Removing: %s[reset]"),
			c.Path)
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	log.Printf(colorstring.Color("[yellow]Removing: %s[reset]"), u.SFV.Path)
	if err := os.Remove(u.SFV.Path); err != nil {
		return err
	}
	log.Print(colorstring.Color("[yellow]Files removed[reset]"))
	return nil
}

func (u *Unpack) AllFilesExist() bool {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists += 1
		}
	}
	if exists != len(u.SFV.Checksums) {
		log.Printf("%d/%d files in %s", exists, len(u.SFV.Checksums),
			u.SFV.Path)
		return false
	}
	return true
}

func (u *Unpack) VerifyFiles() bool {
	log.Printf(colorstring.Color("[yellow]Verifying SFV: %s[reset]"),
		u.SFV.Path)
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			log.Print(err)
			return false
		}
		if !ok {
			log.Printf(colorstring.Color(
				"[red]Invalid checksum: %s[reset]"), c.Path)
			return false
		}
	}
	log.Print(colorstring.Color("[green]All files OK[reset]"))
	return true
}

func onFile(e *Event, p *Path) {
	sfv, err := readSFV(e.Dir())
	if err != nil {
		log.Print(err)
		return
	}

	u := Unpack{
		Event: e,
		Path:  p,
		SFV:   sfv,
	}

	if !u.AllFilesExist() || !u.VerifyFiles() {
		return
	}

	if err := u.Run(); err != nil {
		log.Printf("Failed to unpack: %s", err)
		return
	}

	if err := u.RemoveFiles(); err != nil {
		log.Printf("Failed to delete files: %s", err)
	}
}
