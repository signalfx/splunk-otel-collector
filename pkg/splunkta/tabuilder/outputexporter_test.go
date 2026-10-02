// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package tabuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func hecOutput(name string, params ...conf.Param) *conf.Output {
	return &conf.Output{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{Name: name, Params: params},
		},
	}
}

// TestCreateOutputExporterHECOut covers the stanza -> exporter path for the one
// output kind the TA runner builds in-process. It had no coverage before the
// [hecout] translation moved into the splunk_hecout factory.
func TestCreateOutputExporterHECOut(t *testing.T) {
	out := hecOutput("hecout",
		conf.Param{Name: "uri", Value: "https://hec.example.com:8088"},
		conf.Param{Name: "httpEventCollectorToken", Value: "tok"},
	)

	e, err := CreateOutputExporter(out, zap.NewNop(), component.TelemetrySettings{})
	require.NoError(t, err)
	require.NotNil(t, e)
}

// TestCreateOutputExporterUnsupportedKind pins the documented contract: an
// unsupported kind yields a nil exporter and no error, leaving the caller to log
// it. createReceivers-style callers skip on nil.
func TestCreateOutputExporterUnsupportedKind(t *testing.T) {
	e, err := CreateOutputExporter(hecOutput("tcpout:primary",
		conf.Param{Name: "server", Value: "idx1:9997"},
	), zap.NewNop(), component.TelemetrySettings{})
	require.NoError(t, err)
	require.Nil(t, e)
}

// TestCreateOutputExporterIncompleteStanza proves a [hecout] missing its token
// fails at build time rather than at first export.
func TestCreateOutputExporterIncompleteStanza(t *testing.T) {
	for _, tt := range []struct {
		name   string
		errMsg string
		params []conf.Param
	}{
		{
			name:   "no token",
			params: []conf.Param{{Name: "uri", Value: "https://hec:8088"}},
			errMsg: "token is required",
		},
		{
			name:   "no uri",
			params: []conf.Param{{Name: "httpEventCollectorToken", Value: "tok"}},
			errMsg: "endpoint is required",
		},
		{
			name: "malformed uri",
			params: []conf.Param{
				{Name: "uri", Value: "http://[::1"},
				{Name: "httpEventCollectorToken", Value: "tok"},
			},
			errMsg: `invalid "endpoint"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, err := CreateOutputExporter(hecOutput("hecout", tt.params...), zap.NewNop(), component.TelemetrySettings{})
			require.ErrorContains(t, err, tt.errMsg)
			require.Nil(t, e)
		})
	}
}
