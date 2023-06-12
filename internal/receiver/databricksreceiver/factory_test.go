// Copyright The OpenTelemetry Authors
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

package databricksreceiver

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestFactory(t *testing.T) {
	f := NewFactory()
	assert.EqualValues(t, "databricks", f.Type())
	cfg := f.CreateDefaultConfig()
	assert.NotNil(t, cfg)
	duration, _ := time.ParseDuration("30s")
	assert.Equal(t, duration, cfg.(*Config).ScraperControllerSettings.CollectionInterval)
}

func TestCreateReceiver(t *testing.T) {
	ctx := context.Background()
	f := newReceiverFactory()
	receiver, err := f(
		ctx,
		otelcolreceiver.CreateSettings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.NewNop(),
				MeterProvider:  noop.NewMeterProvider(),
				TracerProvider: trace.NewNoopTracerProvider(),
			},
		},
		createDefaultConfig(),
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	err = receiver.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	err = receiver.Shutdown(ctx)
	require.NoError(t, err)
}

func TestParseConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub(component.NewID(typeStr).String())
	require.NoError(t, err)
	rcfg := createDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, rcfg)
	require.NoError(t, err)
	assert.Equal(t, "my-instance", rcfg.InstanceName)
	assert.Equal(t, "abc123", rcfg.Token)
	assert.Equal(t, "https://dbr.example.net", rcfg.Endpoint)
	duration, _ := time.ParseDuration("10s")
	assert.Equal(t, duration, rcfg.CollectionInterval)
}
