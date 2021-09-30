// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentextension

import (
	"context"
	"path"
	"path/filepath"
	"testing"
	"time"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/file"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

var tru = true
var flse = false

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfig(
		path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, len(cfg.Extensions), 3)

	emptyConfig := cfg.Extensions[config.NewIDWithName(typeStr, "default_settings")]
	require.NotNil(t, emptyConfig)
	require.NoError(t, configtest.CheckConfigStruct(emptyConfig))
	require.Equal(t, func() *Config {
		cfg := defaultConfig()
		cfg.ExtensionSettings.SetIDName("default_settings")
		return &cfg
	}(), emptyConfig)

	allSettingsConfig := cfg.Extensions[config.NewIDWithName(typeStr, "all_settings")]
	require.NotNil(t, allSettingsConfig)
	require.NoError(t, configtest.CheckConfigStruct(allSettingsConfig))
	require.Equal(t, func() *Config {
		cfg := defaultConfig()
		cfg.ExtensionSettings.SetIDName("all_settings")
		cfg.BundleDir = "/opt/bin/collectd/"
		cfg.ProcPath = "/my_proc"
		cfg.EtcPath = "/my_etc"
		cfg.VarPath = "/my_var"
		cfg.RunPath = "/my_run"
		cfg.SysPath = "/my_sys"
		cfg.Collectd.Timeout = 10
		cfg.Collectd.ReadThreads = 1
		cfg.Collectd.WriteThreads = 4
		cfg.Collectd.WriteQueueLimitHigh = 5
		cfg.Collectd.WriteQueueLimitLow = 1
		cfg.Collectd.LogLevel = "info"
		cfg.Collectd.IntervalSeconds = 5
		cfg.Collectd.WriteServerIPAddr = "10.100.12.1"
		cfg.Collectd.WriteServerPort = 9090
		cfg.Collectd.ConfigDir = "/etc/"
		cfg.Collectd.BundleDir = "/opt/bin/collectd/"
		cfg.Collectd.HasGenericJMXMonitor = false
		return &cfg
	}(), allSettingsConfig)

	partialSettingsConfig := cfg.Extensions[config.NewIDWithName(typeStr, "partial_settings")]
	require.NotNil(t, partialSettingsConfig)
	require.NoError(t, configtest.CheckConfigStruct(partialSettingsConfig))
	require.Equal(t, func() *Config {
		cfg := defaultConfig()
		cfg.ExtensionSettings.SetIDName("partial_settings")
		cfg.BundleDir = "/opt/"
		cfg.Collectd.Timeout = 10
		cfg.Collectd.ReadThreads = 1
		cfg.Collectd.WriteThreads = 4
		cfg.Collectd.WriteQueueLimitHigh = 5
		cfg.Collectd.ConfigDir = "/var/run/signalfx-agent/collectd"
		cfg.Collectd.BundleDir = "/opt/"
		return &cfg
	}(), partialSettingsConfig)
}

func TestSmartAgentConfigProvider(t *testing.T) {
	factories, err := componenttest.NopFactories()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfig(
		path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.GreaterOrEqual(t, len(cfg.Extensions), 1)

	allSettingsConfig := cfg.Extensions[config.NewIDWithName(typeStr, "all_settings")]
	require.NotNil(t, allSettingsConfig)

	ext, err := factory.CreateExtension(context.Background(), component.ExtensionCreateSettings{}, allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, ext)

	saConfigProvider, ok := ext.(SmartAgentConfigProvider)
	require.True(t, ok)

	require.Equal(t, func() saconfig.CollectdConfig {
		return saconfig.CollectdConfig{
			Timeout:              10,
			ReadThreads:          1,
			WriteThreads:         4,
			WriteQueueLimitHigh:  5,
			WriteQueueLimitLow:   1,
			LogLevel:             "info",
			IntervalSeconds:      5,
			WriteServerIPAddr:    "10.100.12.1",
			WriteServerPort:      9090,
			BundleDir:            "/opt/bin/collectd/",
			ConfigDir:            "/etc/",
			HasGenericJMXMonitor: false,
		}
	}(), saConfigProvider.SmartAgentConfig().Collectd)
	require.Equal(t, "/opt/bin/collectd/", saConfigProvider.SmartAgentConfig().BundleDir)
}

func TestLoadInvalidConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfig(
		path.Join(".", "testdata", "invalid_config.yaml"), factories,
	)

	require.Error(t, err)
	require.Nil(t, cfg)
}

