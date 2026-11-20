// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkmonitor_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/stanza"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkmonitor"
)

func input(t *testing.T, name string, params conf.Params) splunkconf.Input {
	t.Helper()
	parsed, err := stanza.ParseName(name)
	require.NoError(t, err)
	return splunkconf.Input{
		Input: conf.Input{
			Configuration: conf.Configuration{
				Stanza: conf.Stanza{Name: name, Params: params},
			},
		},
		Name: parsed,
	}
}

// The three params that bind to typed fields rather than riding through Extra.
// TRUNCATE becomes an int, so it is the one that would break strict unmarshal if
// it were carried as a string.
func TestMapInputsTypedParams(t *testing.T) {
	got, err := splunkmonitor.NewFactory().MapInputs(splunkconf.Conf{
		Inputs: []splunkconf.Input{input(t, "monitor:///var/log/typed", conf.Params{
			{Name: "CHARSET", Value: "UTF-8"},
			{Name: "EVENT_BREAKER", Value: "BRK"},
			{Name: "TRUNCATE", Value: "1024"},
			{Name: "NOT_A_NUMBER", Value: "x"},
		})},
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, &splunkmonitor.Config{
		Path:         "/var/log/typed",
		EventBreaker: "BRK",
		Truncate:     1024,
		Extra:        map[string]string{"NOT_A_NUMBER": "x"},
	}, got["monitor:///var/log/typed"])
}

// A non-numeric TRUNCATE leaves the field zero, which means no limit, rather
// than failing the whole resolve over one stanza param.
func TestMapInputsUnparsableTruncate(t *testing.T) {
	got, err := splunkmonitor.NewFactory().MapInputs(splunkconf.Conf{
		Inputs: []splunkconf.Input{input(t, "monitor:///var/log/t", conf.Params{
			{Name: "TRUNCATE", Value: "lots"},
		})},
	})
	require.NoError(t, err)
	require.Equal(t, 0, got["monitor:///var/log/t"].(*splunkmonitor.Config).Truncate)
}

// A mapper only claims its own kind, which is what lets the provider report a
// stanza no mapper owns instead of silently dropping it.
func TestMapInputsIgnoresOtherKinds(t *testing.T) {
	got, err := splunkmonitor.NewFactory().MapInputs(splunkconf.Conf{
		Inputs: []splunkconf.Input{
			input(t, "tcp://5514", nil),
			input(t, "batch:///var/lib/b.txt", nil),
		},
	})
	require.NoError(t, err)
	require.Empty(t, got)
}
