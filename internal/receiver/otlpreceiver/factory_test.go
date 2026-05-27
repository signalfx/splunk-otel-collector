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

package otlpreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/receiver/xreceiver"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.IsType(t, &Config{}, cfg)
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
	assert.Zero(t, cfg.(*Config).StartDelay)
}

func TestConfigStartDelay(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig().(*Config)
	cm := confmap.NewFromStringMap(map[string]any{
		"start_delay": "25ms",
		"protocols": map[string]any{
			"grpc": map[string]any{},
			"http": map[string]any{},
		},
	})
	require.NoError(t, cm.Unmarshal(cfg))
	assert.Equal(t, 25*time.Millisecond, cfg.StartDelay)
	assert.NoError(t, cfg.Validate())
}

func TestConfigRejectsNegativeStartDelay(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig().(*Config)
	cfg.StartDelay = -time.Second
	assert.EqualError(t, cfg.Validate(), "start_delay must be non-negative")
}

func TestCreateSameReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint = "127.0.0.1:0"
	cfg.Protocols.HTTP.GetOrInsertDefault().ServerConfig.NetAddr.Endpoint = "127.0.0.1:0"

	settings := receivertest.NewNopSettings(component.MustNewType(typeStr))

	tracesReceiver, err := factory.CreateTraces(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	metricsReceiver, err := factory.CreateMetrics(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	logsReceiver, err := factory.CreateLogs(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	profilesReceiver, err := factory.(xreceiver.Factory).CreateProfiles(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	assert.Same(t, tracesReceiver, metricsReceiver)
	assert.Same(t, tracesReceiver, logsReceiver)
	assert.Same(t, tracesReceiver, profilesReceiver)
}

func TestStartDelaysAndLogsElapsedTime(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.StartDelay = 30 * time.Millisecond
	cfg.Protocols.GRPC.GetOrInsertDefault().NetAddr = confignet.AddrConfig{
		Endpoint:  "127.0.0.1:0",
		Transport: confignet.TransportTypeTCP,
	}
	cfg.Protocols.HTTP.GetOrInsertDefault().ServerConfig.NetAddr = confignet.AddrConfig{
		Endpoint:  "127.0.0.1:0",
		Transport: confignet.TransportTypeTCP,
	}

	core, logs := observer.New(zap.InfoLevel)
	settings := receivertest.NewNopSettings(component.MustNewType(typeStr))
	settings.Logger = zap.New(core)

	receiver, err := factory.CreateTraces(context.Background(), settings, cfg, consumertest.NewNop())
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})

	assert.GreaterOrEqual(t, time.Since(start), cfg.StartDelay)
	entries := logs.FilterMessage("Starting OTLP receiver after configured delay").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.Equal(t, cfg.StartDelay, fields["configured_delay"])
	assert.GreaterOrEqual(t, fields["elapsed_since_collector_start"], cfg.StartDelay)

	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	assert.Equal(t, 1, logs.FilterMessage("Starting OTLP receiver after configured delay").Len())
}
