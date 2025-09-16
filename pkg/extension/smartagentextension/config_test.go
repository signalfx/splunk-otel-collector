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

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/extension"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, 3, len(cfg.ToStringMap()))

	defaultSettingsID := component.MustNewIDWithName("smartagent", "default_settings")
	cm, err := cfg.Sub(defaultSettingsID.String())
	require.NoError(t, err)
	emptyConfig := createDefaultConfig()
	err = cm.Unmarshal(&emptyConfig)
	require.NoError(t, err)
	require.NotNil(t, emptyConfig)
	require.NoError(t, componenttest.CheckConfigStruct(emptyConfig))
	require.Equal(t, func() *Config {
		c := defaultConfig()
		return &c
	}(), emptyConfig)

	allSettingsID := component.MustNewIDWithName("smartagent", "all_settings")
	cm, err = cfg.Sub(allSettingsID.String())
	require.NoError(t, err)
	allSettingsConfig := NewFactory().CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&allSettingsConfig)
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
		c.Collectd.ReadThreads = 1
		c.Collectd.WriteThreads = 4
		c.Collectd.WriteQueueLimitHigh = 5
		c.Collectd.WriteQueueLimitLow = 1
		c.Collectd.IntervalSeconds = 5
		c.Collectd.WriteServerIPAddr = "10.100.12.1"
		c.Collectd.WriteServerPort = 9090
		c.Collectd.ConfigDir = "/etc/"
		c.Collectd.BundleDir = "/opt/bin/collectd/"
		return &c
	}(), allSettingsConfig)

	partialSettingsID := component.MustNewIDWithName("smartagent", "partial_settings")
	cm, err = cfg.Sub(partialSettingsID.String())
	require.NoError(t, err)
	partialSettingsConfig := NewFactory().CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&partialSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, partialSettingsConfig)
	require.NoError(t, componenttest.CheckConfigStruct(partialSettingsConfig))
	require.Equal(t, func() *Config {
		c := defaultConfig()
		c.BundleDir = "/opt/"
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

	idWithName := component.MustNewIDWithName(typeStr, "all_settings")
	cm, err := cfg.Sub(idWithName.String())
	require.NoError(t, err)
	allSettingsConfig := createDefaultConfig()
	err = cm.Unmarshal(&allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, allSettingsConfig)

	ext, err := NewFactory().Create(context.Background(), extension.Settings{ID: idWithName}, allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, ext)

	saConfigProvider, ok := ext.(SmartAgentConfigProvider)
	require.True(t, ok)

	require.Equal(t, func() saconfig.CollectdConfig {
		return saconfig.CollectdConfig{
			Timeout:             40,
			LogLevel:            "notice",
			ReadThreads:         1,
			WriteThreads:        4,
			WriteQueueLimitHigh: 5,
			WriteQueueLimitLow:  1,
			IntervalSeconds:     5,
			WriteServerIPAddr:   "10.100.12.1",
			WriteServerPort:     9090,
			BundleDir:           "/opt/bin/collectd/",
			ConfigDir:           "/etc/",
		}
	}(), saConfigProvider.SmartAgentConfig().Collectd)
	require.Equal(t, "/opt/bin/collectd/", saConfigProvider.SmartAgentConfig().BundleDir)
}

func TestLoadInvalidConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_config.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/invalid_settings")
	require.NoError(t, err)
	invalidSettingsConfig := createDefaultConfig()
	err = cm.Unmarshal(&invalidSettingsConfig)
	require.Error(t, err)
}

func defaultConfig() Config {
	return Config{
		Config: saconfig.Config{
			BundleDir: bundleDir,
			ProcPath:  "/proc",
			EtcPath:   "/etc",
			VarPath:   "/var",
			RunPath:   "/run",
			SysPath:   "/sys",
			Collectd: saconfig.CollectdConfig{
				Timeout:             40,
				LogLevel:            "notice",
				ReadThreads:         5,
				WriteThreads:        2,
				WriteQueueLimitHigh: 500000,
				WriteQueueLimitLow:  400000,
				IntervalSeconds:     10,
				WriteServerIPAddr:   "127.9.8.7",
				WriteServerPort:     0,
				ConfigDir:           filepath.Join(bundleDir, "run", "collectd"),
				BundleDir:           bundleDir,
			},
		},
	}
}
