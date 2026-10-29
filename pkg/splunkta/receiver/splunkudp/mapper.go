// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkudp

import (
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

// kind is the inputs.conf stanza kind this receiver owns.
const kind = "udp"

// MapInputs builds one config per [udp://...] stanza.
func mapInputs(c splunkconf.Conf) (map[string]component.Config, error) {
	out := map[string]component.Config{}
	for _, in := range c.Inputs {
		if in.Name.Kind != kind {
			continue
		}
		st := in.Configuration.Stanza
		addr, port := stanza.ParseListenAddress(in.Name.Target)
		attrs := splunkconf.StanzaAttrs(st)
		out[st.Name] = &Config{
			ListenAddress: addr,
			Port:          port,
			Index:         attrs.Index,
			Source:        attrs.Source,
			Sourcetype:    attrs.Sourcetype,
			Host:          attrs.Host,
			Extra:         attrs.Extra,
			Props:         in.Props,
			Transforms:    in.Transforms,
		}
	}
	return out, nil
}
