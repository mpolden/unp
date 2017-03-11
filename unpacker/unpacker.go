package unpack

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/martinp/gounpack/dispatcher"
	"github.com/martinp/sfv"
)

type unpacker struct {
	SFV     *sfv.SFV
	Event   dispatcher.Event
	Path    dispatcher.Path
	Archive string
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

func (u *unpacker) unpack() error {
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

func (u *unpacker) postProcess() error {
	if u.Path.PostCommand == "" {
		return nil
	}
	values := dispatcher.CmdValues{
		Name: u.Archive,
		Base: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
	cmd, err := u.Path.NewPostCmd(values)
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
	if !u.Path.Remove {
		return nil
	}
	for _, c := range u.SFV.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	return os.Remove(u.SFV.Path)
}

func (u *unpacker) fileCount() (int, int) {
	exists := 0
	for _, c := range u.SFV.Checksums {
		if c.IsExist() {
			exists++
		}
	}
	return exists, len(u.SFV.Checksums)
}

func (u *unpacker) verify() error {
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

func OnFile(e dispatcher.Event, p dispatcher.Path, log *log.Logger) {
	u, err := newUnpacker(e, p)
	if err != nil {
		log.Printf("Failed to create unpacker: %s", err)
		return
	}

	exists, total := u.fileCount()
	log.Printf("%s: %d/%d files", u.Event.Dir(), exists, total)
	if exists != total {
		return
	}

	if err := u.verify(); err != nil {
		log.Printf("Verification failed: %s", err)
		return
	}

	if err := u.unpack(); err != nil {
		log.Printf("Unpacking failed: %s", err)
		return
	}

	if err := u.remove(); err != nil {
		log.Printf("Failed to delete files: %s", err)
	}

	if err := u.postProcess(); err != nil {
		log.Printf("Failed to run post-process command: %s", err)
	}
}
