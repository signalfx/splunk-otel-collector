// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkscript

import (
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// kind is the inputs.conf stanza kind this receiver owns. An unprefixed stanza
// declares no input, so it is deliberately not claimed.
const kind = "script"

// MapInputs builds one config per [script://...] stanza.
//
// AppDir carries the app that declared the stanza so a relative script target
// resolves against it rather than the collector's working directory.
func mapInputs(c splunkconf.Conf) (map[string]component.Config, error) {
	out := map[string]component.Config{}
	for _, in := range c.Inputs {
		if in.Name.Kind != kind {
			continue
		}
		st := in.Configuration.Stanza
		attrs := splunkconf.StanzaAttrs(st, "interval")
		cfg := &Config{
			ScriptFilename: in.Name.Target,
			AppDir:         in.AppDir,
			Index:          attrs.Index,
			Source:         attrs.Source,
			Sourcetype:     attrs.Sourcetype,
			Host:           attrs.Host,
			Extra:          attrs.Extra,
			Props:          in.Props,
			Transforms:     in.Transforms,
		}
		if p := st.Params.Get("interval"); p != nil {
			cfg.Interval = p.Value
		}
		out[st.Name] = cfg
	}
	return out, nil
}
