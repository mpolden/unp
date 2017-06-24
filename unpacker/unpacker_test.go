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
	files := []string{
		"test1",
		"test2",
		"test3",
		filepath.Join("test/test4"),
		"nested.rar",
	}
	u, err := New(testdata)
	if err != nil {
		t.Fatal(err)
	}
	if err := u.Run(false); err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, f := range files {
			os.Remove(filepath.Join(testdata, f))
		}
		os.Remove(filepath.Join(testdata, "test"))
	}()
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(testdata, f)); os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
}
