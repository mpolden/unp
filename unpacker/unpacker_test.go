package unpacker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mpolden/sfv"
)

func TestFindRAR(t *testing.T) {
	sfv := &sfv.SFV{Checksums: []sfv.Checksum{
		sfv.Checksum{Path: "foo.r00"},
		sfv.Checksum{Path: "foo.rar"},
	}}
	archive, err := findRAR(sfv)
	if err != nil {
		t.Fatal(err)
	}
	if want := "foo.rar"; archive != want {
		t.Errorf("want %q, got %q", want, archive)
	}
}

func TestUnpacking(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testdata := filepath.Join(wd, "testdata")
	f1 := filepath.Join(testdata, "test1")
	f2 := filepath.Join(testdata, "test2")
	u, err := New(testdata)
	if err != nil {
		t.Fatal(err)
	}
	if err := u.Run(false); err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(f1)
		os.Remove(f2)
	}()
	if _, err := os.Stat(f1); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(f2); os.IsNotExist(err) {
		t.Fatal(err)
	}
}
