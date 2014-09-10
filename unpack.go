package main

import (
	"bytes"
	"fmt"
	"github.com/martinp/gosfv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Unpack struct {
	SFV     *sfv.SFV
	Event   *Event
	Path    *Path
	RARFile string
}

type templateValues struct {
	Path string
	Dir  string
	File string
}

func createCommand(v templateValues, tmpl string) (*exec.Cmd, error) {
	t := template.Must(template.New("command").Parse(tmpl))
	var b bytes.Buffer
	if err := t.Execute(&b, v); err != nil {
		return nil, err
	}
	argv := strings.Split(b.String(), " ")
	if len(argv) == 0 {
		return nil, fmt.Errorf("template compiled to empty command")
	}
	if len(argv) == 1 {
		return exec.Command(argv[0]), nil
	}
	return exec.Command(argv[0], argv[1:]...), nil
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
	return "", nil
}

func readSFV(path string) (*sfv.SFV, error) {
	sfvPath, err := findSFV(path)
	if err != nil {
		return nil, err
	}
	if sfvPath == "" {
		return nil, fmt.Errorf("no sfv found in %s", path)
	}
	sfv, err := sfv.Read(sfvPath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", sfv.Path)
	if !sfv.IsExist() {
		return nil, fmt.Errorf("not all files exist")
	}
	return sfv, nil
}

func (u *Unpack) values() templateValues {
	return templateValues{
		Path: u.RARFile,
		File: u.Event.Base(),
		Dir:  u.Event.Dir(),
	}
}

func (u *Unpack) findRARFile() (string, error) {
	for _, c := range u.SFV.Checksums {
		if filepath.Ext(c.Path) == ".rar" {
			return c.Path, nil
		}
	}
	return "", fmt.Errorf("no rar file found in %s", u.SFV.Path)
}

func (u *Unpack) Run() error {
	rar, err := u.findRARFile()
	if err != nil {
		return err
	}
	u.RARFile = rar
	cmd, err := createCommand(u.values(), u.Path.UnpackCommand)
	if err != nil {
		log.Printf("Failed to create command: %s", err)
		return nil
	}
	log.Printf("Unpacking: %s", rar)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (u *Unpack) RemoveFiles() error {
	if !u.Path.Remove {
		return nil
	}
	log.Printf("Removing RAR files")
	for _, c := range u.SFV.Checksums {
		log.Printf("Removing: %s", c.Path)
		if err := os.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}

func onFile(e *Event, p *Path) {
	sfv, err := readSFV(e.Dir())
	if err != nil {
		log.Print(err)
		return
	}

	u := Unpack{
		Event: e,
		Path:  p,
		SFV:   sfv,
	}

	log.Printf("Verifying SFV: %s", u.SFV.Path)
	for _, c := range u.SFV.Checksums {
		ok, err := c.Verify()
		if err != nil {
			log.Print(err)
			return
		}
		if !ok {
			log.Printf("Invalid checksum: %s", c.Path)
			return
		}

	}

	if err := u.Run(); err != nil {
		log.Printf("Failed to unpack: %s", err)
		return
	}

	if err := u.RemoveFiles(); err != nil {
		log.Printf("Failed to delete files: %s", err)
	}
}
