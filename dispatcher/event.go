package dispatcher

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rjeczalik/notify"
)

type Event struct {
	notify.EventInfo
}

func (e *Event) Depth() int {
	name := filepath.Clean(e.Path())
	return strings.Count(name, string(os.PathSeparator))
}

func (e *Event) IsDir() bool {
	return e.Sys().(*syscall.InotifyEvent).Mask&syscall.IN_ISDIR != 0
}

func (e *Event) IsCreate() bool {
	return e.Event() == notify.InCreate
}

func (e *Event) IsCloseWrite() bool {
	return e.Event() == notify.InCloseWrite
}

func (e *Event) Dir() string {
	if e.IsDir() {
		return e.Path()
	}
	return filepath.Dir(e.Path())
}

func (e *Event) Base() string {
	return filepath.Base(e.Path())
}

func hidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (e *Event) IsHidden() bool {
	return hidden(e.Base())
}

func (e *Event) IsParentHidden() bool {
	components := strings.Split(e.Dir(), string(os.PathSeparator))
	for _, c := range components {
		if hidden(c) {
			return true
		}
	}
	return false
}
