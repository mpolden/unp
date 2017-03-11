// +build linux

package dispatcher

import (
	"syscall"

	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.Create, notify.InCloseWrite}

func (e *Event) IsDir() bool {
	inotifyEvent, ok := e.eventInfo.Sys().(*syscall.InotifyEvent)
	if !ok {
		return false
	}
	return inotifyEvent.Mask&syscall.IN_ISDIR != 0
}

func (e *Event) IsCloseWrite() bool {
	return e.eventInfo.Event() == notify.InCloseWrite
}
