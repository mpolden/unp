// +build linux

package dispatcher

import (
	"syscall"

	"github.com/rjeczalik/notify"
)

const (
	IsDir        = syscall.IN_ISDIR
	IsCloseWrite = syscall.IN_CLOSE_WRITE
)

type mockEvent struct {
	name         string
	mask         notify.Event
	inotifyEvent *syscall.InotifyEvent
}

func (e *mockEvent) Sys() interface{} { return e.inotifyEvent }

func newEvent(name string, mask notify.Event) Event {
	e := mockEvent{
		name:         name,
		mask:         mask,
		inotifyEvent: &syscall.InotifyEvent{Mask: uint32(mask)},
	}
	return Event{&e}
}
