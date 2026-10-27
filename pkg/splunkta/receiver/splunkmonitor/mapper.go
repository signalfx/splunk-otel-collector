// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkmonitor

import (
	"strconv"

	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// kind is the inputs.conf stanza kind this receiver owns. Kinds are matched
// against the spelling inputs.conf.spec defines, exactly rather than
// case-folded, since UF would not honor a spelling its spec does not define.
const kind = "monitor"

// MapInputs builds one config per [monitor://...] stanza.
//
// whitelist and blacklist are deliberately not modeled: the underlying receiver
// reads both straight off the stanza params, so they round-trip verbatim through
// Extra instead of being re-encoded here.
func (Factory) MapInputs(c splunkconf.Conf) (map[string]component.Config, error) {
	out := map[string]component.Config{}
	for _, in := range c.Inputs {
		if in.Name.Kind != kind {
			continue
		}
		st := in.Configuration.Stanza
		attrs := splunkconf.StanzaAttrs(st, "CHARSET", "EVENT_BREAKER", "TRUNCATE")
		cfg := &Config{
			Path:       in.Name.Target,
			Index:      attrs.Index,
			Source:     attrs.Source,
			Sourcetype: attrs.Sourcetype,
			Host:       attrs.Host,
			Extra:      attrs.Extra,
			Props:      in.Props,
			Transforms: in.Transforms,
		}
		if p := st.Params.Get("EVENT_BREAKER"); p != nil {
			cfg.EventBreaker = p.Value
		}
		// A non-numeric TRUNCATE leaves the field zero, which means no limit,
		// matching what the stanza would do with an unparsable value.
		if p := st.Params.Get("TRUNCATE"); p != nil {
			if n, err := strconv.Atoi(p.Value); err == nil {
				cfg.Truncate = n
			}
		}
		out[st.Name] = cfg
	}
	return out, nil
}
