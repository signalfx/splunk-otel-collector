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

package envvarconfigsource

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

func TestEnvVarConfigSourceLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := confmaptest.LoadConf(fileName)
	require.NoError(t, err)

	factories := map[component.Type]configsource.Factory{
		typeStr: NewFactory(),
	}

	actualSettings, splitConf, err := configsource.SettingsFromConf(context.Background(), v, factories, nil)
	require.NoError(t, err)
	require.NotNil(t, splitConf)

	expectedSettings := map[string]configsource.Settings{
		"env": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
		},
		"env/with_fallback": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewIDWithName(typeStr, "with_fallback")),
			Defaults: map[string]any{
				"k0": 42,
				"m0": map[string]any{
					"k0": "v0",
					"k1": "v1",
				},
			},
		},
	}

	require.Equal(t, expectedSettings, actualSettings)
	require.Empty(t, splitConf.ToStringMap())

	_, err = configsource.BuildConfigSources(context.Background(), actualSettings, zap.NewNop(), factories)
	require.NoError(t, err)
}
