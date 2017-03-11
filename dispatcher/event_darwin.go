// +build darwin

package dispatcher

import "github.com/rjeczalik/notify"

var flags = []notify.Event{notify.Create, notify.Write}

func (e *Event) IsDir() bool {
	fsEvent, ok := e.eventInfo.Sys().(*notify.FSEvent)
	if !ok {
		return false
	}
	return fsEvent.Flags&notify.FSEventsIsDir != 0
}

func (e *Event) IsCloseWrite() bool {
	return e.eventInfo.Event() == notify.Write
}
