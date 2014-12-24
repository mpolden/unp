package unpack

import (
	"fmt"
	"github.com/martinp/gosfv"
	"github.com/martinp/gounpack/dispatcher"
	"github.com/mitchellh/colorstring"
	"io/ioutil"
	"os"
	"path/filepath"
)

var Colorize colorstring.Colorize

func init() {
	Colorize = colorstring.Colorize{
		Colors: colorstring.DefaultColors,
		Reset:  true,
	}
}

type unpack struct {
	SFV      *sfv.SFV
	Event    dispatcher.Event
	Path     dispatcher.Path
	messages chan<- string
}

func (u *unpack) log(format string, v ...interface{}) {
	u.messages <- fmt.Sprintf(Colorize.Color(format), v...)
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
	return "", fmt.Errorf("no sfv found in %s", path)
}

func readSFV(path string) (*sfv.SFV, error) {
	sfvFile, err := findSFV(path)
	if err != nil {
		return nil, err
	}
	sfv, err := sfv.Read(sfvFile)
	if err != nil {
		return nil, err
	}
	return sfv, nil
}

func (u *unpack) findArchive() (string, error) {
	for _, c := range u.SFV.Checksums {
		if filepath.Ext(c.Path) == u.Path.ArchiveExtWithDot() {
			return c.Path, nil
		}
	}
	return "", fmt.Errorf("no archive file found in %s", u.SFV.Path)
}

func (u *unpack) Run(archive string) error {
	archiveDirBase := dispatcher.DirBase(archive)
	u.log("[yellow]Unpacking: %s[reset]", archiveDirBase)
	values := dispatcher.CommandValues{
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
	u.log("[green]Unpacked: %s[reset]", archiveDirBase)
	return nil
}

func (u *unpack) PostRun(archive string) error {
	if u.Path.PostCommand == "" {
		return nil
	}
	values := dispatcher.CommandValues{
		Name: archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewPostCommand(values)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	u.log("[green]Executed post command: %s[reset]",
		u.Path.PostCommand)
	return nil
}

func (u *unpack) RemoveFiles() error {
	dir := filepath.Dir(u.SFV.Path)
	u.log("[yellow]Cleaning up archives and SFV: %s[reset]", dir)
	for _, c := range u.SFV.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	if err := os.Remove(u.SFV.Path); err != nil {
		return err
	}
	u.log("[green]Cleanup done: %s[reset]", dir)
	return nil
}

func (u *unpack) StatFiles() error {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists += 1
		}
	}
	if exists != len(u.SFV.Checksums) {
		return fmt.Errorf("%s: %d/%d files",
			dispatcher.DirBase(u.SFV.Path),
			exists, len(u.SFV.Checksums))
	}
	return nil
}

func (u *unpack) VerifyFiles() error {
	sfvFile := dispatcher.DirBase(u.SFV.Path)
	u.log("[yellow]Verifying: %s[reset]", sfvFile)
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s: failed checksum: %s", sfvFile,
				c.Filename)
		}
	}
	u.log("[green]OK: %s[reset]", sfvFile)
	return nil
}

func OnFile(e dispatcher.Event, p dispatcher.Path, m chan<- string) {
	sfv, err := readSFV(e.Dir())
	if err != nil {
		m <- err.Error()
		return
	}
	u := unpack{
		Event:    e,
		Path:     p,
		SFV:      sfv,
		messages: m,
	}
	if err := u.StatFiles(); err != nil {
		u.log(err.Error())
		return
	}
	if err := u.VerifyFiles(); err != nil {
		u.log("[red]Verification failed: %s[reset]", err)
		return
	}
	archive, err := u.findArchive()
	if err != nil {
		u.log("[red]File not found: %s[reset]", err)
		return
	}
	if err := u.Run(archive); err != nil {
		u.log("[red]Failed to unpack: %s[reset]", err)
		return
	}
	if err := u.PostRun(archive); err != nil {
		u.log("[red]Failed to run post command: %s[reset]", err)
		return
	}
	if u.Path.Remove {
		if err := u.RemoveFiles(); err != nil {
			u.log("[red]Failed to delete files: %s[reset]", err)
		}
	}
}
