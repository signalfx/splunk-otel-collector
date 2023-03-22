package utils

import "github.com/signalfx/golib/v3/event"

// CloneEvent creates a new instance of the event with a shallow copy of all
// map data structures.
func CloneEvent(ev *event.Event) *event.Event {
	return event.NewWithProperties(ev.EventType, ev.Category, CloneStringMap(ev.Dimensions), CloneInterfaceMap(ev.Properties), ev.Timestamp)
}
