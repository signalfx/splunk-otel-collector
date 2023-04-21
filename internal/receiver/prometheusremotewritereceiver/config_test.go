// Copyright Splunk, Inc.
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

package prometheusremotewritereceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestValidateConfig(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	assert.NoError(t, cfg.Validate())
	assert.NotEmpty(t, cfg.Endpoint)
	assert.NotEmpty(t, cfg.ListenPath)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 0, cfg.BufferSize)
}

func TestParseConfig(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()
	require.NotNil(t, cfg)

	rawCfgs := make(map[string]map[component.ID]map[string]any)
	conf, err := confmaptest.LoadConf("internal/testdata/otel-collector-config.yaml")
	require.NoError(t, err)
	require.NoError(t, conf.Unmarshal(&rawCfgs, confmap.WithErrorUnused()))
	require.NotEmpty(t, rawCfgs)

	require.NotEmpty(t, rawCfgs["receivers"])
	for id, value := range rawCfgs["receivers"] {
		require.NotEmpty(t, id)
		require.NotEmpty(t, value)
		assert.NoError(t, component.UnmarshalConfig(confmap.NewFromStringMap(value), cfg))
	}
	assert.NotEmpty(t, cfg)
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}
