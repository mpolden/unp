// +build darwin

package dispatcher

import "github.com/rjeczalik/notify"

var flags = []notify.Event{notify.Create, notify.Write}

func isCloseWrite(event notify.Event) bool {
	return event == notify.Write
}
