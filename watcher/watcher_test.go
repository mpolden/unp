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

func testWatcher(dir string, onFile OnFile) *watcher {
	cfg := Config{
		BufferSize: 10,
		Paths:      []Path{{Name: dir, MaxDepth: 100, Patterns: []string{"*"}}},
	}
	log := log.New(ioutil.Discard, "", 0)
	return New(cfg, onFile, log)
}

func awaitFile(files *[]string, file string) (bool, error) {
	ts := time.Now()
	for len(*files) == 0 {
		time.Sleep(10 * time.Millisecond)
		if time.Since(ts) > 2*time.Second {
			return false, fmt.Errorf("timed out waiting for file notification")
		}
	}
	return (*files)[0] == file, nil
}

func TestWatching(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	var files []string
	w := testWatcher(dir, func(name, postCommand string, remove bool) error {
		files = append(files, name)
		return nil
	})
	w.goServe()
	w.watch()
	defer w.Stop()

	f := filepath.Join(dir, "foo")
	if err := ioutil.WriteFile(f, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := awaitFile(&files, f)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f, files[0])
	}
}

func TestRescanning(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	var files []string
	w := testWatcher(dir, func(name, postCommand string, remove bool) error {
		files = append(files, name)
		return nil
	})
	defer w.Stop()

	// File is written before watcher is started
	f := filepath.Join(dir, "foo")
	if err := ioutil.WriteFile(f, []byte{0}, 0644); err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatal("no files expected yet")
	}

	w.goServe()
	w.watch()

	// USR1 triggers rescan
	w.signal <- syscall.SIGUSR1

	ok, err := awaitFile(&files, f)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f, files[0])
	}
}

func TestReloading(t *testing.T) {
	var files []string
	w := testWatcher("", func(name, postCommand string, remove bool) error {
		files = append(files, name)
		return nil
	})
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

	ok, err := awaitFile(&files, f)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("want %s, got %s", f, files[0])
	}
}
