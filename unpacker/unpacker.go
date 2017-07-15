package unpacker

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/mpolden/sfv"
	"github.com/mpolden/unpacker/event"
	"github.com/nwaples/rardecode"
	"github.com/pkg/errors"
)

type unpacker struct {
	SFV  *sfv.SFV
	Dir  string
	Name string
}

func New(dir string) (*unpacker, error) {
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
		Dir:  dir,
		Name: rar,
	}, nil
}

func isRAR(name string) bool {
	return filepath.Ext(name) == ".rar"
}

func findRAR(s *sfv.SFV) (string, error) {
	for _, c := range s.Checksums {
		if isRAR(c.Path) {
			return c.Path, nil
		}
	}
	return "", errors.Errorf("no rar file found in %s", s.Path)
}

func (u *unpacker) unpack(name string) error {
	r, err := rardecode.OpenReader(name, "")
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", name)
	}
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Join(u.Dir, header.Name)
		if header.IsDir {
			if err := os.MkdirAll(name, 0755); err != nil {
				return err
			}
			continue
		}
		// Ensure parent directory is created
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err
		}
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
		// Unpack recursively if unpacked file is also a RAR
		if isRAR(name) {
			if err := u.unpack(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *unpacker) remove() error {
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

func (u *unpacker) Run(removeRARs bool) error {
	if exists, total := u.fileCount(); exists != total {
		return errors.Errorf("%s is incomplete: %d/%d files", u.Dir, exists, total)
	}
	if err := u.verify(); err != nil {
		return errors.Wrapf(err, "verification of %s failed", u.Dir)
	}
	if err := u.unpack(u.Name); err != nil {
		return errors.Wrapf(err, "unpacking %s failed", u.Dir)
	}
	if removeRARs {
		if err := u.remove(); err != nil {
			return errors.Wrapf(err, "cleaning up %s failed", u.Dir)
		}
	}
	return nil
}

func postProcess(u *unpacker, command string) error {
	if command == "" {
		return nil
	}
	values := cmdValues{
		Name: u.Name,
		Base: filepath.Base(u.Dir),
		Dir:  u.Dir,
	}
	cmd, err := newCmd(command, values)
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

func OnFile(name string, p event.Path) error {
	u, err := New(name)
	if err != nil {
		return errors.Wrap(err, "failed to initialize unpacker")
	}
	if err := u.Run(p.Remove); err != nil {
		return err
	}
	if err := postProcess(u, p.PostCommand); err != nil {
		return errors.Wrap(err, "post-process command failed")
	}
	return nil
}
