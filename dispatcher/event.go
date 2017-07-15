package dispatcher

import (
	"github.com/rjeczalik/notify"
)

type Event struct {
	eventInfo notify.EventInfo
}

func (e *Event) Name() string {
	return e.eventInfo.Path()
}
