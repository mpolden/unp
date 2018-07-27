package rar

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/mpolden/sfv"
	"github.com/nwaples/rardecode"
	"github.com/pkg/errors"
)

var rarPartRE = regexp.MustCompile(`\.part0*(\d+)\.rar$`)

type archive struct {
	Name string
	Dir  string
	Base string
}

type unpacker struct {
	sfv  *sfv.SFV
	dir  string
	name string
}

func newCmd(tmpl string, a archive) (*exec.Cmd, error) {
	t, err := template.New("cmd").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, a); err != nil {
		return nil, err
	}
	argv := strings.Split(b.String(), " ")
	if len(argv) == 0 {
		return nil, errors.New("template compiled to empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = a.Dir
	return cmd, nil
}

func newUnpacker(dir string) (*unpacker, error) {
	sfv, err := sfv.Find(dir)
	if err != nil {
		return nil, err
	}
	rar, err := findFirstRAR(sfv)
	if err != nil {
		return nil, err
	}
	return &unpacker{
		sfv:  sfv,
		dir:  dir,
		name: rar,
	}, nil
}

func isRAR(name string) bool { return filepath.Ext(name) == ".rar" }

func isFirstRAR(name string) bool {
	m := rarPartRE.FindStringSubmatch(name)
	if len(m) == 2 {
		return m[1] == "1"
	}
	return isRAR(name)
}

func findFirstRAR(s *sfv.SFV) (string, error) {
	for _, c := range s.Checksums {
		if isFirstRAR(c.Path) {
			return c.Path, nil
		}
	}
	return "", errors.Errorf("no rar found in %s", s.Path)
}

func chtimes(name string, header *rardecode.FileHeader) error {
	if header.ModificationTime.IsZero() {
		return nil
	}
	return os.Chtimes(name, header.ModificationTime, header.ModificationTime)
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
		name := filepath.Join(u.dir, header.Name)
		// If entry is a directory, create it and set correct ctime
		if header.IsDir {
			if err := os.MkdirAll(name, 0755); err != nil {
				return err
			}
			if err := chtimes(name, header); err != nil {
				return err
			}
			continue
		}
		// Files can come before their containing folders, ensure that parent is created
		parent := filepath.Dir(name)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return err
		}
		if err := chtimes(parent, header); err != nil {
			return err
		}
		// Unpack file
		f, err := os.Create(name)
		if err != nil {
			return errors.Wrapf(err, "failed to create file: %s", name)
		}
		if _, err = io.Copy(f, r); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		// Set correct ctime of unpacked file
		if err := chtimes(name, header); err != nil {
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

func (u *unpacker) Run(removeRARs bool) error {
	if exists, total := u.fileCount(); exists != total {
		return errors.Errorf("incomplete: %s: %d/%d files", u.dir, exists, total)
	}
	if err := u.verify(); err != nil {
		return errors.Wrapf(err, "verification failed: %s", u.dir)
	}
	if err := u.unpack(u.name); err != nil {
		return errors.Wrapf(err, "unpacking failed: %s", u.dir)
	}
	if removeRARs {
		if err := u.remove(); err != nil {
			return errors.Wrapf(err, "removal failed: %s", u.dir)
		}
	}
	return nil
}

func postProcess(u *unpacker, command string) error {
	if command == "" {
		return nil
	}
	cmd, err := newCmd(command, archive{
		Name: u.name,
		Base: filepath.Base(u.dir),
		Dir:  u.dir,
	})
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "stderr: %q", stderr.String())
	}
	return nil
}

func Unpack(name, postCommand string, remove bool) error {
	path := filepath.Dir(name)
	u, err := newUnpacker(path)
	if err != nil {
		return err
	}
	if err := u.Run(remove); err != nil {
		return err
	}
	if err := postProcess(u, postCommand); err != nil {
		return errors.Wrapf(err, "post-process command failed: %s", path)
	}
	return nil
}
