// +build linux

package dispatcher

import (
	"syscall"

	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.Create, notify.InCloseWrite}

func (e *Event) IsDir() bool {
	inotifyEvent, ok := e.Sys().(*syscall.InotifyEvent)
	if !ok {
		return false
	}
	return inotifyEvent.Mask&syscall.IN_ISDIR != 0
}

func (e *Event) IsCloseWrite() bool {
	return e.Event() == notify.InCloseWrite
}
