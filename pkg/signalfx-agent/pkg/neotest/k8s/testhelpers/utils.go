package testhelpers

import (
	sfxproto "github.com/signalfx/com_signalfx_metrics_protobuf"
)

// ProtoDimensionsToMap takes a slice of protobuf dimensions and turns them
// into a regular map for easier testing.
func ProtoDimensionsToMap(dims []*sfxproto.Dimension) (m map[string]string) {
	m = make(map[string]string)

	for _, d := range dims {
		m[d.GetKey()] = d.GetValue()
	}
	return m
}
