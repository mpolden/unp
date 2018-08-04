package watcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type testHandler struct {
	wantFile string
	files    []string
}

func (h *testHandler) Handle(filename, postCommand string, remove bool) error {
	if h.wantFile != "" && filename != h.wantFile {
		return fmt.Errorf("unhandled file: %q", filename)
	}
	h.files = append(h.files, filename)
	return nil
}

func (h *testHandler) awaitFile(file string) (bool, error) {
	ts := time.Now()
	for len(h.files) == 0 {
		time.Sleep(10 * time.Millisecond)
		if time.Since(ts) > 2*time.Second {
			return false, fmt.Errorf("timed out waiting for file notification")
		}
	}
	return h.files[0] == file, nil
}

func tempDir() string {
	dir, err := ioutil.TempDir("", "unp")
	if err != nil {
		panic(err)
	}
	path, err := filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	return path
}

func testWatcher(dir string, handler Handler) *Watcher {
	cfg := Config{
		BufferSize: 10,
		Paths:      []Path{{Name: dir, MaxDepth: 100, Patterns: []string{"*"}}},
	}
	log := log.New(ioutil.Discard, "", 0)
	return New(cfg, handler, log)
}

func TestWatching(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	h := testHandler{}
	w := testWatcher(dir, &h)
	w.goServe()
	w.watch()
	defer w.Stop()

	f := filepath.Join(dir, "foo")
	if err := ioutil.WriteFile(f, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := h.awaitFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f, h.files[0])
	}
}

func TestRescanning(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	f1 := filepath.Join(dir, "foo")
	f2 := filepath.Join(dir, "bar")
	var files []string
	h := &testHandler{wantFile: f1}
	w := testWatcher(dir, h)
	defer w.Stop()

	// Files are written before watcher is started
	if err := ioutil.WriteFile(f1, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(f2, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatal("no files expected yet")
	}

	w.goServe()
	w.watch()

	// USR1 triggers rescan
	w.signal <- syscall.SIGUSR1

	ok, err := h.awaitFile(f1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f1, files[0])
	}
}

func TestReloading(t *testing.T) {
	var files []string
	h := &testHandler{}
	w := testWatcher("", h)
	defer w.Stop()

	tmp, err := ioutil.TempFile("", "unp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	w.config = Config{filename: tmp.Name()}

	// Start serving with empty config
	w.goServe()
	w.watch()

	// Create a new directory
	dir := tempDir()
	defer os.RemoveAll(dir)

	// Write config that references directory
	cfg := fmt.Sprintf(`{"Paths": [{"Name": "%s", "Patterns": ["*"], "MaxDepth": 100}]}`, dir)
	if err := ioutil.WriteFile(tmp.Name(), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	// USR2 triggers config reload
	w.signal <- syscall.SIGUSR2

	// Wait until config is loaded
	ts := time.Now()
	for len(w.config.Paths) == 0 {
		time.Sleep(10 * time.Millisecond)
		if time.Since(ts) > 2*time.Second {
			t.Fatal("timed out waiting for new config")
		}
	}

	f := filepath.Join(dir, "foo")
	if err := ioutil.WriteFile(f, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := h.awaitFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f, files[0])
	}
}
