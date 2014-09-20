package unpack

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFindSFV(t *testing.T) {
	dir, err := ioutil.TempDir("", "gounpack")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "test.sfv")
	if err := ioutil.WriteFile(file, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}
	sfv, err := findSFV(dir)
	if err != nil {
		t.Fatal(err)
	}
	if sfv != file {
		t.Fatalf("Expected %s, got %s", file, sfv)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if _, err := findSFV(dir); err == nil {
		t.Fatal("Expected error")
	}
	os.Remove(dir) // ignore error
}

func TestReadSFV(t *testing.T) {
	dir, err := ioutil.TempDir("", "gounpack")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "test.sfv")
	const sfv = "foo 7E3265A8\nbar 04A2B3E9"
	if err := ioutil.WriteFile(file, []byte(sfv), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readSFV(dir); err != nil {
		t.Fatal(err)
	}
	os.Remove(dir) // ignore error
}
