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

package prometheusremotewritereceiver

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/prometheusremotewritereceiver/internal/testdata"
)

func TestEmptySend(t *testing.T) {
	timeout := time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)

	cfg.Endpoint = expectedEndpoint
	cfg.ListenPath = "/metrics"
	cfg.SfxGatewayCompatability = true

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()
	mockreporter := newMockReporter()
	receiver, err := New(mockSettings, cfg, mockConsumer)
	remoteWriteReceiver := receiver.(*prometheusRemoteWriteReceiver)
	remoteWriteReceiver.reporter = mockreporter

	assert.NoError(t, err)
	require.NotNil(t, remoteWriteReceiver)
	require.NoError(t, remoteWriteReceiver.Start(ctx, nopHost))
	require.NotEmpty(t, remoteWriteReceiver.server)
	require.NotEmpty(t, remoteWriteReceiver.cancel)
	require.NotEmpty(t, remoteWriteReceiver.config)
	require.Equal(t, remoteWriteReceiver.config.Endpoint, fmt.Sprintf("localhost:%d", freePort))
	require.NotEmpty(t, remoteWriteReceiver.settings)
	require.NotNil(t, remoteWriteReceiver.reporter)
	require.Equal(t, expectedEndpoint, remoteWriteReceiver.server.Addr)

	client, err := NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
		time.Second*5,
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.SendWriteRequest(&prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{},
		Metadata:   []prompb.MetricMetadata{},
	}))
	require.NoError(t, mockreporter.WaitAllOnMetricsProcessedCalls(10*time.Second))
	require.NoError(t, remoteWriteReceiver.Shutdown(ctx))
}

func TestActualSend(t *testing.T) {
	timeout := time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)

	cfg.Endpoint = expectedEndpoint
	cfg.ListenPath = "/metrics"
	cfg.SfxGatewayCompatability = true

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()

	sampleNoMdMetrics := testdata.GetWriteRequestsOfAllTypesWithoutMetadata()
	mockreporter := newMockReporter()

	receiver, err := New(mockSettings, cfg, mockConsumer)
	remoteWriteReceiver := receiver.(*prometheusRemoteWriteReceiver)
	remoteWriteReceiver.reporter = mockreporter

	assert.NoError(t, err)
	require.NotNil(t, remoteWriteReceiver)
	require.NoError(t, remoteWriteReceiver.Start(ctx, nopHost))
	require.NotEmpty(t, remoteWriteReceiver.server)
	require.NotEmpty(t, remoteWriteReceiver.cancel)
	require.NotEmpty(t, remoteWriteReceiver.config)
	require.Equal(t, remoteWriteReceiver.config.Endpoint, fmt.Sprintf("localhost:%d", freePort))
	require.NotEmpty(t, remoteWriteReceiver.settings)
	require.NotNil(t, remoteWriteReceiver.reporter)
	require.Equal(t, expectedEndpoint, remoteWriteReceiver.server.Addr)

	client, err := NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
		time.Second*5,
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	for index, wq := range sampleNoMdMetrics {
		mockreporter.AddExpectedStart(1)
		mockreporter.AddExpectedSuccess(1)
		err = client.SendWriteRequest(wq)
		assert.Nil(t, err, "failed to write %d", index)
		if nil != err {
			assert.Nil(t, errors.Unwrap(err))
		}
	}

	require.NoError(t, remoteWriteReceiver.Shutdown(ctx))
}
