// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhome

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// s2sOutputMapper is a stand-in for the mapper factory the private data-runtimes
// flavor would register: [tcpout] -> splunk_s2sout exporter config. It shows the
// shape a client mapper takes (Scheme + MapOutput) without depending on the
// proprietary code.
type s2sOutputMapper struct{}

func (s2sOutputMapper) Scheme() string { return "tcpout" }

func (s2sOutputMapper) MapOutput(out conf.Output) (Component, error) {
	st := out.Configuration.Stanza
	get := func(k string) string {
		if p := st.Params.Get(k); p != nil {
			return p.Value
		}
		return ""
	}
	return Component{
		ID: "splunk_s2sout/" + StableName(st.Name),
		Cfg: map[string]any{
			"endpoint": get("server"),
		},
	}, nil
}

// TestWithOutputMapper proves a client can register an additional output kind
// ([tcpout]) and have the provider emit its exporter alongside the built-in
// [hecout], with no change to this repo's mapping code.
func TestWithOutputMapper(t *testing.T) {
	reg := newRegistry()
	WithOutputMapper(s2sOutputMapper{})(reg)

	merged := conf.Map{
		"hecout":         {"uri": "https://hec:8088", "httpEventCollectorToken": "tok"},
		"tcpout:primary": {"server": "idx1:9997"},
	}

	comps, err := mapOutputs(reg.outputs, merged)
	require.NoError(t, err)

	byType := map[string]map[string]any{}
	for _, c := range comps {
		typ, _, _ := splitID(c.ID)
		byType[typ] = c.Cfg
	}
	require.Contains(t, byType, "splunk_hecout")
	require.Contains(t, byType, "splunk_s2sout")
	require.Equal(t, "idx1:9997", byType["splunk_s2sout"]["endpoint"])
}

// TestUnregisteredOutputKind proves an output kind with no registered mapper
// fails with a clear error that lists the supported kinds, rather than being
// silently dropped.
func TestUnregisteredOutputKind(t *testing.T) {
	reg := newRegistry() // built-ins only: hecout
	merged := conf.Map{"tcpout:primary": {"server": "idx1:9997"}}

	_, err := mapOutputs(reg.outputs, merged)
	require.ErrorContains(t, err, "unsupported output stanza")
	require.ErrorContains(t, err, "tcpout")
	require.ErrorContains(t, err, "supported kinds: hecout")
}

// splitID splits a "type/name" component ID into its type and name.
func splitID(id string) (typ, name string, ok bool) {
	for i := 0; i < len(id); i++ {
		if id[i] == '/' {
			return id[:i], id[i+1:], true
		}
	}
	return id, "", false
}
