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
		c.BundleDir = "/opt/bin/agent/"
		c.ProcPath = "/my_proc"
		c.EtcPath = "/my_etc"
		c.VarPath = "/my_var"
		c.RunPath = "/my_run"
		c.SysPath = "/my_sys"
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

	require.Equal(t, "/opt/bin/agent/", saConfigProvider.SmartAgentConfig().BundleDir)
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
		},
	}
}
