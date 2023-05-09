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
)

func TestEmptySend(t *testing.T) {
	timeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)

	cfg.Endpoint = expectedEndpoint
	cfg.ListenPath = "/metrics"

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
	require.Eventually(t, func() bool { remoteWriteReceiver.server.ready(); return true }, time.Second*10, 50*time.Millisecond)

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

func TestSuccessfulSend(t *testing.T) {
	timeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)

	cfg.Endpoint = expectedEndpoint
	cfg.ListenPath = "/metrics"

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()

	sampleNoMdMetrics := GetWriteRequestsOfAllTypesWithoutMetadata()
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
	require.Eventually(t, func() bool { remoteWriteReceiver.server.ready(); return true }, time.Second*10, 50*time.Millisecond)

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
		assert.NoError(t, err, "failed to write %d", index)
		if nil != err {
			assert.NoError(t, errors.Unwrap(err))
		}
		// always will have 3 "health" metrics due to sfx gateway compatibility metrics
		assert.GreaterOrEqual(t, mockreporter.TotalSuccessMetrics.Load(), int32(len(wq.Timeseries)+3))
		assert.Equal(t, mockreporter.TotalErrorMetrics.Load(), int32(0))
	}

	require.NoError(t, remoteWriteReceiver.Shutdown(ctx))
}

func TestRealReporter(t *testing.T) {
	timeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)

	cfg.Endpoint = expectedEndpoint
	cfg.ListenPath = "/metrics"

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()

	sampleNoMdMetrics := GetWriteRequestsOfAllTypesWithoutMetadata()

	receiver, err := New(mockSettings, cfg, mockConsumer)
	remoteWriteReceiver := receiver.(*prometheusRemoteWriteReceiver)

	assert.NoError(t, err)
	require.NotNil(t, remoteWriteReceiver)
	require.NoError(t, remoteWriteReceiver.Start(ctx, nopHost))
	require.NotEmpty(t, remoteWriteReceiver.settings.TelemetrySettings)
	require.NotEmpty(t, remoteWriteReceiver.settings.Logger)
	require.NotEmpty(t, remoteWriteReceiver.settings.BuildInfo)
	require.Eventually(t, func() bool { remoteWriteReceiver.server.ready(); return true }, time.Second*10, 50*time.Millisecond)

	client, err := NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
		time.Second*5,
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	for index, wq := range sampleNoMdMetrics {
		err = client.SendWriteRequest(wq)
		require.NoError(t, err, "failed to write %d", index)
	}

	require.NoError(t, remoteWriteReceiver.Shutdown(ctx))
}
