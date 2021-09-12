//go:build !linux

package watcher

import "github.com/rjeczalik/notify"

const notifyFlag = notify.Write | notify.Rename
