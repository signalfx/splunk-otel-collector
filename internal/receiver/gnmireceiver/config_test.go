// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnmireceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	sub, err := cm.Sub("gnmi")
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	require.NoError(t, sub.Unmarshal(cfg))
	require.NoError(t, confmap.Validate(cfg))

	require.Len(t, cfg.Targets, 1)
	target := cfg.Targets[0]

	require.Equal(t, "10.0.0.1:57400", target.ClientConfig.Endpoint)
	require.Equal(t, "admin", string(target.Username))
	require.Equal(t, "admin", string(target.Password))
	require.Equal(t, encodingJSONIETF, target.Encoding) //nolint:testifylint // false positive: "json_ietf" is a gNMI encoding name, not JSON content
	require.Equal(t, 30*time.Second, target.Redial)
	require.True(t, target.ClientConfig.TLS.Insecure)
	require.Len(t, target.Subscriptions, 2)
	require.Equal(t, SubscriptionConfig{
		Path:           "/interfaces/interface/state/counters",
		Origin:         "openconfig",
		Mode:           modeSample,
		SampleInterval: 10 * time.Second,
		Default:        &MetricConfig{Type: metricTypeSum, Unit: "1"},
		Overrides: map[string]MetricConfig{
			"in-octets":  {Type: metricTypeSum, Unit: "By"},
			"out-octets": {Type: metricTypeSum, Unit: "By"},
		},
	}, target.Subscriptions[0])
	require.Equal(t, SubscriptionConfig{
		Path:              "/interfaces/interface/state/oper-status",
		Mode:              modeOnChange,
		HeartbeatInterval: 60 * time.Second,
		Overrides: map[string]MetricConfig{
			"oper-status": {Type: metricTypeGauge},
		},
	}, target.Subscriptions[1])
}

func TestValidate(t *testing.T) {
	t.Parallel()

	validTarget := func() TargetConfig {
		tc := NewDefaultTargetConfig()
		tc.ClientConfig.Endpoint = "10.0.0.1:57400"
		tc.Subscriptions = []SubscriptionConfig{
			{
				Path:           "/interfaces",
				Mode:           modeSample,
				SampleInterval: 10 * time.Second,
				Default:        &MetricConfig{Type: metricTypeSum, Unit: "By"},
			},
		}
		return tc
	}

	tests := []struct {
		name        string
		mutate      func(*Config)
		expectedErr string
	}{
		{
			name:   "valid",
			mutate: func(*Config) {},
		},
		{
			name:        "no targets",
			mutate:      func(c *Config) { c.Targets = nil },
			expectedErr: "at least one target",
		},
		{
			name:        "empty endpoint",
			mutate:      func(c *Config) { c.Targets[0].ClientConfig.Endpoint = "" },
			expectedErr: "endpoint is required",
		},
		{
			name:        "missing encoding",
			mutate:      func(c *Config) { c.Targets[0].Encoding = "" },
			expectedErr: "encoding is required",
		},
		{
			name:        "invalid encoding",
			mutate:      func(c *Config) { c.Targets[0].Encoding = "xml" },
			expectedErr: "invalid encoding",
		},
		{
			name:        "redial too small",
			mutate:      func(c *Config) { c.Targets[0].Redial = 500 * time.Millisecond },
			expectedErr: "redial must be at least 1s",
		},
		{
			name:   "redial zero is valid (disables reconnection)",
			mutate: func(c *Config) { c.Targets[0].Redial = 0 },
		},
		{
			name:        "no subscriptions",
			mutate:      func(c *Config) { c.Targets[0].Subscriptions = nil },
			expectedErr: "at least one subscription",
		},
		{
			name:        "empty path",
			mutate:      func(c *Config) { c.Targets[0].Subscriptions[0].Path = "" },
			expectedErr: "path is required",
		},
		{
			name:        "missing mode",
			mutate:      func(c *Config) { c.Targets[0].Subscriptions[0].Mode = "" },
			expectedErr: "mode is required",
		},
		{
			name:        "invalid mode",
			mutate:      func(c *Config) { c.Targets[0].Subscriptions[0].Mode = "poll" },
			expectedErr: "invalid mode",
		},
		{
			name: "sample without interval",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].SampleInterval = 0
			},
			expectedErr: "sample_interval must be > 0",
		},
		{
			name: "negative heartbeat_interval",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].HeartbeatInterval = -1 * time.Second
			},
			expectedErr: "heartbeat_interval must be >= 0",
		},
		{
			name: "no default and no overrides",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].Default = nil
				c.Targets[0].Subscriptions[0].Overrides = nil
			},
			expectedErr: "at least one of \"default\" or \"overrides\"",
		},
		{
			name: "invalid default type",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].Default = &MetricConfig{Type: "histogram"}
			},
			expectedErr: "invalid type",
		},
		{
			name: "override missing type",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].Overrides = map[string]MetricConfig{
					"in-octets": {Unit: "By"},
				}
			},
			expectedErr: "type is required",
		},
		{
			name: "override only, no default is valid",
			mutate: func(c *Config) {
				c.Targets[0].Subscriptions[0].Default = nil
				c.Targets[0].Subscriptions[0].Overrides = map[string]MetricConfig{
					"in-octets": {Type: metricTypeSum, Unit: "By"},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Targets: []TargetConfig{validTarget()}}
			tt.mutate(cfg)
			err := confmap.Validate(cfg)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}
