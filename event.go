package main // import "go.sbr.pm/ram"

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

type eventOpt struct {
	value fsnotify.Op
}

func (o *eventOpt) Set(value string) error {
	var op fsnotify.Op
	switch value {
	case "create":
		op = fsnotify.Create
	case "write":
		op = fsnotify.Write
	case "remove":
		op = fsnotify.Remove
	case "rename":
		op = fsnotify.Rename
	case "chmod":
		op = fsnotify.Chmod
	default:
		return fmt.Errorf("unknown event: %s", value)
	}
	o.value = o.value | op
	return nil
}

func (o *eventOpt) Type() string {
	return "event"
}

func (o *eventOpt) String() string {
	return fmt.Sprintf("%s", o.value)
}

func (o *eventOpt) Value() fsnotify.Op {
	if o.value == 0 {
		return fsnotify.Write | fsnotify.Create
	}
	return o.value
}
