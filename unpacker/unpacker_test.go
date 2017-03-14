package unpacker

import (
	"testing"

	"github.com/mpolden/sfv"
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
