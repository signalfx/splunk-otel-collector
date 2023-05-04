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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/prometheusremotewritereceiver/internal/testdata"
)

func TestWriteEmpty(t *testing.T) {
	mc := make(chan<- pmetric.Metrics)
	mockReporter := newMockReporter()
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)
	parser := &PrometheusRemoteOtelParser{SfxGatewayCompatability: true}
	require.NoError(t, err)
	cfg := &ServerConfig{
		Path:     "/metrics",
		Reporter: mockReporter,
		Mc:       mc,
		HTTPServerSettings: confighttp.HTTPServerSettings{
			Endpoint: expectedEndpoint,
		},
		Parser: parser,
	}
	require.Equal(t, expectedEndpoint, cfg.Endpoint)
	timeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	remoteWriteServer, err := newPrometheusRemoteWriteServer(cfg)
	assert.NoError(t, err)
	require.NotNil(t, remoteWriteServer)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t.Logf("starting server...")
		require.NoError(t, remoteWriteServer.ListenAndServe())
		t.Logf("stopped server...")
		wg.Done()
	}()

	client, err := NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
		timeout,
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, client.SendWriteRequest(&prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{},
		Metadata:   []prompb.MetricMetadata{},
	}))

	require.NoError(t, mockReporter.WaitAllOnMetricsProcessedCalls(time.Second*5))
	require.NoError(t, remoteWriteServer.Shutdown(ctx))
	require.Eventually(t, func() bool { wg.Wait(); return true }, time.Second*2, 100*time.Millisecond)
}

func TestWriteMany(t *testing.T) {
	mc := make(chan<- pmetric.Metrics, 1000)
	mockReporter := newMockReporter()
	freePort, err := GetFreePort()
	require.NoError(t, err)
	expectedEndpoint := fmt.Sprintf("localhost:%d", freePort)
	parser := &PrometheusRemoteOtelParser{SfxGatewayCompatability: true}
	require.NoError(t, err)
	cfg := &ServerConfig{
		Path:     "/metrics",
		Reporter: mockReporter,
		Mc:       mc,
		HTTPServerSettings: confighttp.HTTPServerSettings{
			Endpoint: expectedEndpoint,
		},
		Parser: parser,
	}
	require.Equal(t, expectedEndpoint, cfg.Endpoint)
	timeout := time.Second * 1000
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	remoteWriteServer, err := newPrometheusRemoteWriteServer(cfg)
	assert.NoError(t, err)
	require.NotNil(t, remoteWriteServer)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t.Logf("starting server...")
		require.NoError(t, remoteWriteServer.ListenAndServe())
		t.Logf("stopped server...")
		wg.Done()
	}()

	client, err := NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
		timeout,
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	time.Sleep(100 * time.Millisecond)
	wqs := testdata.GetWriteRequestsOfAllTypesWithoutMetadata()
	for _, wq := range wqs {
		require.NoError(t, client.SendWriteRequest(wq))
	}

	require.NoError(t, mockReporter.WaitAllOnMetricsProcessedCalls(time.Second*5))
	require.NoError(t, remoteWriteServer.Shutdown(ctx))
	require.Eventually(t, func() bool { wg.Wait(); return true }, time.Second*2, 100*time.Millisecond)
}
