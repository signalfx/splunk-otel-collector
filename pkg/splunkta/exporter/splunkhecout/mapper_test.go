// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhecout_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/exporter/splunkhecout"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

func output(t *testing.T, name string, params ...conf.Param) splunkconf.Output {
	t.Helper()
	parsed, err := stanza.ParseOutputName(name)
	require.NoError(t, err)
	return splunkconf.Output{
		Name: parsed,
		Output: conf.Output{Configuration: conf.Configuration{Stanza: conf.Stanza{
			Name:   name,
			Params: params,
		}}},
	}
}

// The mapper claims [hecout] and nothing else, so a kind it does not own stays
// available for another registered mapper.
func TestMapOutputs(t *testing.T) {
	outputs := []splunkconf.Output{
		output(t, "hecout",
			conf.Param{Name: "uri", Value: "https://hec:8088"},
			conf.Param{Name: "httpEventCollectorToken", Value: "tok"},
		),
		output(t, "hecout:second", conf.Param{Name: "uri", Value: "https://other:8088"}),
		output(t, "tcpout:primary", conf.Param{Name: "server", Value: "idx1:9997"}),
	}

	got, err := splunkhecout.NewFactory().MapOutputs(outputs)
	require.NoError(t, err)

	require.Equal(t, map[string]component.Config{
		"hecout": &splunkhecout.Config{
			Endpoint: "https://hec:8088",
			Token:    "tok",
			TLS:      splunkhecout.TLS{InsecureSkipVerify: true},
		},
		"hecout:second": &splunkhecout.Config{
			Endpoint: "https://other:8088",
			TLS:      splunkhecout.TLS{InsecureSkipVerify: true},
		},
	}, got)
}

// sslVerifyServerCert inverts into insecure_skip_verify, and defaults to false.
func TestSSLVerifyServerCert(t *testing.T) {
	for _, tt := range []struct {
		value              string
		insecureSkipVerify bool
	}{
		{value: "", insecureSkipVerify: true},
		{value: "true", insecureSkipVerify: false},
		{value: "false", insecureSkipVerify: true},
	} {
		t.Run("sslVerifyServerCert="+tt.value, func(t *testing.T) {
			params := []conf.Param{{Name: "uri", Value: "https://hec:8088"}}
			if tt.value != "" {
				params = append(params, conf.Param{Name: "sslVerifyServerCert", Value: tt.value})
			}
			out := output(t, "hecout", params...)
			cfg := splunkhecout.ConfigFromOutput(out.Output)
			require.Equal(t, tt.insecureSkipVerify, cfg.TLS.InsecureSkipVerify)
		})
	}
}

// A stanza with no uri must fail config validation rather than build an exporter
// pointed at nothing.
func TestConfigValidate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		errMsg string
		cfg    splunkhecout.Config
	}{
		{
			name: "complete",
			cfg:  splunkhecout.Config{Endpoint: "https://hec:8088", Token: "tok"},
		},
		{
			name:   "missing endpoint",
			cfg:    splunkhecout.Config{Token: "tok"},
			errMsg: "endpoint is required",
		},
		{
			name:   "missing token",
			cfg:    splunkhecout.Config{Endpoint: "https://hec:8088"},
			errMsg: "token is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
