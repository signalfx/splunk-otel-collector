// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package includeconfigsource

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

func TestIncludeConfigSourceLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := confmaptest.LoadConf(fileName)
	require.NoError(t, err)

	factories := map[component.Type]configsource.Factory{
		component.MustNewType(typeStr): NewFactory(),
	}

	actualSettings, splitConf, err := configsource.SettingsFromConf(context.Background(), v, factories, nil)
	require.NoError(t, err)
	require.NotNil(t, splitConf)

	expectedSettings := map[string]configsource.Settings{
		"include": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
		},
		"include/delete_files": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewIDWithName(typeStr, "delete_files")),
			DeleteFiles:    true, // SA1019: deprecated DeleteFiles
		},
		"include/watch_files": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewIDWithName(typeStr, "watch_files")),
			WatchFiles:     true,
		},
	}

	require.Equal(t, expectedSettings, actualSettings)
	require.Empty(t, splitConf.ToStringMap())

	cfgSrcs, err := configsource.BuildConfigSources(context.Background(), actualSettings, zap.NewNop(), factories)
	require.NoError(t, err)
	for k := range expectedSettings {
		assert.Contains(t, cfgSrcs, k)
	}
}
