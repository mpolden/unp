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

var colorize colorstring.Colorize

func init() {
	colorize = colorstring.Colorize{
		Colors: colorstring.DefaultColors,
		Reset:  true,
	}
}

type Unpack struct {
	SFV   *sfv.SFV
	Event *Event
	Path  *Path
}

func logColorf(format string, v ...interface{}) {
	log.Printf(colorize.Color(format), v...)
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
	logColorf("[yellow]Unpacking: %s[reset]", archive)
	values := CommandValues{
		Name: archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewUnpackCommand(values)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	logColorf("[green]File(s) unpacked![reset]")
	return nil
}

func (u *Unpack) RemoveFiles() error {
	logColorf("[yellow]Removing archive files and SFV[reset]")
	for _, c := range u.SFV.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	if err := os.Remove(u.SFV.Path); err != nil {
		return err
	}
	logColorf("[green]Files removed![reset]")
	return nil
}

func (u *Unpack) StatFiles() error {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists += 1
		}
	}
	if exists != len(u.SFV.Checksums) {
		return fmt.Errorf("%d/%d files in %s", exists,
			len(u.SFV.Checksums), filepath.Base(u.SFV.Path))
	}
	return nil
}

func (u *Unpack) VerifyFiles() error {
	logColorf("[yellow]Verifying: %s[reset]", u.SFV.Path)
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("failed checksum: %s", c.Filename)
		}
	}
	logColorf("[green]All files OK![reset]")
	return nil
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
	if err := u.StatFiles(); err != nil {
		log.Print(err)
		return
	}
	if err := u.VerifyFiles(); err != nil {
		logColorf("[red]Verification failed: %s[reset]", err)
		return
	}
	if err := u.Run(); err != nil {
		logColorf("[red]Failed to unpack: %s[reset]", err)
		return
	}
	if u.Path.Remove {
		if err := u.RemoveFiles(); err != nil {
			logColorf("[red]Failed to delete files: %s[reset]", err)
		}
	}
}
