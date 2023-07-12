// Copyright Splunk, Inc.
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

package signalfxgatewayprometheusremotewritereceiver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver/internal/metadata"
)

func TestFactory(t *testing.T) {
	timeout := time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	cfg.Endpoint = fmt.Sprintf("localhost:%d", freePort)
	cfg.ListenPath = "/metrics"

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()
	receiver, err := New(mockSettings, cfg, mockConsumer)

	assert.NoError(t, err)
	require.NotNil(t, receiver)
	require.NoError(t, receiver.Start(ctx, nopHost))
	require.NoError(t, receiver.Shutdown(ctx))
}

func TestNewFactory(t *testing.T) {
	fact := NewFactory()
	assert.NotNil(t, fact)
	assert.NotEmpty(t, fact.Type())
	assert.NotEmpty(t, fact.CreateDefaultConfig())
}

func TestFactoryOtelIntegration(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()
	require.NotNil(t, cfg)
	factory, err := receiver.MakeFactoryMap(NewFactory())
	factories := otelcol.Factories{Receivers: factory}
	require.NoError(t, err)
	parsedFactory := factories.Receivers[metadata.Type]
	require.NotEmpty(t, parsedFactory)
	assert.EqualValues(t, parsedFactory.Type(), metadata.Type)
	assert.EqualValues(t, 3, parsedFactory.MetricsReceiverStability())
}
