// +build linux

package dispatcher

import (
	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.Create, notify.InCloseWrite}

func (e *Event) IsCloseWrite() bool {
	return e.eventInfo.Event() == notify.InCloseWrite
}
