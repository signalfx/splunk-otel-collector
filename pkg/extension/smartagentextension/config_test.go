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
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/extension"
)

var tru = true
var flse = false

func TestLoadConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, 3, len(cfg.ToStringMap()))

	defaultSettingsID := component.NewIDWithName("smartagent", "default_settings")
	cm, err := cfg.Sub(defaultSettingsID.String())
	require.NoError(t, err)
	emptyConfig := createDefaultConfig()
	err = component.UnmarshalConfig(cm, emptyConfig)
	require.NoError(t, err)
	require.NotNil(t, emptyConfig)
	require.NoError(t, componenttest.CheckConfigStruct(emptyConfig))
	require.Equal(t, func() *Config {
		c := defaultConfig()
		return &c
	}(), emptyConfig)

	allSettingsID := component.NewIDWithName("smartagent", "all_settings")
	cm, err = cfg.Sub(allSettingsID.String())
	require.NoError(t, err)
	allSettingsConfig := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, allSettingsConfig)
	require.NoError(t, componenttest.CheckConfigStruct(allSettingsConfig))
	require.Equal(t, func() *Config {
		c := defaultConfig()
		c.BundleDir = "/opt/bin/collectd/"
		c.ProcPath = "/my_proc"
		c.EtcPath = "/my_etc"
		c.VarPath = "/my_var"
		c.RunPath = "/my_run"
		c.SysPath = "/my_sys"
		c.Collectd.Timeout = 10
		c.Collectd.ReadThreads = 1
		c.Collectd.WriteThreads = 4
		c.Collectd.WriteQueueLimitHigh = 5
		c.Collectd.WriteQueueLimitLow = 1
		c.Collectd.LogLevel = "info"
		c.Collectd.IntervalSeconds = 5
		c.Collectd.WriteServerIPAddr = "10.100.12.1"
		c.Collectd.WriteServerPort = 9090
		c.Collectd.ConfigDir = "/etc/"
		c.Collectd.BundleDir = "/opt/bin/collectd/"
		c.Collectd.HasGenericJMXMonitor = false
		return &c
	}(), allSettingsConfig)

	partialSettingsID := component.NewIDWithName("smartagent", "partial_settings")
	cm, err = cfg.Sub(partialSettingsID.String())
	require.NoError(t, err)
	partialSettingsConfig := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, partialSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, partialSettingsConfig)
	require.NoError(t, componenttest.CheckConfigStruct(partialSettingsConfig))
	require.Equal(t, func() *Config {
		c := defaultConfig()
		c.BundleDir = "/opt/"
		c.Collectd.Timeout = 10
		c.Collectd.ReadThreads = 1
		c.Collectd.WriteThreads = 4
		c.Collectd.WriteQueueLimitHigh = 5
		c.Collectd.ConfigDir = "/var/run/signalfx-agent/collectd"
		c.Collectd.BundleDir = "/opt/"
		return &c
	}(), partialSettingsConfig)
}

func TestSmartAgentConfigProvider(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, 3, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "all_settings").String())
	require.NoError(t, err)
	allSettingsConfig := createDefaultConfig()
	err = component.UnmarshalConfig(cm, allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, allSettingsConfig)

	ext, err := NewFactory().CreateExtension(context.Background(), extension.CreateSettings{}, allSettingsConfig)
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_config.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/invalid_settings")
	require.NoError(t, err)
	err = component.UnmarshalConfig(cm, createDefaultConfig())
	require.Error(t, err)
}

func defaultConfig() Config {
	return Config{
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
