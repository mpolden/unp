// +build darwin

package dispatcher

import "github.com/rjeczalik/notify"

const (
	IsDir        = notify.FSEventsIsDir
	IsCloseWrite = notify.Write
)

type mockEvent struct {
	name    string
	mask    notify.Event
	fsEvent *notify.FSEvent
}

func (e *mockEvent) Sys() interface{} { return e.fsEvent }

func newEvent(name string, mask notify.Event) Event {
	e := mockEvent{
		name:    name,
		mask:    mask,
		fsEvent: &notify.FSEvent{Flags: uint32(mask)},
	}
	return Event{&e}
}
