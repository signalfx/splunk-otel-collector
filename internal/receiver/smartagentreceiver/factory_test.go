// Copyright OpenTelemetry Authors
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

package smartagentreceiver

import (
	"context"
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/monitors/haproxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	cfg.(*Config).monitorConfig = &haproxy.Config{}

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	assert.Same(t, receiver, receiverStore[cfg.(*Config)])
}

func TestCreateMetricsReceiverWithInvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.validate())

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "you must supply a valid Smart Agent Monitor config")
	assert.Nil(t, receiver)

	assert.NotContains(t, receiverStore, cfg)
}

func TestCreateLogsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	cfg.(*Config).monitorConfig = &haproxy.Config{}

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	assert.Same(t, receiver, receiverStore[cfg.(*Config)])
}

func TestCreateLogsReceiverWithInvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.validate())

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "you must supply a valid Smart Agent Monitor config")
	assert.Nil(t, receiver)

	assert.NotContains(t, receiverStore, cfg)
}

func TestCreateTracesReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	cfg.(*Config).monitorConfig = &haproxy.Config{}

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateTracesReceiver(context.Background(), params, cfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	assert.Same(t, receiver, receiverStore[cfg.(*Config)])
}

func TestCreateTracesReceiverWithInvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.validate())

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	receiver, err := factory.CreateTracesReceiver(context.Background(), params, cfg, consumertest.NewNop())
	require.Error(t, err)
	assert.EqualError(t, err, "you must supply a valid Smart Agent Monitor config")
	assert.Nil(t, receiver)

	assert.NotContains(t, receiverStore, cfg)
}

func TestCreateMetricsThenLogsAndThenTracesReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	cfg.(*Config).monitorConfig = &haproxy.Config{}

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	nextMetricsConsumer := consumertest.NewNop()
	metricsReceiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, nextMetricsConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, metricsReceiver)

	nextLogsConsumer := consumertest.NewNop()
	logsReceiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, nextLogsConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, logsReceiver)

	nextTracesConsumer := consumertest.NewNop()
	tracesReceiver, err := factory.CreateTracesReceiver(context.Background(), params, cfg, nextTracesConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, tracesReceiver)

	assert.Same(t, metricsReceiver, logsReceiver)
	assert.Same(t, logsReceiver, tracesReceiver)
	assert.Same(t, metricsReceiver, receiverStore[cfg.(*Config)])

	assert.Same(t, nextMetricsConsumer, metricsReceiver.(*Receiver).nextMetricsConsumer)
	assert.Same(t, nextLogsConsumer, metricsReceiver.(*Receiver).nextLogsConsumer)
	assert.Same(t, nextTracesConsumer, metricsReceiver.(*Receiver).nextTracesConsumer)
}

func TestCreateTracesThenLogsAndThenMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	cfg.(*Config).monitorConfig = &haproxy.Config{}

	params := component.ReceiverCreateSettings{Logger: zap.NewNop()}
	nextTracesConsumer := consumertest.NewNop()
	tracesReceiver, err := factory.CreateTracesReceiver(context.Background(), params, cfg, nextTracesConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, tracesReceiver)

	nextLogsConsumer := consumertest.NewNop()
	logsReceiver, err := factory.CreateLogsReceiver(context.Background(), params, cfg, nextLogsConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, logsReceiver)

	nextMetricsConsumer := consumertest.NewNop()
	metricsReceiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, nextMetricsConsumer)
	assert.NoError(t, err)
	assert.NotNil(t, metricsReceiver)

	assert.Same(t, metricsReceiver, logsReceiver)
	assert.Same(t, logsReceiver, tracesReceiver)
	assert.Same(t, metricsReceiver, receiverStore[cfg.(*Config)])

	assert.Same(t, nextMetricsConsumer, metricsReceiver.(*Receiver).nextMetricsConsumer)
	assert.Same(t, nextLogsConsumer, metricsReceiver.(*Receiver).nextLogsConsumer)
	assert.Same(t, nextTracesConsumer, metricsReceiver.(*Receiver).nextTracesConsumer)
}
