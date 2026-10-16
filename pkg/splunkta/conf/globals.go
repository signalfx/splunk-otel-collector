// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import "github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"

// defaultStanza is the inputs.conf stanza holding settings that apply to every
// other stanza in the file, including the ones written above the first stanza
// header.
const defaultStanza = "default"

// schemeGlobalKeys are the only [default] settings a modular input inherits.
// Its own [<scheme>] stanza replaces [default] rather than layering under it, so
// splunkd puts these two back by hand
// (PropertyPages::propsByInputSchemeAndMode: "we want modular inputs to default
// to these global settings").
var schemeGlobalKeys = []string{"host", "index"}

// ApplyGlobals folds inputs.conf's global stanzas into the stanzas they
// configure and drops them from the result, since they declare no input of
// their own. A stanza's own value always wins.
//
// splunkd nominates one defaults stanza per read, and which one depends on
// whether the scheme is built in (PropertyPages::propsObjByFilterAndMode and
// inputsPropsByPrefixAndMode):
//
//   - A built-in scheme ([monitor://...], [tcp://...]) gets [default]. A bare
//     [monitor] is just another stanza and configures nothing, which the
//     global-stanzas parity case records UF doing.
//   - Any other scheme is assumed to be a modular input, so [<scheme>] is its
//     defaults stanza instead of [default], and only schemeGlobalKeys come from
//     [default]. This is how Splunk's own [WinEventLog] and [perfmon] defaults
//     reach their instances.
//
// A modular input's [<scheme>] stanza is consumed either way, even with no
// instances to configure, so an unprefixed stanza never comes back as an input.
//
// Call this after merging the conf layers, not per file. Stanza specificity
// outranks layer precedence, so a [default] in a higher layer must not override
// a value the instance sets in a lower one, and resolving each stanza across
// layers first is what gives that.
func ApplyGlobals(inputs []Input) []Input {
	var (
		global      Input
		hasGlobal   bool
		ordinary    []Input
		modInstance []Input
		schemeOf    = map[string]Input{}
	)
	for _, in := range inputs {
		name := in.Configuration.Stanza.Name
		switch {
		case name == defaultStanza:
			global, hasGlobal = in, true
		case !stanza.IsModularInput(name):
			ordinary = append(ordinary, in)
		default:
			switch parsed, err := stanza.ParseName(name); {
			case err != nil:
				// Not ours to adjudicate: leave it so the name parse reports it
				// instead of it vanishing as a defaults stanza.
				ordinary = append(ordinary, in)
			case parsed.Kind == "":
				// The scheme on its own, so it is that scheme's defaults stanza.
				schemeOf[name] = in
			default:
				modInstance = append(modInstance, in)
			}
		}
	}

	out := make([]Input, 0, len(ordinary)+len(modInstance))
	for _, in := range ordinary {
		if hasGlobal {
			in = layer(global, in)
		}
		out = append(out, in)
	}
	for _, in := range modInstance {
		parsed, _ := stanza.ParseName(in.Configuration.Stanza.Name)
		if scheme, ok := schemeOf[parsed.Kind]; ok {
			in = layer(scheme, in)
		}
		if hasGlobal {
			in = layerKeys(global, in, schemeGlobalKeys)
		}
		out = append(out, in)
	}
	return out
}

// layer merges base beneath in, keeping in's identity. mergeInput returns the
// base layer's own, so it has to be put back.
func layer(base, in Input) Input {
	merged := mergeInput(base, in)
	merged.AppDir = in.AppDir
	merged.Configuration.Stanza.Name = in.Configuration.Stanza.Name
	return merged
}

// layerKeys merges only the named settings from base, and only where in does
// not set them already.
func layerKeys(base, in Input, keys []string) Input {
	for _, key := range keys {
		if in.Configuration.Stanza.Params.Get(key) != nil {
			continue
		}
		if p := base.Configuration.Stanza.Params.Get(key); p != nil {
			in.Configuration.Stanza.Params = append(in.Configuration.Stanza.Params, *p)
		}
	}
	return in
}
