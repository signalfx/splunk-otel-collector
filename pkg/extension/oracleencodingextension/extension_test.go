// Copyright Splunk, Inc.
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

package oracleencodingextension

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestExtensionLifecycle(t *testing.T) {
	ctx := context.Background()
	createParams := extension.Settings{ID: component.MustNewID(typeStr)}
	cfg := &Config{}

	f := NewFactory()
	ext, err := f.Create(ctx, createParams, cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)

	require.NoError(t, ext.Start(ctx, componenttest.NewNopHost()))
	require.NoError(t, ext.Shutdown(ctx))
}

func TestUnmarshalMetrics(t *testing.T) {
	ctx := context.Background()
	createParams := extension.Settings{ID: component.MustNewID(typeStr)}
	f := NewFactory()
	ext, err := f.Create(ctx, createParams, &Config{})
	require.NoError(t, err)

	unmarshaler, ok := ext.(pmetric.Unmarshaler)
	require.True(t, ok)

	buf, err := os.ReadFile(filepath.Join("testdata", "metrics.jsonl"))
	require.NoError(t, err)

	metrics, err := unmarshaler.UnmarshalMetrics(buf)
	require.NoError(t, err)
	require.Equal(t, 1, metrics.ResourceMetrics().Len())

	scopeMetrics := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0)
	require.Equal(t, 2, scopeMetrics.Metrics().Len())
	require.Equal(t, pmetric.MetricTypeGauge, scopeMetrics.Metrics().At(0).Type())
	require.Equal(t, 2, scopeMetrics.Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, pmetric.MetricTypeGauge, scopeMetrics.Metrics().At(1).Type())
	require.Equal(t, 1, scopeMetrics.Metrics().At(1).Gauge().DataPoints().Len())
}
