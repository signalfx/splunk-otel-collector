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

package discoveryreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	require.Equal(t, &Config{
		ReceiverSettings:    config.NewReceiverSettings(component.NewID(typeStr)),
		LogEndpoints:        false,
		EmbedReceiverConfig: false,
		CorrelationTTL:      10 * time.Minute,
	}, cfg)
}

func TestCreateLogsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	params := component.ReceiverCreateSettings{}
	receiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	assert.Error(t, err)
	assert.EqualError(t, err, "`watch_observers` must be defined and include at least one configured observer extension")
	assert.Nil(t, receiver)
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.Validate())

	params := component.ReceiverCreateSettings{}
	receiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "telemetry type is not supported")
	assert.Nil(t, receiver)
}

func TestCreateTracesReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.Validate())

	params := component.ReceiverCreateSettings{}
	receiver, err := factory.CreateTracesReceiver(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "telemetry type is not supported")
	assert.Nil(t, receiver)
}
