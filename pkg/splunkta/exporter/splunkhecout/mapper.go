// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhecout

import (
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// kind is the outputs.conf stanza kind this exporter owns.
const kind = "hecout"

// MapOutputs builds one config per [hecout] stanza.
func mapOutputs(outputs []splunkconf.Output) (map[string]component.Config, error) {
	out := map[string]component.Config{}
	for _, o := range outputs {
		if o.Name.Kind != kind {
			continue
		}
		out[o.Configuration.Stanza.Name] = ConfigFromOutput(o.Output)
	}
	return out, nil
}
