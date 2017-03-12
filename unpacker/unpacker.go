package unpacker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinp/sfv"
	"github.com/martinp/unpacker/dispatcher"
	"github.com/pkg/errors"
)

type unpacker struct {
	sfv         *sfv.SFV
	event       dispatcher.Event
	path        dispatcher.Path
	archivePath string
}

func newUnpacker(e dispatcher.Event, p dispatcher.Path) (*unpacker, error) {
	sfv, err := sfv.Find(e.Dir())
	if err != nil {
		return nil, err
	}
	archive, err := findArchive(sfv, p.ArchiveExt)
	if err != nil {
		return nil, err
	}
	return &unpacker{
		sfv:         sfv,
		event:       e,
		path:        p,
		archivePath: archive,
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

func (u *unpacker) unpack() error {
	values := cmdValues{
		Name: u.archivePath,
		Base: u.event.Base(),
		Dir:  u.event.Dir(),
	}
	cmd, err := newCmd(u.path.UnpackCommand, values)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (u *unpacker) postProcess() error {
	if u.path.PostCommand == "" {
		return nil
	}
	values := cmdValues{
		Name: u.archivePath,
		Base: u.event.Base(),
		Dir:  u.event.Dir(),
	}
	cmd, err := newCmd(u.path.PostCommand, values)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, stderr.String())
	}
	return nil
}

func (u *unpacker) remove() error {
	if !u.path.Remove {
		return nil
	}
	for _, c := range u.sfv.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	return os.Remove(u.sfv.Path)
}

func (u *unpacker) fileCount() (int, int) {
	exists := 0
	for _, c := range u.sfv.Checksums {
		if c.IsExist() {
			exists++
		}
	}
	return exists, len(u.sfv.Checksums)
}

func (u *unpacker) verify() error {
	for _, c := range u.sfv.Checksums {
		ok, err := c.Verify()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s: failed checksum: %s", u.sfv.Path, c.Filename)
		}
	}
	return nil
}

func OnFile(e dispatcher.Event, p dispatcher.Path) error {
	u, err := newUnpacker(e, p)
	if err != nil {
		return errors.Wrap(err, "failed to create unpacker")
	}
	if exists, total := u.fileCount(); exists != total {
		return fmt.Errorf("%s is incomplete: %d/%d files", u.event.Dir(), exists, total)
	}
	if err := u.verify(); err != nil {
		return errors.Wrapf(err, "verification of %s failed", u.event.Dir())
	}
	if err := u.unpack(); err != nil {
		return errors.Wrapf(err, "unpacking %s failed", u.event.Dir())
	}
	if err := u.remove(); err != nil {
		return errors.Wrapf(err, "cleaning up %s failed", u.event.Dir())
	}
	if err := u.postProcess(); err != nil {
		return errors.Wrap(err, "running post-process command failed")
	}
	return nil
}
