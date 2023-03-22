package emitter

import (
	"strings"

	"github.com/influxdata/telegraf"
)

// Emitter interface to telegraf accumulator for processing metrics from
// telegraf
type Emitter interface {
	// Add is a function used by the telegraf accumulator to emit events
	// through the agent.  Pleaes note that if the emitter is a BatchEmitter
	// you will have to invoke the Send() function to send the batch of
	// datapoints and events collected by the Emit function
	AddMetric(telegraf.Metric)
	// AddTag adds a key/value pair to all measurement tags.  If a key conflicts
	// the key value pair in AddTag will override the original key on the
	// measurement
	AddTag(key string, val string)
	// AddTags adds a map of key value pairs to all measurement tags.  If a key
	// conflicts the key value pair in AddTags will override the original key on
	// the measurement.
	AddTags(tags map[string]string)
	// IncludeEvent a thread safe function for registering an event name to
	// include during emission. We disable all events by default because
	// Telegraf has some junk events.
	IncludeEvent(name string)
	// IncludeEvents is a thread safe function for registering a list of event
	// names to include during emission. We disable all events by default
	// because Telegraf has some junk events.
	IncludeEvents(names []string)
	// ExcludeDatum adds a name to the list of metrics and events to
	// exclude
	ExcludeDatum(name string)
	// ExcludeData adds a list of names the list of metrics and events
	// to exclude
	ExcludeData(names []string)
	// OmitTag adds a tag to the list of tags to remove from measurements
	OmitTag(tag string)
	// OmitTags adds a list of tags the list of tags to remove from measurements
	OmitTags(tags []string)
	// AddError handles errors added to the accumulator by telegraf plugins
	// the default behavior is to log the error
	AddError(err error)
	// AddDebug logs a debug statement through the emitter
	AddDebug(deb string, args ...interface{})
}

// RenameFieldWithTag - takes the value of a specified tag and uses it to rename a specified field
// the tag is deleted and the original field name is overwritten
func RenameFieldWithTag(m telegraf.Metric, tagName string, fieldName string, replacer *strings.Replacer) {
	if tagVal, ok := m.GetTag(tagName); ok {
		tagVal = replacer.Replace(tagVal)
		if val, ok := m.GetField(fieldName); ok && tagVal != "" {
			m.AddField(tagVal, val)
			m.RemoveField(fieldName)
			m.RemoveTag(tagName)
		}
	}
}
