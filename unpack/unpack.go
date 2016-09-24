package unpack

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"

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
	sfv, err := sfv.Find(e.Dir())
	if err != nil {
		return nil, err
	}
	archive, err := findArchive(sfv, p.ArchiveExt)
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
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

func (u *Unpack) Stat() (int, int) {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists++
		}
	}
	return exists, len(u.SFV.Checksums)
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

func OnFile(e dispatcher.Event, p dispatcher.Path, log *logrus.Logger) {
	u, err := New(e, p)
	if err != nil {
		log.Error(err)
		return
	}

	if exists, total := u.Stat(); exists != total {
		log.Infof("%s: %d/%d files", u.Event.Dir(), exists, total)
		return
	}

	if err := u.Verify(); err != nil {
		log.WithError(err).Warn("Verification failed")
		return
	}

	log.Info("Verified successfully")

	if err := u.Run(); err != nil {
		log.WithError(err).Error("Unpacking failed")
		return
	}
	log.Info("Unpacked successfully")

	if ok, err := u.Remove(); err != nil {
		log.WithError(err).Error("Failed to delete files")
	} else if ok {
		log.WithFields(logrus.Fields{"path": u.Event.Dir()}).Info("Cleaned up")
	}

	if cmd, err := u.PostRun(); err != nil {
		log.WithFields(logrus.Fields{"command": cmd}).WithError(err).Warn("Failed to run post-command")
	} else if cmd != "" {
		log.WithFields(logrus.Fields{"command": cmd}).Info("Executed post-command")
	}
}
