package unpacker

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/mpolden/sfv"
	"github.com/mpolden/unpacker/dispatcher"
	"github.com/nwaples/rardecode"
	"github.com/pkg/errors"
)

type unpacker struct {
	sfv         *sfv.SFV
	dir         string
	path        dispatcher.Path
	archivePath string
}

func New(dir string, p dispatcher.Path) (*unpacker, error) {
	sfv, err := sfv.Find(dir)
	if err != nil {
		return nil, err
	}
	archive, err := findArchive(sfv, p.ArchiveExt)
	if err != nil {
		return nil, err
	}
	return &unpacker{
		sfv:         sfv,
		dir:         dir,
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
	return "", errors.Errorf("no archive file found in %s", s.Path)
}

func (u *unpacker) unpack() error {
	r, err := rardecode.OpenReader(u.archivePath, "")
	if err != nil {
		return errors.Wrapf(err, "failed to unpack: %s", u.archivePath)
	}
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.IsDir {
			return errors.Errorf("unexpected directory in %s: %s", u.archivePath, header.Name)
		}
		name := filepath.Join(u.dir, header.Name)
		f, err := os.Create(name)
		if err != nil {
			return errors.Wrapf(err, "failed to create file: %s", name)
		}
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
		if err1 := f.Close(); err1 != nil {
			err = err1
		}
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
		Base: filepath.Base(u.dir),
		Dir:  u.dir,
	}
	cmd, err := newCmd(u.path.PostCommand, values)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Errorf("%s: %s", err, stderr.String())
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
			return errors.Errorf("%s: failed checksum: %s", u.sfv.Path, c.Filename)
		}
	}
	return nil
}

func (u *unpacker) Run() error {
	if exists, total := u.fileCount(); exists != total {
		return errors.Errorf("%s is incomplete: %d/%d files", u.dir, exists, total)
	}
	if err := u.verify(); err != nil {
		return errors.Wrapf(err, "verification of %s failed", u.dir)
	}
	if err := u.unpack(); err != nil {
		return errors.Wrapf(err, "unpacking %s failed", u.dir)
	}
	if err := u.remove(); err != nil {
		return errors.Wrapf(err, "cleaning up %s failed", u.dir)
	}
	if err := u.postProcess(); err != nil {
		return errors.Wrap(err, "running post-process command failed")
	}
	return nil
}

func OnFile(e dispatcher.Event, p dispatcher.Path) error {
	u, err := New(e.Dir(), p)
	if err != nil {
		return errors.Wrap(err, "failed to initialize unpacker")
	}
	return u.Run()
}
