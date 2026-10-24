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

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestRedirectTALogsToStderr(t *testing.T) {
	cfg := confmap.NewFromStringMap(map[string]any{
		"service": map[string]any{
			"telemetry": map[string]any{
				"logs": map[string]any{
					"level":              "debug",
					"output_paths":       []any{"stdout", "stderr", "C:/collector.log"},
					"error_output_paths": []any{"/dev/stdout"},
				},
			},
		},
	})

	require.NoError(t, RedirectTALogsToStderr(t.Context(), cfg))
	logs := cfg.ToStringMap()["service"].(map[string]any)["telemetry"].(map[string]any)["logs"].(map[string]any)
	require.Equal(t, []any{"stderr", "C:/collector.log"}, logs["output_paths"])
	require.Equal(t, []any{"stderr"}, logs["error_output_paths"])
	require.Equal(t, "debug", logs["level"])
}

func TestRedirectTALogsToStderrLeavesOtherDestinations(t *testing.T) {
	cfg := confmap.NewFromStringMap(map[string]any{
		"service": map[string]any{
			"telemetry": map[string]any{
				"logs": map[string]any{
					"output_paths": []any{"stderr", "C:/collector.log"},
				},
			},
		},
	})
	original := cfg.ToStringMap()

	require.NoError(t, RedirectTALogsToStderr(t.Context(), cfg))
	require.Equal(t, original, cfg.ToStringMap())
	require.NoError(t, RedirectTALogsToStderr(t.Context(), nil))
}

func TestRedirectTALogsToStderrIgnoresMissingLogSettings(t *testing.T) {
	for _, tc := range []struct {
		cfg  map[string]any
		name string
	}{
		{name: "service", cfg: map[string]any{}},
		{name: "telemetry", cfg: map[string]any{"service": map[string]any{}}},
		{name: "logs", cfg: map[string]any{"service": map[string]any{"telemetry": map[string]any{}}}},
		{name: "paths", cfg: map[string]any{"service": map[string]any{"telemetry": map[string]any{"logs": map[string]any{"output_paths": "stdout"}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := confmap.NewFromStringMap(tc.cfg)
			original := cfg.ToStringMap()
			require.NoError(t, RedirectTALogsToStderr(t.Context(), cfg))
			require.Equal(t, original, cfg.ToStringMap())
		})
	}
}

func TestRedirectTALogsToStderrPreservesOtherPathTypes(t *testing.T) {
	cfg := confmap.NewFromStringMap(map[string]any{
		"service": map[string]any{
			"telemetry": map[string]any{
				"logs": map[string]any{"output_paths": []any{42, "stdout"}},
			},
		},
	})

	require.NoError(t, RedirectTALogsToStderr(t.Context(), cfg))
	logs := cfg.ToStringMap()["service"].(map[string]any)["telemetry"].(map[string]any)["logs"].(map[string]any)
	require.Equal(t, []any{42, "stderr"}, logs["output_paths"])
}
