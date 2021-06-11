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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configparser"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestEnvVarConfigSourceNew(t *testing.T) {
	tests := []struct {
		config *Config
		name   string
	}{
		{
			name:   "minimal",
			config: &Config{},
		},
		{
			name: "with_defaults",
			config: &Config{
				Defaults: map[string]interface{}{
					"k0": "v0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgSrc := newConfigSource(zap.NewNop(), tt.config)
			require.NotNil(t, cfgSrc)
			require.NotNil(t, cfgSrc.defaults)
		})
	}
}

func TestEnvVarConfigSource_End2End(t *testing.T) {
	require.NoError(t, os.Setenv("_TEST_ENV_VAR_CFG_SRC", "test_env_var"))
	defer func() {
		assert.NoError(t, os.Unsetenv("_TEST_ENV_VAR_CFG_SRC"))
	}()

	file := path.Join("testdata", "env_config_source_end_2_end.yaml")
	p, err := configparser.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, p)

	factories := configprovider.Factories{
		"env": NewFactory(),
	}
	m, err := configprovider.NewManager(p, zap.NewNop(), component.DefaultBuildInfo(), factories)
	require.NoError(t, err)
	require.NotNil(t, m)

	ctx := context.Background()
	r, err := m.Resolve(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, r)

	go func() {
		_ = m.WatchForUpdate()
	}()
	m.WaitForWatcher()

	assert.NoError(t, m.Close(ctx))

	file = path.Join("testdata", "env_config_source_end_2_end_expected.yaml")
	expected, err := configparser.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, expected)

	assert.Equal(t, expected.ToStringMap(), r.ToStringMap())
}
