// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// params flattens a stanza's params for comparison.
func params(t *testing.T, inputs []Input, name string) map[string]string {
	t.Helper()
	for i := range inputs {
		if inputs[i].Configuration.Stanza.Name != name {
			continue
		}
		got := map[string]string{}
		for _, p := range inputs[i].Configuration.Stanza.Params {
			got[p.Name] = p.Value
		}
		return got
	}
	t.Fatalf("stanza %q not found in %v", name, names(inputs))
	return nil
}

func names(inputs []Input) []string {
	out := make([]string, 0, len(inputs))
	for i := range inputs {
		out = append(out, inputs[i].Configuration.Stanza.Name)
	}
	return out
}

// Settings written outside any stanza are global, per inputs.conf.spec: "You
// can also define global settings outside of any stanza, at the top of the
// file."
func TestReadInputKeepsPreStanzaGlobals(t *testing.T) {
	got, err := ReadInput([]byte("index = global_idx\n\n[monitor:///var/log/a]\nsourcetype = st\n"), "/app")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"index": "global_idx"}, params(t, got, defaultStanza))
}

// A literal [default] and the pre-stanza globals are the same global layer, and
// the spec gives the later definition in the file precedence.
func TestReadInputMergesPreStanzaGlobalsIntoDefault(t *testing.T) {
	got, err := ReadInput([]byte("index = from_top\nhost = from_top\n\n[default]\nindex = from_default\n"), "/app")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"index": "from_default", "host": "from_top"}, params(t, got, defaultStanza))
}

// splunkd merges [default] into every stanza, with the stanza's own value
// winning, and [default] itself declares no input.
func TestApplyGlobals(t *testing.T) {
	inputs := []Input{
		inputStanza("default", map[string]string{"index": "default_idx", "host": "default_host"}),
		inputStanza("monitor:///var/log/a", nil),
		inputStanza("monitor:///var/log/b", map[string]string{"index": "own_idx"}),
	}

	got, _ := ApplyGlobals(inputs)

	require.ElementsMatch(t,
		[]string{"monitor:///var/log/a", "monitor:///var/log/b"},
		names(got), "[default] is config, not an input, so it must not be returned")

	require.Equal(t, map[string]string{
		"index": "default_idx",
		"host":  "default_host",
	}, params(t, got, "monitor:///var/log/a"))

	require.Equal(t, map[string]string{
		"index": "own_idx", // the stanza's own value wins
		"host":  "default_host",
	}, params(t, got, "monitor:///var/log/b"))
}

// A stanza that is just a non-built-in scheme name is that scheme's defaults
// stanza, so splunkd consumes it even with no instances to configure and never
// runs it as an input.
func TestApplyGlobalsConsumesSchemeStanzaWithNoInstances(t *testing.T) {
	inputs := []Input{
		inputStanza("my_bare_script", map[string]string{"interval": "60"}),
		inputStanza("monitor:///var/log/a", nil),
	}
	require.Equal(t, []string{"monitor:///var/log/a"}, namesOf(ApplyGlobals(inputs)))
}

// A modular input takes its defaults from its own [<scheme>] stanza instead of
// [default], and only host and index come from [default]. This is how Splunk's
// shipped [WinEventLog] defaults reach [WinEventLog://...].
func TestApplyGlobalsSchemeDefaults(t *testing.T) {
	inputs := []Input{
		inputStanza("default", map[string]string{
			"index": "default_idx", "host": "default_host", "sourcetype": "default_st",
		}),
		inputStanza("WinEventLog", map[string]string{"interval": "60", "index": "scheme_idx"}),
		inputStanza("WinEventLog://Application", nil),
		inputStanza("WinEventLog://Security", map[string]string{"interval": "10"}),
	}

	got, schemeDefaults := ApplyGlobals(inputs)
	require.ElementsMatch(t,
		[]string{"WinEventLog://Application", "WinEventLog://Security"}, names(got))

	// The folded stanza comes back too: after the fold, Security's own interval
	// of 10 is indistinguishable from an inherited one, and a
	// use_single_instance scheme has to schedule from the 60 below instead.
	require.Contains(t, schemeDefaults, "WinEventLog")
	require.Equal(t, "60", schemeDefaults["WinEventLog"].Configuration.Stanza.Params.Get("interval").Value)

	require.Equal(t, map[string]string{
		"interval": "60",
		"index":    "scheme_idx",   // [<scheme>] wins over [default]
		"host":     "default_host", // layered back from [default]
	}, params(t, got, "WinEventLog://Application"),
		"sourcetype must not come from [default]: a modular input only inherits host and index")

	require.Equal(t, map[string]string{
		"interval": "10", // the instance's own value wins
		"index":    "scheme_idx",
		"host":     "default_host",
	}, params(t, got, "WinEventLog://Security"))
}

// A kindless stanza naming an input kind is not a second global layer: splunkd
// nominates one defaults stanza per read and for a plain conf read that is
// [default]. The global-stanzas parity case records UF ignoring [monitor].
func TestApplyGlobalsIgnoresKindStanza(t *testing.T) {
	inputs := []Input{
		inputStanza("default", map[string]string{"index": "default_idx"}),
		inputStanza("monitor", map[string]string{"sourcetype": "kind_st"}),
		inputStanza("monitor:///var/log/a", nil),
	}
	got, _ := ApplyGlobals(inputs)
	require.ElementsMatch(t, []string{"monitor", "monitor:///var/log/a"}, names(got))
	require.Equal(t, map[string]string{"index": "default_idx"}, params(t, got, "monitor:///var/log/a"))
	require.Equal(t, map[string]string{"sourcetype": "kind_st", "index": "default_idx"}, params(t, got, "monitor"))
}

// A name that does not parse is left alone rather than taken for a defaults
// stanza, so the name parse still reports it.
func TestApplyGlobalsKeepsUnparsableName(t *testing.T) {
	inputs := []Input{inputStanza("://no-kind", map[string]string{"index": "idx"})}
	require.Equal(t, []string{"://no-kind"}, namesOf(ApplyGlobals(inputs)))
}

func inputStanza(name string, kv map[string]string) Input {
	var ps Params
	for k, v := range kv {
		ps = append(ps, Param{Name: k, Value: v})
	}
	return Input{Configuration: Configuration{Stanza: Stanza{Name: name, Params: ps}}}
}

// namesOf adapts ApplyGlobals' two results for a test that only cares about the
// inputs it returned.
func namesOf(inputs []Input, _ map[string]Input) []string {
	return names(inputs)
}
