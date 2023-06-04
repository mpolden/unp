//go:build linux

package watcher

import "github.com/rjeczalik/notify"

// Watch IN_CLOSE_WRITE events on Linux to reduce the number of events to process
const notifyFlag = notify.InCloseWrite | notify.InMovedTo
