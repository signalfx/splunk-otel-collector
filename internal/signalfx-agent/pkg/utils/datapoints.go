package utils

import (
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
)

// TruncateDimensionValue truncates the given string to 256 characters.
func TruncateDimensionValue(value string) string {
	// Not sure if our backend enforces character length or byte length.
	// If values include multi-byte unicode chars, this might not work.
	if len(value) > 256 {
		return value[:256]
	}
	return value
}

// SetDatapointMeta sets a field on the datapoint.Meta field, initializing the
// Meta map if it is nil.
func SetDatapointMeta(dp *datapoint.Datapoint, name, val interface{}) {
	if dp.Meta == nil {
		dp.Meta = make(map[interface{}]interface{})
	}
	dp.Meta[name] = val
}
