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
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/service/servicetest"
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
	f := createReceiverFunc(func(string, string, *http.Client, *zap.Logger) apiClientInterface { return &testdataClient{} })
	receiver, err := f(
		ctx,
		component.ReceiverCreateSettings{
			TelemetrySettings: component.TelemetrySettings{
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
	factories, err := componenttest.NopFactories()
	require.NoError(t, err)
	factories.Receivers[typeStr] = NewFactory()
	cfg, err := servicetest.LoadConfigAndValidate(path.Join("testdata", "config.yaml"), factories)
	require.NoError(t, err)
	rcfg := cfg.Receivers[config.NewComponentID(typeStr)].(*Config)
	assert.Equal(t, "my-instance", rcfg.InstanceName)
	assert.Equal(t, "abc123", rcfg.Token)
	assert.Equal(t, "https://my.databricks.instance", rcfg.Endpoint)
	duration, _ := time.ParseDuration("10s")
	assert.Equal(t, duration, rcfg.CollectionInterval)
}
