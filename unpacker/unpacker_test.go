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
		filepath.Join(testdata, "test1"),
		filepath.Join(testdata, "test2"),
		filepath.Join(testdata, "test3"),
		filepath.Join(testdata, "test", "test4"),
		filepath.Join(testdata, "test"),
		filepath.Join(testdata, "nested.rar"),
	}
	if err := OnFile(files[0], "", false); err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, f := range files {
			os.Remove(f)
		}
	}()
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
}
