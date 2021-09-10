package rar

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/mpolden/sfv"
	"github.com/mpolden/unp/executil"
	"github.com/nwaples/rardecode"
)

var rarPartRE = regexp.MustCompile(`\.part0*(\d+)\.rar$`)

type event struct {
	Base string
	Dir  string
	Name string
	sfv  *sfv.SFV
}

type Handler struct {
	mu    sync.Mutex
	cache map[string]bool
}

func eventFrom(filename string) (event, error) {
	dir := filepath.Dir(filename)
	sfv, err := sfv.Find(dir)
	if err != nil {
		return event{}, err
	}
	rar, err := findFirstRAR(sfv)
	if err != nil {
		return event{}, err
	}
	return event{
		sfv:  sfv,
		Base: filepath.Base(rar),
		Dir:  dir,
		Name: rar,
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
	return "", fmt.Errorf("no rar found in %s", s.Path)
}

func chtimes(name string, header *rardecode.FileHeader) error {
	if header.ModificationTime.IsZero() {
		return nil
	}
	return os.Chtimes(name, header.ModificationTime, header.ModificationTime)
}

func unpack(filename string) error {
	r, err := rardecode.OpenReader(filename, "")
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filename, err)
	}
	dir := filepath.Dir(filename)
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Join(dir, header.Name)
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
			return fmt.Errorf("failed to create file: %s: %w", name, err)
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
			if err := unpack(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Handler) remove(sfv *sfv.SFV) error {
	for _, c := range sfv.Checksums {
		if err := os.Remove(c.Path); err != nil {
			return err
		}
		delete(h.cache, c.Path)
	}
	return os.Remove(sfv.Path)
}

func NewHandler() *Handler { return &Handler{cache: make(map[string]bool)} }

func (h *Handler) verify(sfv *sfv.SFV) (int, int, error) {
	passed := 0
	for _, c := range sfv.Checksums {
		ok := h.cache[c.Path]
		if !ok && c.IsExist() {
			var err error
			ok, err = c.Verify()
			if err != nil {
				return 0, 0, err
			}
			h.cache[c.Path] = ok
		}
		if ok {
			passed++
		}
	}
	return passed, len(sfv.Checksums), nil
}

func (h *Handler) Handle(name, postCommand string, removeRARs bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ev, err := eventFrom(name)
	if err != nil {
		return err
	}
	passed, total, err := h.verify(ev.sfv)
	if err != nil {
		return fmt.Errorf("verification failed: %s: %w", ev.Dir, err)
	}
	if passed != total {
		return fmt.Errorf("incomplete: %s: %d/%d files", ev.Dir, passed, total)
	}
	if err := unpack(ev.Name); err != nil {
		return fmt.Errorf("unpacking failed: %s: %w", ev.Dir, err)
	}
	if removeRARs {
		if err := h.remove(ev.sfv); err != nil {
			return fmt.Errorf("removal failed: %s: %w", ev.Dir, err)
		}
	}
	cd := executil.CommandData{Base: ev.Base, Dir: ev.Dir, Name: ev.Name}
	if err := executil.Run(postCommand, cd); err != nil {
		return fmt.Errorf("post-process command failed: %s: %w", ev.Dir, err)
	}
	return nil
}
