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
	SFV  *sfv.SFV
	Path dispatcher.Path
	Dir  string
	Name string
}

func New(dir string, p dispatcher.Path) (*unpacker, error) {
	sfv, err := sfv.Find(dir)
	if err != nil {
		return nil, err
	}
	rar, err := findRAR(sfv)
	if err != nil {
		return nil, err
	}
	return &unpacker{
		SFV:  sfv,
		Path: p,
		Dir:  dir,
		Name: rar,
	}, nil
}

func findRAR(s *sfv.SFV) (string, error) {
	for _, c := range s.Checksums {
		if filepath.Ext(c.Path) == ".rar" {
			return c.Path, nil
		}
	}
	return "", errors.Errorf("no rar file found in %s", s.Path)
}

func (u *unpacker) unpack() error {
	r, err := rardecode.OpenReader(u.Name, "")
	if err != nil {
		return errors.Wrapf(err, "failed to unpack: %s", u.Name)
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
			return errors.Errorf("unexpected directory in %s: %s", u.Name, header.Name)
		}
		name := filepath.Join(u.Dir, header.Name)
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
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *unpacker) postProcess() error {
	if u.Path.PostCommand == "" {
		return nil
	}
	values := cmdValues{
		Name: u.Name,
		Base: filepath.Base(u.Dir),
		Dir:  u.Dir,
	}
	cmd, err := newCmd(u.Path.PostCommand, values)
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
			return errors.Errorf("%s: failed checksum: %s", u.SFV.Path, c.Filename)
		}
	}
	return nil
}

func (u *unpacker) Run() error {
	if exists, total := u.fileCount(); exists != total {
		return errors.Errorf("%s is incomplete: %d/%d files", u.Dir, exists, total)
	}
	if err := u.verify(); err != nil {
		return errors.Wrapf(err, "verification of %s failed", u.Dir)
	}
	if err := u.unpack(); err != nil {
		return errors.Wrapf(err, "unpacking %s failed", u.Dir)
	}
	if err := u.remove(); err != nil {
		return errors.Wrapf(err, "cleaning up %s failed", u.Dir)
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
