package rar

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mpolden/sfv"
)

func TestNewCmd(t *testing.T) {
	tmpl := "tar -xf {{.Name}} {{.Base}} {{.Dir}}"
	values := archive{
		Name: "/foo/bar/baz.rar",
		Base: "baz.rar",
		Dir:  "/foo/bar",
	}

	cmd, err := newCmd(tmpl, values)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != values.Dir {
		t.Fatalf("Expected %s, got %s", values.Dir, cmd.Dir)
	}
	if !strings.Contains(cmd.Path, string(os.PathSeparator)) {
		t.Fatalf("Expected %s to contain a path separator", cmd.Path)
	}
	if cmd.Args[0] != "tar" {
		t.Fatalf("Expected 'tar', got '%s'", cmd.Args[0])
	}
	if cmd.Args[1] != "-xf" {
		t.Fatalf("Expected '-xf', got '%s'", cmd.Args[1])
	}
	if cmd.Args[2] != values.Name {
		t.Fatalf("Expected '%s', got '%s'", values.Name, cmd.Args[2])
	}
	if cmd.Args[3] != values.Base {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[3])
	}
	if cmd.Args[4] != values.Dir {
		t.Fatalf("Expected '%s', got '%s'", values.Base, cmd.Args[4])
	}

	if _, err := newCmd("tar -xf {{.Bar}}", values); err == nil {
		t.Fatal("Expected error")
	}
}

func TestFindFirstRAR(t *testing.T) {
	var tests = []struct {
		sfv      *sfv.SFV
		firstRAR string
	}{
		{&sfv.SFV{Checksums: []sfv.Checksum{
			{Path: "foo.r00"},
			{Path: "foo.rar"},
			{Path: "foo.r01"},
		}}, "foo.rar"},
		{&sfv.SFV{Checksums: []sfv.Checksum{
			{Path: "foo.part3.rar"},
			{Path: "foo.part2.rar"},
			{Path: "foo.part1.rar"},
		}}, "foo.part1.rar"},
		{&sfv.SFV{Checksums: []sfv.Checksum{
			{Path: "foo.part03.rar"},
			{Path: "foo.part01.rar"},
			{Path: "foo.part10.rar"},
			{Path: "foo.part02.rar"},
		}}, "foo.part01.rar"},
		{&sfv.SFV{Checksums: []sfv.Checksum{
			{Path: "foo.part003.rar"},
			{Path: "foo.part100.rar"},
			{Path: "foo.part001.rar"},
			{Path: "foo.part002.rar"},
		}}, "foo.part001.rar"},
	}

	for i, tt := range tests {
		got, err := findFirstRAR(tt.sfv)
		if err != nil {
			t.Fatal(err)
		}
		if want := tt.firstRAR; want != got {
			t.Errorf("#%d: want %q, got %q", i, want, got)
		}
	}
}

func TestUnpack(t *testing.T) {
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
	if err := Unpack(tests[0].file, "", false); err != nil {
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
		_, offset := time.Unix(tt.mtime, 0).Zone()
		if got := fi.ModTime().Unix() + int64(offset); got != tt.mtime {
			t.Errorf("#%d: want mtime = %d, got %d for file %s", i, tt.mtime, got, tt.file)
		}
	}

}
