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
	"go.opentelemetry.io/collector/featuregate"
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
	assert.NoError(t, err)
	require.NotNil(t, retrieved)

	conf, err := retrieved.AsRaw()
	assert.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	assert.NoError(t, configD.Shutdown(context.Background()))
}

func TestConfigDProviderDifferentConfigDirs(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	configD := provider.ConfigDProviderFactory().Create(confmap.ProviderSettings{})
	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err := retrieved.AsRaw()
	assert.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	configDir = filepath.Join(".", "testdata", "another-config.d")
	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err = retrieved.AsRaw()
	assert.NoError(t, err)
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
	assert.EqualError(t, err, `uri "not.a.thing:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)

	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:not.a.path", discoveryModeScheme), nil)
	assert.EqualError(t, err, `uri "splunk.discovery:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)
}

func TestDiscoveryProvider_ContinuousDiscoveryConfig(t *testing.T) {
	require.NoError(t, featuregate.GlobalRegistry().Set(continuousDiscoveryFGKey, true))
	t.Setenv("SPLUNK_INGEST_URL", "https://ingest.fake-realm.signalfx.com")
	t.Setenv("SPLUNK_ACCESS_TOKEN", "fake-token")

	confmapProvider, err := New()
	require.NoError(t, err)

	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{
				fmt.Sprintf("file:%s", filepath.Join("testdata", "base-config.yaml")),
				fmt.Sprintf("%s:%s", discoveryModeScheme, filepath.Join("testdata", "config.d")),
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

	assert.Equal(t, 1, len(conf.Receivers))
	drc, ok := conf.Receivers[component.MustNewIDWithName("discovery", "host_observer")].(*discoveryreceiver.Config)
	require.True(t, ok)
	assert.NotNil(t, drc)

	assert.Equal(t, 1, len(conf.Exporters))
	oec, ok := conf.Exporters[component.MustNewIDWithName("otlphttp", "entities")].(*otlphttpexporter.Config)
	require.True(t, ok)
	expectedOtlpExporterConfig := otlphttpexporter.NewFactory().CreateDefaultConfig().(*otlphttpexporter.Config)
	expectedOtlpExporterConfig.LogsEndpoint = "https://ingest.fake-realm.signalfx.com/v3/event"
	expectedOtlpExporterConfig.ClientConfig.Headers = map[string]configopaque.String{
		"X-SF-Token": "fake-token",
	}
	assert.Equal(t, expectedOtlpExporterConfig, oec)

	assert.Equal(t, 1, len(conf.Extensions))
	hoc, ok := conf.Extensions[component.MustNewID("host_observer")].(*hostobserver.Config)
	require.True(t, ok)
	assert.Equal(t, hostobserver.NewFactory().CreateDefaultConfig(), hoc)

	assert.EqualValues(t, []component.ID{component.MustNewID("host_observer")}, conf.Service.Extensions)
	pipelines := conf.Service.Pipelines
	assert.Equal(t, 2, len(pipelines))
	assert.Equal(t, []component.ID{component.MustNewIDWithName("discovery", "host_observer")},
		pipelines[pipeline.NewID(pipeline.SignalMetrics)].Receivers)
	assert.Equal(t, []component.ID{component.MustNewIDWithName("discovery", "host_observer")},
		pipelines[pipeline.NewIDWithName(pipeline.SignalLogs, "entities")].Receivers)
	assert.Equal(t, []component.ID{component.MustNewIDWithName("otlphttp", "entities")},
		pipelines[pipeline.NewIDWithName(pipeline.SignalLogs, "entities")].Exporters)
}

func TestDiscoveryProvider_HostObserverDisabled(t *testing.T) {
	confmapProvider, err := New()
	require.NoError(t, err)

	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{
				fmt.Sprintf("file:%s", filepath.Join("testdata", "base-config.yaml")),
				fmt.Sprintf("%s:%s", propertiesFileScheme, filepath.Join("testdata", "disable-host-observer.properties.yaml")),
				fmt.Sprintf("%s:%s", discoveryModeScheme, filepath.Join("testdata", "config.d")),
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
