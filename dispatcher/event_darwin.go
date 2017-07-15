// +build darwin

package dispatcher

import "github.com/rjeczalik/notify"

var flags = []notify.Event{notify.Create, notify.Write}

func (e *Event) IsCloseWrite() bool {
	return e.eventInfo.Event() == notify.Write
}
