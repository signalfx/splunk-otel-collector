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
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/featuregate"
)

func setOpAMPGate(t *testing.T, enabled bool) {
	t.Helper()
	require.NoError(t, featuregate.GlobalRegistry().Set(opampFeatureGateID, enabled))
	t.Cleanup(func() {
		_ = featuregate.GlobalRegistry().Set(opampFeatureGateID, false)
	})
}

func opampConf(extensions map[string]any, serviceExtensions []any) *confmap.Conf {
	return confmap.NewFromStringMap(map[string]any{
		"extensions": extensions,
		"service": map[string]any{
			"extensions": serviceExtensions,
		},
	})
}

func TestRemoveOpAMP_GateDisabled_OpAMPInServiceExtensions_IsRemoved(t *testing.T) {
	setOpAMPGate(t, false)

	conf := opampConf(
		map[string]any{
			"opamp":        map[string]any{"server": "http://localhost:4320"},
			"health_check": map[string]any{},
		},
		[]any{"opamp", "health_check"},
	)

	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))

	out := conf.ToStringMap()
	extensions := out["extensions"].(map[string]any)
	assert.Contains(t, extensions, "opamp")
	assert.Contains(t, extensions, "health_check")

	_, serviceExts, err := getServiceExtensions(out)
	require.NoError(t, err)
	assert.NotContains(t, serviceExts, "opamp")
	assert.Contains(t, serviceExts, "health_check")
}

func TestRemoveOpAMP_GateDisabled_OpAMPDefinedButNotInService_NoOp(t *testing.T) {
	setOpAMPGate(t, false)

	conf := opampConf(
		map[string]any{
			"opamp":        map[string]any{"server": "http://localhost:4320"},
			"health_check": map[string]any{},
		},
		[]any{"health_check"},
	)

	original := conf.ToStringMap()
	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))
	assert.Equal(t, original, conf.ToStringMap())
}

func TestRemoveOpAMP_GateDisabled_NamedOpAMPRemovedFromServiceExtensionsOnly(t *testing.T) {
	setOpAMPGate(t, false)

	conf := opampConf(
		map[string]any{
			"opamp/prod":   map[string]any{"server": "http://localhost:4320"},
			"health_check": map[string]any{},
		},
		[]any{"opamp/prod", "health_check"},
	)

	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))

	out := conf.ToStringMap()
	extensions := out["extensions"].(map[string]any)
	assert.Contains(t, extensions, "opamp/prod")
	assert.Contains(t, extensions, "health_check")

	_, serviceExts, err := getServiceExtensions(out)
	require.NoError(t, err)
	assert.NotContains(t, serviceExts, "opamp/prod")
	assert.Contains(t, serviceExts, "health_check")
}

func TestRemoveOpAMP_GateDisabled_OpAMPNotInConfig_NoOp(t *testing.T) {
	setOpAMPGate(t, false)

	conf := opampConf(
		map[string]any{"health_check": map[string]any{}},
		[]any{"health_check"},
	)

	original := conf.ToStringMap()
	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))
	assert.Equal(t, original, conf.ToStringMap())
}

func TestRemoveOpAMP_GateDisabled_NilConf_NoOp(t *testing.T) {
	setOpAMPGate(t, false)
	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), nil))
}

func TestRemoveOpAMP_GateEnabled_OpAMPInConfig_Untouched(t *testing.T) {
	setOpAMPGate(t, true)

	conf := opampConf(
		map[string]any{
			"opamp":        map[string]any{"server": "http://localhost:4320"},
			"health_check": map[string]any{},
		},
		[]any{"opamp", "health_check"},
	)

	original := conf.ToStringMap()
	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))
	assert.Equal(t, original, conf.ToStringMap())
}

func TestRemoveOpAMP_GateEnabled_OpAMPNotInConfig_LogsWarning(t *testing.T) {
	setOpAMPGate(t, true)

	conf := opampConf(
		map[string]any{"health_check": map[string]any{}},
		[]any{"health_check"},
	)
	original := conf.ToStringMap()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	require.NoError(t, RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf))

	assert.Equal(t, original, conf.ToStringMap())
	assert.Contains(t, buf.String(), opampFeatureGateID)
}

func TestRemoveOpAMP_GateDisabled_InvalidService_ReturnsError(t *testing.T) {
	setOpAMPGate(t, false)

	conf := confmap.NewFromStringMap(map[string]any{
		"extensions": map[string]any{"opamp": map[string]any{}},
		"service":    "not-a-map",
	})

	err := RemoveOpAMPIfFeatureGateDisabled(t.Context(), conf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service is of unexpected form")
}

func TestIsOpAMPKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"opamp", true},
		{"opamp/prod", true},
		{"opamp/staging", true},
		{"opampextension", false},
		{"health_check", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.expected, isOpAMPKey(tt.key))
		})
	}
}
