// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
)

func serviceExtensionsFromConf(t *testing.T, conf *confmap.Conf) []any {
	t.Helper()
	service, err := getService(conf.ToStringMap())
	require.NoError(t, err)
	exts, err := getExtensionsFromService(service)
	require.NoError(t, err)
	return exts
}

func TestInjectConfigSourceTelemetryExtension_NilConf(t *testing.T) {
	require.NoError(t, InjectConfigSourceTelemetryExtension(t.Context(), nil))
}

func TestInjectConfigSourceTelemetryExtension_EmptyConf(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{})
	require.NoError(t, InjectConfigSourceTelemetryExtension(t.Context(), conf))

	out := conf.ToStringMap()

	extensions, ok := out["extensions"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, extensions, configsourcetelemetryextension.TypeStr)

	serviceExtensions := serviceExtensionsFromConf(t, conf)
	assert.Contains(t, serviceExtensions, configsourcetelemetryextension.TypeStr)
}

func TestInjectConfigSourceTelemetryExtension_ExistingExtensions(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{
		"extensions": map[string]any{
			"health_check": map[string]any{},
		},
		"service": map[string]any{
			"extensions": []any{"health_check"},
		},
	})

	require.NoError(t, InjectConfigSourceTelemetryExtension(t.Context(), conf))

	out := conf.ToStringMap()

	extensions, ok := out["extensions"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, extensions, "health_check")
	assert.Contains(t, extensions, configsourcetelemetryextension.TypeStr)

	serviceExtensions := serviceExtensionsFromConf(t, conf)
	assert.Contains(t, serviceExtensions, "health_check")
	assert.Contains(t, serviceExtensions, configsourcetelemetryextension.TypeStr)
}

func TestInjectConfigSourceTelemetryExtension_AlreadyPresent(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{
		"extensions": map[string]any{
			configsourcetelemetryextension.TypeStr: map[string]any{},
		},
		"service": map[string]any{
			"extensions": []any{configsourcetelemetryextension.TypeStr},
		},
	})

	require.NoError(t, InjectConfigSourceTelemetryExtension(t.Context(), conf))

	serviceExtensions := serviceExtensionsFromConf(t, conf)
	count := 0
	for _, e := range serviceExtensions {
		if e == configsourcetelemetryextension.TypeStr {
			count++
		}
	}
	assert.Equal(t, 1, count, "extension should appear exactly once in service.extensions")
}

func TestInjectConfigSourceTelemetryExtension_InvalidService(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{
		"service": "not-a-map",
	})
	err := InjectConfigSourceTelemetryExtension(t.Context(), conf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service is of unexpected form")
}

func TestInjectConfigSourceTelemetryExtension_InvalidExtensions(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{
		"extensions": "not-a-map",
	})
	err := InjectConfigSourceTelemetryExtension(t.Context(), conf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extensions is of unexpected form")
}
