package rar

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mpolden/sfv"
)

func symlink(t *testing.T, oldname, newname string) {
	if err := os.Symlink(oldname, newname); err != nil {
		t.Fatal(err)
	}
}

func testDir(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata")
}

func TestCmdFrom(t *testing.T) {
	tmpl := "tar -xf {{.Name}} {{.Base}} {{.Dir}}"
	values := event{
		Name: "/foo/bar/baz.rar",
		Base: "baz.rar",
		Dir:  "/foo/bar",
	}
	cmd, err := cmdFrom(tmpl, values)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != values.Dir {
		t.Fatalf("want %q, got %q", values.Dir, cmd.Dir)
	}
	if !strings.Contains(cmd.Path, string(os.PathSeparator)) {
		t.Fatalf("want %q to contain a path separator", cmd.Path)
	}
	if cmd.Args[0] != "tar" {
		t.Fatalf("want %q, got %q", "tar", cmd.Args[0])
	}
	if cmd.Args[1] != "-xf" {
		t.Fatalf("want %q, got %q", "-xf", cmd.Args[1])
	}
	if cmd.Args[2] != values.Name {
		t.Fatalf("want %q, got %q", values.Name, cmd.Args[2])
	}
	if cmd.Args[3] != values.Base {
		t.Fatalf("want %q, got %q", values.Base, cmd.Args[3])
	}
	if cmd.Args[4] != values.Dir {
		t.Fatalf("want %q, got %q", values.Base, cmd.Args[4])
	}
	if _, err := cmdFrom("tar -xf {{.Bar}}", values); err == nil {
		t.Fatal("want error")
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

func TestHandle(t *testing.T) {
	td := testDir(t)

	var tests = []struct {
		file  string
		mtime int64
	}{
		{filepath.Join(td, "test1"), 1498246676},
		{filepath.Join(td, "test2"), 1498246678},
		{filepath.Join(td, "test3"), 1498246680},
		{filepath.Join(td, "test", "test4"), 1498316526},
		{filepath.Join(td, "test"), 1498316526},
		{filepath.Join(td, "nested.rar"), 1498246697},
	}

	defer func() {
		for _, tt := range tests {
			os.Remove(tt.file)
		}
	}()

	// Trigger unpacking by passing in a file contained in testdata
	h := NewHandler()
	if err := h.Handle(tests[0].file, "", false); err != nil {
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

func TestHandleIncomplete(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "unp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	var (
		td       = testDir(t)
		realRAR1 = filepath.Join(td, "test.rar")
		realRAR2 = filepath.Join(td, "test.r00")
		realRAR3 = filepath.Join(td, "test.r01")
		realSFV  = filepath.Join(td, "test.sfv")
		rar1     = filepath.Join(tempdir, filepath.Base(realRAR1))
		rar2     = filepath.Join(tempdir, filepath.Base(realRAR2))
		rar3     = filepath.Join(tempdir, filepath.Base(realRAR3))
		sfv      = filepath.Join(tempdir, filepath.Base(realSFV))
	)

	symlink(t, realSFV, sfv)
	symlink(t, realRAR1, rar1)
	symlink(t, realRAR2, rar2)

	h := NewHandler()

	// Verified checksums are cached while RAR set is incomplete
	want := "incomplete: " + tempdir + ": 2/3 files"
	if err := h.Handle(rar1, "", true); err.Error() != want {
		t.Errorf("want err = %q, got %q", want, err.Error())
	}
	if want, got := 2, len(h.cache); want != got {
		t.Errorf("want len = %d, got %d", want, got)
	}

	// Completing the set clears cache
	symlink(t, realRAR3, rar3)
	if err := h.Handle(rar3, "", true); err != nil {
		t.Fatal(err)
	}
	if want, got := 0, len(h.cache); want != got {
		t.Errorf("want %d cache entries, got %d", want, got)
	}
}
