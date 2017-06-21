package unpacker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mpolden/sfv"
	"github.com/mpolden/unpacker/dispatcher"
)

func TestFindArchive(t *testing.T) {
	sfv := &sfv.SFV{Checksums: []sfv.Checksum{
		sfv.Checksum{Path: "foo.r00"},
		sfv.Checksum{Path: "foo.rar"},
	}}
	archive, err := findArchive(sfv, ".rar")
	if err != nil {
		t.Fatal(err)
	}
	if expected := "foo.rar"; archive != expected {
		t.Errorf("Expected %q, got %q", expected, archive)
	}
	if _, err := findArchive(sfv, ".bar"); err == nil {
		t.Error("Expected error")
	}
}

func TestUnpacking(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testdata := filepath.Join(wd, "testdata")
	path := dispatcher.Path{ArchiveExt: ".rar"}
	unpackedFile := filepath.Join(testdata, "test")
	u, err := New(testdata, path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(unpackedFile)
	if err := u.Run(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(unpackedFile); os.IsNotExist(err) {
		t.Fatal(err)
	}
}
