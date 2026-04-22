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

package configsourcetelemetryextension

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	require.NotNil(t, factory)
	assert.Equal(t, component.MustNewType(TypeStr), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	require.NotNil(t, cfg)
	assert.IsType(t, &Config{}, cfg)
}

func TestCreateExtension(t *testing.T) {
	factory := NewFactory()

	cfg := factory.CreateDefaultConfig()
	settings := extensiontest.NewNopSettings(component.MustNewType(TypeStr))

	ext, err := factory.Create(context.Background(), settings, cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
}

func TestExtension_Start(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() {
		require.NoError(t, provider.Shutdown(context.Background()))
	}()

	// Create the hook (which registers itself as global); restore global on exit.
	hook := configsource.NewTelemetryHook()
	defer configsource.SetGlobalHook(nil)

	assert.Nil(t, hook.GetTelemetrySettings())

	factory := NewFactory()
	settings := extensiontest.NewNopSettings(component.MustNewType(TypeStr))
	settings.TelemetrySettings.MeterProvider = provider

	ext, err := factory.Create(context.Background(), settings, factory.CreateDefaultConfig())
	require.NoError(t, err)

	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	telSettings := hook.GetTelemetrySettings()
	require.NotNil(t, telSettings)
	assert.Equal(t, provider, telSettings.MeterProvider)

	require.NoError(t, ext.Shutdown(context.Background()))
}

func TestExtension_StartWithNilHook(t *testing.T) {
	configsource.SetGlobalHook(nil)

	factory := NewFactory()
	settings := extensiontest.NewNopSettings(component.MustNewType(TypeStr))

	ext, err := factory.Create(context.Background(), settings, factory.CreateDefaultConfig())
	require.NoError(t, err)

	require.NoError(t, ext.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, ext.Shutdown(context.Background()))
}

func TestExtension_IntegrationWithMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() {
		require.NoError(t, provider.Shutdown(context.Background()))
	}()

	// Create the hook (registers as global); restore global on exit.
	hook := configsource.NewTelemetryHook()
	defer configsource.SetGlobalHook(nil)

	// Simulate config sources being used before the extension starts —
	// this mirrors the real-world flow where config is loaded before service startup.
	hook.OnRetrieve("file", map[string]any{
		"config_sources": map[string]any{
			"vault": map[string]any{
				"endpoint": "http://localhost:8200",
			},
		},
	})

	factory := NewFactory()
	settings := extensiontest.NewNopSettings(component.MustNewType(TypeStr))
	settings.TelemetrySettings.MeterProvider = provider

	ext, err := factory.Create(context.Background(), settings, factory.CreateDefaultConfig())
	require.NoError(t, err)

	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	found := false
outer:
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "otelcol_splunk_configsource_usage" {
				continue
			}
			found = true
			gauge, ok := m.Data.(metricdata.Gauge[int64])
			require.True(t, ok)

			reported := make(map[string]int64)
			for _, dp := range gauge.DataPoints {
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == "config_source_type" {
						reported[attr.Value.AsString()] = dp.Value
					}
				}
			}

			assert.Equal(t, int64(1), reported["vault"], "vault should be reported as in use")
			assert.Len(t, gauge.DataPoints, 1, "only active config sources should have datapoints")
			break outer
		}
	}
	assert.True(t, found, "otelcol_splunk_configsource_usage metric should be present")

	require.NoError(t, ext.Shutdown(context.Background()))
}
