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
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	require.Equal(t, &Config{
		EmbedReceiverConfig: false,
		CorrelationTTL:      10 * time.Minute,
	}, cfg)
}

func TestLogsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	params := receivertest.NewNopSettings()
	receiver, err := factory.CreateLogs(context.Background(), params, cfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, receiver.Shutdown(context.Background()))
}

func TestMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.Validate())

	params := receivertest.NewNopSettings()
	receiver, err := factory.CreateMetrics(context.Background(), params, cfg, consumertest.NewNop())
	require.NoError(t, err)
	assert.NotNil(t, receiver)

	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, receiver.Shutdown(context.Background()))
}

func TestTracesReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.Validate())

	params := receivertest.NewNopSettings()
	receiver, err := factory.CreateTraces(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "telemetry type is not supported")
	assert.Nil(t, receiver)
}
