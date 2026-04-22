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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
)

func TestInjectConfigSourceTelemetryExtension_NilConf(t *testing.T) {
	err := InjectConfigSourceTelemetryExtension(context.Background(), nil)
	require.NoError(t, err)
}

func TestInjectConfigSourceTelemetryExtension_EmptyConf(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]any{})
	err := InjectConfigSourceTelemetryExtension(context.Background(), conf)
	require.NoError(t, err)

	out := conf.ToStringMap()

	// Extension should be registered
	extensions, ok := out["extensions"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, extensions, configsourcetelemetryextension.TypeStr)

	// Service extensions list should contain our extension
	service, ok := out["service"].(map[string]any)
	require.True(t, ok)
	serviceExtensions := toStringSlice(service["extensions"])
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

	err := InjectConfigSourceTelemetryExtension(context.Background(), conf)
	require.NoError(t, err)

	out := conf.ToStringMap()

	// Both extensions should be present
	extensions, ok := out["extensions"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, extensions, "health_check")
	assert.Contains(t, extensions, configsourcetelemetryextension.TypeStr)

	// Service extensions list should contain both
	service, ok := out["service"].(map[string]any)
	require.True(t, ok)
	serviceExtensions := toStringSlice(service["extensions"])
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

	err := InjectConfigSourceTelemetryExtension(context.Background(), conf)
	require.NoError(t, err)

	out := conf.ToStringMap()

	// Should appear only once in service extensions
	service, ok := out["service"].(map[string]any)
	require.True(t, ok)
	serviceExtensions := toStringSlice(service["extensions"])
	count := 0
	for _, e := range serviceExtensions {
		if e == configsourcetelemetryextension.TypeStr {
			count++
		}
	}
	assert.Equal(t, 1, count, "extension should appear exactly once")
}
