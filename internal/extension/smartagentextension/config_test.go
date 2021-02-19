// Copyright 2021, OpenTelemetry Authors
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
	"path"
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.ExampleComponents()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, len(cfg.Extensions), 3)

	emptyConfig := cfg.Extensions["smartagent/default_settings"]
	require.NotNil(t, emptyConfig)
	require.NoError(t, configcheck.ValidateConfig(emptyConfig))
	require.Equal(t, func() *Config {
		return &Config{
			ExtensionSettings: configmodels.ExtensionSettings{
				TypeVal: "smartagent",
				NameVal: "smartagent/default_settings",
			},
			bundleDir: bundleDir,
			collectdConfig: config.CollectdConfig{
				Timeout:              40,
				ReadThreads:          5,
				WriteThreads:         2,
				WriteQueueLimitHigh:  500000,
				WriteQueueLimitLow:   400000,
				LogLevel:             "notice",
				IntervalSeconds:      10,
				WriteServerIPAddr:    "127.9.8.7",
				WriteServerPort:      0,
				ConfigDir:            "/var/run/signalfx-agent/collectd",
				HasGenericJMXMonitor: true,
			},
		}
	}(), emptyConfig)

	allSettingsConfig := cfg.Extensions["smartagent/all_settings"]
	require.NotNil(t, allSettingsConfig)
	require.NoError(t, configcheck.ValidateConfig(allSettingsConfig))
	require.Equal(t, func() *Config {
		return &Config{
			ExtensionSettings: configmodels.ExtensionSettings{
				TypeVal: "smartagent",
				NameVal: "smartagent/all_settings",
			},
			bundleDir: "/opt/bin/collectd/",
			collectdConfig: config.CollectdConfig{
				Timeout:              10,
				ReadThreads:          1,
				WriteThreads:         4,
				WriteQueueLimitHigh:  5,
				WriteQueueLimitLow:   1,
				LogLevel:             "info",
				IntervalSeconds:      5,
				WriteServerIPAddr:    "10.100.12.1",
				WriteServerPort:      9090,
				ConfigDir:            "/etc/",
				BundleDir:            "/opt/bin/collectd/",
				HasGenericJMXMonitor: true,
			},
		}
	}(), allSettingsConfig)

	partialSettingsConfig := cfg.Extensions["smartagent/partial_settings"]
	require.NotNil(t, partialSettingsConfig)
	require.NoError(t, configcheck.ValidateConfig(partialSettingsConfig))
	require.Equal(t, func() *Config {
		return &Config{
			ExtensionSettings: configmodels.ExtensionSettings{
				TypeVal: "smartagent",
				NameVal: "smartagent/partial_settings",
			},
			bundleDir: "/opt/",
			collectdConfig: config.CollectdConfig{
				Timeout:              10,
				ReadThreads:          1,
				WriteThreads:         4,
				WriteQueueLimitHigh:  5,
				WriteQueueLimitLow:   400000,
				LogLevel:             "notice",
				IntervalSeconds:      10,
				WriteServerIPAddr:    "127.9.8.7",
				WriteServerPort:      0,
				BundleDir:            "/opt/",
				ConfigDir:            "/etc/",
				HasGenericJMXMonitor: true,
			},
		}
	}(), partialSettingsConfig)
}

func TestSmartAgentConfigProvider(t *testing.T) {
	factories, err := componenttest.ExampleComponents()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.GreaterOrEqual(t, len(cfg.Extensions), 1)

	allSettingsConfig := cfg.Extensions["smartagent/all_settings"]
	require.NotNil(t, allSettingsConfig)

	saConfigProvider, ok := allSettingsConfig.(SmartAgentConfigProvider)
	require.True(t, ok)

	require.Equal(t, func() *config.CollectdConfig {
		return &config.CollectdConfig{
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
			HasGenericJMXMonitor: true,
		}
	}(), saConfigProvider.CollectdConfig())
	require.Equal(t, "/opt/bin/collectd/", saConfigProvider.BundleDir())
}

func TestLoadInvalidConfig(t *testing.T) {
	factories, err := componenttest.ExampleComponents()
	require.Nil(t, err)

	factory := NewFactory()
	factories.Extensions[typeStr] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_config.yaml"), factories,
	)

	require.Error(t, err)
	require.Nil(t, cfg)
}
