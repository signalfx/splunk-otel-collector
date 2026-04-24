// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package configsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestNewTelemetryHook(t *testing.T) {
	hook := NewTelemetryHook()
	require.NotNil(t, hook)
	assert.NotNil(t, hook.usedSources)
}

func TestTelemetryHook_OnRetrieve(t *testing.T) {
	tests := []struct {
		retrieved       map[string]any
		expectedSources map[string]bool
		name            string
	}{
		{
			name: "with named vault config source (vault/prod)",
			retrieved: map[string]any{
				"config_sources": map[string]any{
					"vault/prod": map[string]any{
						"endpoint": "http://vault-prod:8200",
					},
				},
			},
			expectedSources: map[string]bool{
				"vault": true,
			},
		},
		{
			name: "with multiple named config sources",
			retrieved: map[string]any{
				"config_sources": map[string]any{
					"vault/prod":    map[string]any{},
					"vault/staging": map[string]any{},
					"env/test":      map[string]any{},
					"include/base":  map[string]any{},
				},
			},
			expectedSources: map[string]bool{
				"vault":   true,
				"env":     true,
				"include": true,
			},
		},
		{
			name: "with vault config source",
			retrieved: map[string]any{
				"config_sources": map[string]any{
					"vault": map[string]any{
						"endpoint": "http://localhost:8200",
					},
				},
			},
			expectedSources: map[string]bool{
				"vault": true,
			},
		},
		{
			name: "with multiple config sources",
			retrieved: map[string]any{
				"config_sources": map[string]any{
					"vault": map[string]any{
						"endpoint": "http://localhost:8200",
					},
					"etcd2": map[string]any{
						"endpoints": []string{"http://localhost:2379"},
					},
					"include": map[string]any{},
				},
			},
			expectedSources: map[string]bool{
				"vault":   true,
				"etcd2":   true,
				"include": true,
			},
		},
		{
			name: "no config sources",
			retrieved: map[string]any{
				"receivers": map[string]any{
					"otlp": map[string]any{},
				},
			},
			expectedSources: map[string]bool{},
		},
		{
			name: "with non-custom config source",
			retrieved: map[string]any{
				"config_sources": map[string]any{
					"file": map[string]any{
						"path": "/etc/config",
					},
				},
			},
			expectedSources: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := NewTelemetryHook()
			hook.OnRetrieve("file", tt.retrieved)

			usedSources := hook.GetUsedSources()
			for source, expected := range tt.expectedSources {
				assert.Equal(t, expected, usedSources[source], "source: %s", source)
			}

			// Verify sources not in expectedSources are false
			if len(tt.expectedSources) == 0 {
				for _, source := range usedSources {
					assert.False(t, source)
				}
			}
		})
	}
}

func TestTelemetryHook_Metrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() {
		require.NoError(t, provider.Shutdown(t.Context()))
	}()

	hook := NewTelemetryHook()
	hook.SetTelemetrySettings(component.TelemetrySettings{MeterProvider: provider})

	// Simulate vault and include config sources being used
	hook.OnRetrieve("file", map[string]any{
		"config_sources": map[string]any{
			"vault": map[string]any{
				"endpoint": "http://localhost:8200",
			},
			"include": map[string]any{},
		},
	})

	var rm metricdata.ResourceMetrics
	err := reader.Collect(t.Context(), &rm)
	require.NoError(t, err)

	require.NotEmpty(t, rm.ScopeMetrics)

	var found bool
	for _, sm := range rm.ScopeMetrics {
		assert.Equal(t, meterName, sm.Scope.Name)
		for _, m := range sm.Metrics {
			if m.Name != configSourceUsageMetricName {
				continue
			}
			found = true
			assert.Equal(t, "Indicates whether a custom config source is in use (1 = in use)", m.Description)
			assert.Equal(t, "{usage}", m.Unit)

			gauge, ok := m.Data.(metricdata.Gauge[int64])
			require.True(t, ok)

			// Collect reported config source types from datapoints
			reported := make(map[string]int64)
			for _, dp := range gauge.DataPoints {
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == configSourceTypeAttributeKey {
						reported[attr.Value.AsString()] = dp.Value
					}
				}
			}

			assert.Equal(t, int64(1), reported["vault"], "vault should be reported as in use")
			assert.Equal(t, int64(1), reported["include"], "include should be reported as in use")

			assert.NotContains(t, reported, "zookeeper", "zookeeper should not have a datapoint")
			assert.NotContains(t, reported, "etcd2", "etcd2 should not have a datapoint")
			assert.NotContains(t, reported, "env", "env should not have a datapoint")

			// Exactly two datapoints expected
			assert.Len(t, gauge.DataPoints, 2, "only active config sources should have datapoints")
		}
	}
	assert.True(t, found, "config source usage metric should be found")
}

func TestTelemetryHook_OnShutdown(t *testing.T) {
	hook := NewTelemetryHook()

	hook.OnRetrieve("file", map[string]any{
		"config_sources": map[string]any{
			"vault": map[string]any{},
		},
	})

	usedSources := hook.GetUsedSources()
	assert.True(t, usedSources["vault"])

	hook.OnShutdown()

	usedSources = hook.GetUsedSources()
	assert.Empty(t, usedSources)
}

func TestCustomConfigSources_MatchesFactories(t *testing.T) {
	expected := make(map[string]struct{}, len(configSourceFactories))
	for typ := range configSourceFactories {
		expected[typ.String()] = struct{}{}
	}

	assert.Equal(t, expected, customConfigSources,
		"customConfigSources must match configSourceFactories exactly; add the new factory type to both")
}

func TestIsCustomConfigSource(t *testing.T) {
	tests := []struct {
		name     string
		csType   string
		expected bool
	}{
		{"env is custom", "env", true},
		{"include is custom", "include", true},
		{"vault is custom", "vault", true},
		{"zookeeper is custom", "zookeeper", true},
		{"etcd2 is custom", "etcd2", true},
		{"file is not custom", "file", false},
		{"http is not custom", "http", false},
		{"unknown is not custom", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isCustomConfigSource(tt.csType))
		})
	}
}

func TestTelemetryHook_LazyRegistration(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() {
		require.NoError(t, provider.Shutdown(t.Context()))
	}()

	hook := NewTelemetryHook()

	// No TelemetrySettings injected yet — collecting should yield nothing
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))
	assert.Empty(t, rm.ScopeMetrics, "no metrics expected before TelemetrySettings is injected")

	// Inject TelemetrySettings (as the extension does on Start)
	hook.SetTelemetrySettings(component.TelemetrySettings{MeterProvider: provider})

	hook.OnRetrieve("file", map[string]any{
		"config_sources": map[string]any{
			"vault": map[string]any{},
		},
	})

	rm = metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(t.Context(), &rm))
	require.NotEmpty(t, rm.ScopeMetrics, "metric should be present after TelemetrySettings is injected and a config source is tracked")
}
