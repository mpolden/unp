package event

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "unpacker")
	if err != nil {
		panic(err)
	}
	path, err := filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	return path
}

func newWatcher(dir string, onFile OnFile) *watcher {
	cfg := Config{
		BufferSize: 10,
		Paths:      []Path{{Name: dir, MaxDepth: 100, Patterns: []string{"*"}}},
	}
	log := log.New(ioutil.Discard, "", log.LstdFlags)
	return NewWatcher(cfg, onFile, log)
}

func TestWatching(t *testing.T) {
	var files []string
	onFile := func(name string, path Path) error {
		files = append(files, name)
		return nil
	}
	dir := tempDir()
	f := filepath.Join(dir, "foo")
	w := newWatcher(dir, onFile)
	w.start()
	w.watch()
	defer w.Stop()
	defer os.RemoveAll(dir)

	if err := ioutil.WriteFile(f, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	// Sleep until file is caught
	ts := time.Now()
	for len(files) == 0 {
		time.Sleep(10 * time.Millisecond)
		if time.Since(ts) > 2*time.Second {
			t.Fatal("timed out waiting for file notification")
		}

	}
	if files[0] != f {
		t.Errorf("want %s, got %s", f, files[0])
	}
}
