package unpacker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

	var tests = []struct {
		file  string
		mtime int64
	}{
		{filepath.Join(testdata, "test1"), 1498246676},
		{filepath.Join(testdata, "test2"), 1498246678},
		{filepath.Join(testdata, "test3"), 1498246680},
		{filepath.Join(testdata, "test", "test4"), 1498316526},
		{filepath.Join(testdata, "test"), 1498316526},
		{filepath.Join(testdata, "nested.rar"), 1498246697},
	}

	defer func() {
		for _, tt := range tests {
			os.Remove(tt.file)
		}
	}()

	// Trigger unpacking by passing in a file contained in testdata
	if err := OnFile(tests[0].file, "", false); err != nil {
		t.Fatal(err)
	}

	for i, tt := range tests {
		// Verify that file have been unpacked
		fi, err := os.Stat(tt.file)
		if err != nil {
			t.Fatalf("#%d: %s", i, err)
		}
		// rardecode uses time.Local when parsing modification time, so the local time offset must
		// be added when comparing
		_, offset := time.Now().Zone()
		if got := fi.ModTime().Unix() + int64(offset); got != tt.mtime {
			t.Errorf("#%d: want mtime = %d, got %d for file %s", i, tt.mtime, got, tt.file)
		}
	}

}
