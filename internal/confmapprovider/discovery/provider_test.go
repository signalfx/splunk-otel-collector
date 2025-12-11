// Copyright  Splunk, Inc.
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

package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pipeline"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver"
)

func TestConfigDProviderHappyPath(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "splunk.configd", provider.ConfigDScheme())
	configD := provider.ConfigDProviderFactory().Create(confmap.ProviderSettings{})
	assert.Equal(t, "splunk.configd", configD.Scheme())

	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	conf, err := retrieved.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	require.NoError(t, configD.Shutdown(context.Background()))
}

func TestConfigDProviderDifferentConfigDirs(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	configD := provider.ConfigDProviderFactory().Create(confmap.ProviderSettings{})
	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err := retrieved.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	configDir = filepath.Join(".", "testdata", "another-config.d")
	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err = retrieved.AsRaw()
	require.NoError(t, err)
	anotherExpectedServiceConfig := map[string]any{
		"exporters": map[string]any{
			"signalfx": map[string]any{
				"api_url":    "http://0.0.0.0/different-api",
				"ingest_url": "http://0.0.0.0/different-ingest",
			},
		},
		"extensions": map[string]any{},
		"processors": map[string]any{},
		"receivers":  map[string]any{},
		"service":    map[string]any{},
	}
	assert.Equal(t, anotherExpectedServiceConfig, conf)
}

func TestConfigDProviderInvalidURIs(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	configD := provider.ConfigDProviderFactory().Create(confmap.ProviderSettings{})
	require.NotNil(t, configD)
	retrieved, err := configD.Retrieve(context.Background(), "not.a.thing:not.a.path", nil)
	require.EqualError(t, err, `uri "not.a.thing:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)

	retrieved, err = configD.Retrieve(context.Background(), discoveryModeScheme+":not.a.path", nil)
	require.EqualError(t, err, `uri "splunk.discovery:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)
}

func TestDiscoveryProvider_ContinuousDiscoveryConfig(t *testing.T) {
	t.Setenv("SPLUNK_INGEST_URL", "https://ingest.fake-realm.signalfx.com")
	t.Setenv("SPLUNK_ACCESS_TOKEN", "fake-token")

	confmapProvider, err := New()
	require.NoError(t, err)

	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{
				"file:" + filepath.Join("testdata", "base-config.yaml"),
				discoveryModeScheme + ":" + filepath.Join("testdata", "config.d"),
			},
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				confmapProvider.DiscoveryModeProviderFactory(),
				envprovider.NewFactory(),
			},
			ConverterFactories: []confmap.ConverterFactory{configconverter.ConverterFactoryFromFunc(configconverter.SetupDiscovery)},
			DefaultScheme:      "env",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, provider)

	factories, err := components.Get()
	require.NoError(t, err)

	conf, err := provider.Get(context.Background(), factories)
	require.NoError(t, err)
	assert.NotNil(t, conf)

	assert.Len(t, conf.Receivers, 1)
	drc, ok := conf.Receivers[component.MustNewIDWithName("discovery", "host_observer")].(*discoveryreceiver.Config)
	require.True(t, ok)
	assert.NotNil(t, drc)

	assert.Len(t, conf.Exporters, 1)
	oec, ok := conf.Exporters[component.MustNewIDWithName("otlphttp", "entities")].(*otlphttpexporter.Config)
	require.True(t, ok)
	expectedOtlpExporterConfig := otlphttpexporter.NewFactory().CreateDefaultConfig().(*otlphttpexporter.Config)
	expectedOtlpExporterConfig.LogsEndpoint = "https://ingest.fake-realm.signalfx.com/v3/event"
	expectedOtlpExporterConfig.ClientConfig.Headers = configopaque.MapList{
		{Name: "X-SF-Token", Value: "fake-token"},
	}
	assert.Equal(t, expectedOtlpExporterConfig, oec)

	assert.Len(t, conf.Extensions, 1)
	hoc, ok := conf.Extensions[component.MustNewID("host_observer")].(*hostobserver.Config)
	require.True(t, ok)
	assert.Equal(t, hostobserver.NewFactory().CreateDefaultConfig(), hoc)

	require.Len(t, conf.Service.Extensions, 1)
	assert.Equal(t, component.MustNewID("host_observer"), conf.Service.Extensions[0])

	pipelines := conf.Service.Pipelines
	assert.Len(t, pipelines, 2)

	metricsReceivers := pipelines[pipeline.NewID(pipeline.SignalMetrics)].Receivers
	require.Len(t, metricsReceivers, 1)
	assert.Equal(t, component.MustNewIDWithName("discovery", "host_observer"), metricsReceivers[0])

	logsReceivers := pipelines[pipeline.NewIDWithName(pipeline.SignalLogs, "entities")].Receivers
	require.Len(t, logsReceivers, 1)
	assert.Equal(t, component.MustNewIDWithName("discovery", "host_observer"), logsReceivers[0])

	logsExporters := pipelines[pipeline.NewIDWithName(pipeline.SignalLogs, "entities")].Exporters
	require.Len(t, logsExporters, 1)
	assert.Equal(t, component.MustNewIDWithName("otlphttp", "entities"), logsExporters[0])
}

func TestDiscoveryProvider_HostObserverDisabled(t *testing.T) {
	confmapProvider, err := New()
	require.NoError(t, err)

	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{
				"file:" + filepath.Join("testdata", "base-config.yaml"),
				propertiesFileScheme + ":" + filepath.Join("testdata", "disable-host-observer.properties.yaml"),
				discoveryModeScheme + ":" + filepath.Join("testdata", "config.d"),
			},
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				confmapProvider.DiscoveryModeProviderFactory(),
				envprovider.NewFactory(),
				confmapProvider.PropertiesFileProviderFactory(),
			},
			ConverterFactories: []confmap.ConverterFactory{configconverter.ConverterFactoryFromFunc(configconverter.SetupDiscovery)},
			DefaultScheme:      "env",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, provider)

	factories, err := components.Get()
	require.NoError(t, err)

	conf, err := provider.Get(context.Background(), factories)
	require.NoError(t, err)
	assert.NotNil(t, conf)

	// The only functional host observer must be disabled in the provided properties file.
	assert.Empty(t, conf.Extensions)

	// Discovery receivers should not be created if no discovery observers available.
	assert.Empty(t, conf.Receivers)
}