func defaultConfig() Config {
	return Config{
		ExtensionSettings: config.NewExtensionSettings(config.NewID(typeStr)),
		Config: saconfig.Config{
			BundleDir:              bundleDir,
			SignalFxRealm:          "us0",
			IntervalSeconds:        10,
			CloudMetadataTimeout:   timeutil.Duration(2 * time.Second),
			GlobalDimensions:       map[string]string{},
			GlobalSpanTags:         map[string]string{},
			ValidateDiscoveryRules: &flse,
			Observers:              []saconfig.ObserverConfig{},
			Monitors:               []saconfig.MonitorConfig{},
			EnableBuiltInFiltering: &tru,
			InternalStatusHost:     "localhost",
			InternalStatusPort:     8095,
			ProfilingHost:          "127.0.0.1",
			ProfilingPort:          6060,
			ProcPath:               "/proc",
			EtcPath:                "/etc",
			VarPath:                "/var",
			RunPath:                "/run",
			SysPath:                "/sys",
			Collectd: saconfig.CollectdConfig{
				Timeout:              40,
				ReadThreads:          5,
				WriteThreads:         2,
				WriteQueueLimitHigh:  500000,
				WriteQueueLimitLow:   400000,
				LogLevel:             "notice",
				IntervalSeconds:      10,
				WriteServerIPAddr:    "127.9.8.7",
				WriteServerPort:      0,
				ConfigDir:            filepath.Join(bundleDir, "run", "collectd"),
				BundleDir:            bundleDir,
				HasGenericJMXMonitor: false,
			},
			Writer: saconfig.WriterConfig{
				DatapointMaxBatchSize:                 1000,
				MaxDatapointsBuffered:                 25000,
				TraceSpanMaxBatchSize:                 1000,
				TraceExportFormat:                     "zipkin",
				MaxRequests:                           10,
				Timeout:                               timeutil.Duration(5 * time.Second),
				EventSendIntervalSeconds:              1,
				PropertiesMaxRequests:                 20,
				PropertiesMaxBuffered:                 10000,
				PropertiesSendDelaySeconds:            30,
				PropertiesHistorySize:                 10000,
				LogTraceSpans:                         false,
				LogDimensionUpdates:                   false,
				LogDroppedDatapoints:                  false,
				SendTraceHostCorrelationMetrics:       &tru,
				StaleServiceTimeout:                   timeutil.Duration(5 * time.Minute),
				TraceHostCorrelationPurgeInterval:     timeutil.Duration(1 * time.Minute),
				TraceHostCorrelationMetricsInterval:   timeutil.Duration(1 * time.Minute),
				TraceHostCorrelationMaxRequestRetries: 2,
				MaxTraceSpansInFlight:                 100000,
				Splunk:                                nil,
				SignalFxEnabled:                       &tru,
				ExtraHeaders:                          nil,
				HostIDDims:                            nil,
				IngestURL:                             "",
				APIURL:                                "",
				EventEndpointURL:                      "",
				TraceEndpointURL:                      "",
				SignalFxAccessToken:                   "",
				GlobalDimensions:                      nil,
				GlobalSpanTags:                        nil,
				MetricsToInclude:                      nil,
				MetricsToExclude:                      nil,
				PropertiesToExclude:                   nil,
			},
			Logging: saconfig.LogConfig{
				Level:  "info",
				Format: "text",
			},
			PropertiesToExclude: []saconfig.PropertyFilterConfig{},
			MetricsToExclude:    []saconfig.MetricFilter{},
			MetricsToInclude:    []saconfig.MetricFilter{},
			Sources: sources.SourceConfig{
				Watch: &tru,
				File: file.Config{
					PollRateSeconds: 5,
				},
				RemoteWatch: &tru,
			},
		},
	}
}
