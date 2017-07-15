// +build linux

package event

import (
	"github.com/rjeczalik/notify"
)

var flags = []notify.Event{notify.Create, notify.InCloseWrite}

func isCloseWrite(event notify.Event) bool {
	return event == notify.InCloseWrite
}
