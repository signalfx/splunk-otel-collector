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
// +build windows

package smartagentextension

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestBundleDirDefault(t *testing.T) {
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

	allSettingsConfig := cfg.Extensions[config.NewComponentIDWithName(typeStr, "default_settings")]
	require.NotNil(t, allSettingsConfig)

	ext, err := factory.CreateExtension(context.Background(), component.ExtensionCreateSettings{}, allSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, ext)

	saConfigProvider, ok := ext.(SmartAgentConfigProvider)
	require.True(t, ok)

	require.Equal(t, "C:\\Program Files\\Splunk\\OpenTelemetry Collector\\agent-bundle", saConfigProvider.SmartAgentConfig().BundleDir)
}
