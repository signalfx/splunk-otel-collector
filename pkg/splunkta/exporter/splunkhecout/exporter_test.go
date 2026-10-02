// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhecout

import (
	"context"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter/exportertest"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func output(params ...conf.Param) conf.Output {
	return conf.Output{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{Name: "hecout", Params: params},
		},
	}
}

// These two .conf key names are the whole contract with outputs.conf.
func TestConfigFromOutput(t *testing.T) {
	cfg := ConfigFromOutput(output(
		conf.Param{Name: "uri", Value: "https://hec.example.com:8088"},
		conf.Param{Name: "httpEventCollectorToken", Value: "tok"},
	))

	require.Equal(t, &Config{
		Endpoint: "https://hec.example.com:8088",
		Token:    "tok",
		TLS:      TLS{InsecureSkipVerify: true},
	}, cfg)
}

// Absent params must empty out, not panic, leaving Validate to report it.
func TestConfigFromOutputMissingParams(t *testing.T) {
	cfg := ConfigFromOutput(output())
	require.Empty(t, cfg.Endpoint)
	require.Empty(t, cfg.Token)
	require.ErrorContains(t, cfg.Validate(), "endpoint is required")
}

func TestValidate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		errMsg string
		cfg    Config
	}{
		{
			name: "valid",
			cfg:  Config{Endpoint: "https://hec:8088", Token: "tok"},
		},
		{
			name:   "no endpoint",
			cfg:    Config{Token: "tok"},
			errMsg: "endpoint is required",
		},
		{
			name:   "no token",
			cfg:    Config{Endpoint: "https://hec:8088"},
			errMsg: "token is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.errMsg == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

// Unlike the receiver wrappers this config has no catch-all, so an unknown key
// must be rejected rather than silently dropped.
func TestConfigUnmarshal(t *testing.T) {
	cm := confmap.NewFromStringMap(map[string]any{
		"endpoint": "https://hec:8088",
		"token":    "tok",
		"tls":      map[string]any{"insecure_skip_verify": true},
	})
	cfg := NewFactory().CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(cfg))
	require.Equal(t, &Config{
		Endpoint: "https://hec:8088",
		Token:    "tok",
		TLS:      TLS{InsecureSkipVerify: true},
	}, cfg)

	unknown := confmap.NewFromStringMap(map[string]any{
		"endpoint":   "https://hec:8088",
		"nonsuch":    "x",
		"tls":        map[string]any{"insecure_skip_verify": true},
		"token":      "tok",
		"sourcetype": "st",
	})
	require.Error(t, unknown.Unmarshal(&Config{}), "strict unmarshal must reject unmodeled keys")
}

// Assert the values land on real splunkhec fields, not just that unmarshal ran.
func TestHECConfigTranslation(t *testing.T) {
	cfg := &Config{Endpoint: "https://hec:8088", Token: "tok", TLS: TLS{InsecureSkipVerify: true}}

	hecCfg, err := cfg.hecConfig(splunkhecexporter.NewFactory())
	require.NoError(t, err)

	typed, ok := hecCfg.(*splunkhecexporter.Config)
	require.True(t, ok)
	require.Equal(t, "https://hec:8088", typed.ClientConfig.Endpoint)
	require.Equal(t, "tok", string(typed.Token))
	require.True(t, typed.ClientConfig.TLS.InsecureSkipVerify)
}

// The wrapper's own Validate only checks for non-empty, so this pins that
// splunkhec's stricter endpoint check still runs.
func TestHECConfigRejectsBadEndpoint(t *testing.T) {
	cfg := &Config{Endpoint: "http://[::1", Token: "tok"}
	require.NoError(t, cfg.Validate(), "non-empty endpoint passes the wrapper's own check")

	_, err := cfg.hecConfig(splunkhecexporter.NewFactory())
	require.ErrorContains(t, err, "invalid splunkhec config")
	require.ErrorContains(t, err, `invalid "endpoint"`)
}

func TestFactory(t *testing.T) {
	f := NewFactory()
	cfg := &Config{Endpoint: "https://hec:8088", Token: "tok", TLS: TLS{InsecureSkipVerify: true}}
	e, err := f.CreateLogs(context.Background(), exportertest.NewNopSettings(f.Type()), cfg)
	require.NoError(t, err)
	require.NotNil(t, e)
}
