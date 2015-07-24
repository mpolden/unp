package unpack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	sfv "github.com/martinp/gosfv"
	"github.com/martinp/gounpack/dispatcher"
)

type Unpack struct {
	SFV     *sfv.SFV
	Event   dispatcher.Event
	Path    dispatcher.Path
	Archive string
}

func New(e dispatcher.Event, p dispatcher.Path) (*Unpack, error) {
	sfv, err := readSFV(e.Dir())
	if err != nil {
		return nil, err
	}
	archive, err := findArchive(sfv, p.ArchiveExtWithDot())
	if err != nil {
		return nil, err
	}
	return &Unpack{
		SFV:     sfv,
		Event:   e,
		Path:    p,
		Archive: archive,
	}, nil
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

func findArchive(s *sfv.SFV, ext string) (string, error) {
	for _, c := range s.Checksums {
		if filepath.Ext(c.Path) == ext {
			return c.Path, nil
		}
	}
	return "", fmt.Errorf("no archive file found in %s", s.Path)
}

func (u *Unpack) Run() error {
	values := dispatcher.CmdValues{
		Name: u.Archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewUnpackCmd(values)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (u *Unpack) PostRun() (string, error) {
	if u.Path.PostCommand == "" {
		return "", nil
	}
	values := dispatcher.CmdValues{
		Name: u.Archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewPostCmd(values)
	if err != nil {
		return "", err
	}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.Join(cmd.Args, " "), nil
}

func (u *Unpack) Remove() (bool, error) {
	if !u.Path.Remove {
		return false, nil
	}
	for _, c := range u.SFV.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return false, err
		}
	}
	if err := os.Remove(u.SFV.Path); err != nil {
		return false, err
	}
	return true, nil
}

func (u *Unpack) Stat() error {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists++
		}
	}
	if exists != len(u.SFV.Checksums) {
		return fmt.Errorf("%s: %d/%d files", u.SFV.Path, exists, len(u.SFV.Checksums))
	}
	return nil
}

func (u *Unpack) Verify() error {
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s: failed checksum: %s", u.SFV.Path, c.Filename)
		}
	}
	return nil
}

func OnFile(e dispatcher.Event, p dispatcher.Path, m chan<- string) {
	u, err := New(e, p)
	if err != nil {
		m <- err.Error()
		return
	}
	if err := u.Stat(); err != nil {
		m <- err.Error()
		return
	}

	if err := u.Verify(); err != nil {
		m <- fmt.Sprintf("Verification failed: %s", err)
		return
	}
	m <- fmt.Sprintf("Verified: %s", u.SFV.Path)

	if err := u.Run(); err != nil {
		m <- fmt.Sprintf("Failed to unpack %s: %s", u.Archive, err)
		return
	}
	m <- fmt.Sprintf("Unpacked: %s", u.Archive)

	if ok, err := u.Remove(); err != nil {
		m <- fmt.Sprintf("Failed to delete files: %s", err)
	} else if ok {
		m <- fmt.Sprintf("Cleaned up: %s", u.Event.Dir())
	}

	if cmd, err := u.PostRun(); err != nil {
		m <- fmt.Sprintf("Failed to run post command: %s", err)
	} else if cmd != "" {
		m <- fmt.Sprintf("Executed post command: %s", cmd)
	}
}
