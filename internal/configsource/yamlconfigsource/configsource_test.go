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

package yamlconfigsource

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestYAMLConfigSourceNew(t *testing.T) {
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

func TestYAMLConfigSource_End2End(t *testing.T) {
	file := path.Join("testdata", "templated.yaml")
	p, err := config.NewParserFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, p)

	factories := configprovider.Factories{
		"yaml": NewFactory(),
	}
	m, err := configprovider.NewManager(p, zap.NewNop(), component.DefaultBuildInfo(), factories)
	require.NoError(t, err)
	require.NotNil(t, m)

	ctx := context.Background()
	r, err := m.Resolve(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, r)

	var watchErr error
	watchDone := make(chan struct{})
	t.Cleanup(func() {
		assert.NoError(t, m.Close(ctx))
		<-watchDone
		assert.ErrorIs(t, watchErr, configsource.ErrSessionClosed)
	})

	go func() {
		defer close(watchDone)
		watchErr = m.WatchForUpdate()
	}()
	m.WaitForWatcher()

	file = path.Join("testdata", "templated_expected.yaml")
	expected, err := config.NewParserFromFile(file)
	assert.NoError(t, err)
	assert.NotNil(t, expected)
	assert.Equal(t, expected.ToStringMap(), r.ToStringMap())
}
