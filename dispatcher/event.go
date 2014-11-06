package dispatcher

import (
	"golang.org/x/exp/inotify"
	"os"
	"path/filepath"
	"strings"
)

type Event inotify.Event

func (e *Event) Depth() int {
	name := filepath.Clean(e.Name)
	return strings.Count(name, string(os.PathSeparator))
}

func (e *Event) IsDir() bool {
	return e.Mask&inotify.IN_ISDIR == inotify.IN_ISDIR
}

func (e *Event) IsCreate() bool {
	return e.Mask&inotify.IN_CREATE == inotify.IN_CREATE
}

func (e *Event) IsClose() bool {
	return e.Mask&inotify.IN_CLOSE == inotify.IN_CLOSE
}

func (e *Event) IsCloseWrite() bool {
	return e.Mask&inotify.IN_CLOSE_WRITE == inotify.IN_CLOSE_WRITE
}

func (e *Event) Dir() string {
	if e.IsDir() {
		return e.Name
	}
	return filepath.Dir(e.Name)
}

func (e *Event) Base() string {
	return filepath.Base(e.Name)
}

func (e *Event) IsHidden() bool {
	return strings.HasPrefix(e.Base(), ".")
}
