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

//go:build windows

package smartagentextension

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/extension"
)

func TestBundleDirDefault(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.GreaterOrEqual(t, len(cfg.AllKeys()), 1)

	defaultSettingsID := component.NewIDWithName("smartagent", "default_settings")
	cm, err := cfg.Sub(defaultSettingsID.String())
	require.NoError(t, err)
	emptyConfig := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, emptyConfig)
	require.NoError(t, err)

	ext, err := factory.CreateExtension(context.Background(), extension.CreateSettings{}, emptyConfig)
	require.NoError(t, err)
	require.NotNil(t, ext)

	saConfigProvider, ok := ext.(SmartAgentConfigProvider)
	require.True(t, ok)

	require.Equal(t, "C:\\Program Files\\Splunk\\OpenTelemetry Collector\\agent-bundle", saConfigProvider.SmartAgentConfig().BundleDir)
}
