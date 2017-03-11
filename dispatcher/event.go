package dispatcher

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rjeczalik/notify"
)

type Event struct {
	eventInfo notify.EventInfo
}

func (e *Event) Name() string {
	return e.eventInfo.Path()
}

func (e *Event) Depth() int {
	name := filepath.Clean(e.Name())
	return strings.Count(name, string(os.PathSeparator))
}

func (e *Event) Dir() string {
	if e.IsDir() {
		return e.Name()
	}
	return filepath.Dir(e.Name())
}

func (e *Event) Base() string {
	return filepath.Base(e.Name())
}

func hidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (e *Event) IsHidden() bool {
	return hidden(e.Base())
}

func (e *Event) IsParentHidden() bool {
	components := strings.Split(e.Dir(), string(os.PathSeparator))
	for _, c := range components {
		if hidden(c) {
			return true
		}
	}
	return false
}
