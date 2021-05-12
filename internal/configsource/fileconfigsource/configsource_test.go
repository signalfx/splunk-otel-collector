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

package fileconfigsource

import (
	"context"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestFileConfigSourceNew(t *testing.T) {
	tests := []struct {
		config *Config
		name   string
	}{
		{
			name:   "default",
			config: &Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgSrc, err := newConfigSource(zap.NewNop(), tt.config)
			require.NoError(t, err)
			require.NotNil(t, cfgSrc)
		})
	}
}

func TestFileConfigSource_End2End(t *testing.T) {
	file := path.Join("testdata", "file_config_source_end_2_end.yaml")
	p, err := config.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, p)

	factories := configprovider.Factories{
		"file": NewFactory(),
	}
	m, err := configprovider.NewManager(p, zap.NewNop(), component.DefaultBuildInfo(), factories)
	require.NoError(t, err)
	require.NotNil(t, m)

	ctx := context.Background()
	r, err := m.Resolve(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, r)
	t.Cleanup(func() {
		assert.NoError(t, m.Close(ctx))
	})

	var watchErr error
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		watchErr = m.WatchForUpdate()
	}()
	m.WaitForWatcher()

	file = path.Join("testdata", "file_config_source_end_2_end_expected.yaml")
	expected, err := config.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, expected)

	assert.Equal(t, expected.ToStringMap(), r.ToStringMap())

	// Touch one of the files to trigger an update.
	yamlDataFile := path.Join("testdata", "yaml_data_file")
	touchFile(t, yamlDataFile)

	select {
	case <-watchDone:
	case <-time.After(3 * time.Second):
		require.Fail(t, "expected file change notification didn't happen")
	}

	assert.ErrorIs(t, watchErr, configsource.ErrValueUpdated)

	// Value should not have changed.
	expected, err = config.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, expected)
}

func touchFile(t *testing.T, file string) {
	contents, err := ioutil.ReadFile(file)
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(file, contents, 0))
}
