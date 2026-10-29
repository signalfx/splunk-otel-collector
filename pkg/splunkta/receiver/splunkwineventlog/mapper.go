// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkwineventlog

import (
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// kind is the inputs.conf stanza kind this receiver owns. It is the only input
// kind whose canonical spelling is not lowercase, and kinds are matched exactly
// rather than case-folded, since UF would not honor a spelling its spec does not
// define.
const kind = "WinEventLog"

// MapInputs builds one config per [WinEventLog://...] stanza.
func mapInputs(c splunkconf.Conf) (map[string]component.Config, error) {
	out := map[string]component.Config{}
	for _, in := range c.Inputs {
		if in.Name.Kind != kind {
			continue
		}
		st := in.Configuration.Stanza
		attrs := splunkconf.StanzaAttrs(st)
		out[st.Name] = &Config{
			EventLogName: in.Name.Target,
			Index:        attrs.Index,
			Source:       attrs.Source,
			Sourcetype:   attrs.Sourcetype,
			Host:         attrs.Host,
			Extra:        attrs.Extra,
			Props:        in.Props,
			Transforms:   in.Transforms,
		}
	}
	return out, nil
}
