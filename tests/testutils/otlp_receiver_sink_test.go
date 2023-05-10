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

//go:build testutils

package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	otelcolexporter "go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

func TestNewOTLPReceiverSink(t *testing.T) {
	otlp := NewOTLPReceiverSink()
	require.NotNil(t, otlp)

	require.Empty(t, otlp.Endpoint)
	require.Nil(t, otlp.Host)
	require.Nil(t, otlp.Logger)
	require.Nil(t, otlp.logsReceiver)
	require.Nil(t, otlp.logsSink)
	require.Nil(t, otlp.metricsReceiver)
	require.Nil(t, otlp.metricsSink)
	require.Nil(t, otlp.tracesReceiver)
	require.Nil(t, otlp.tracesSink)
}

func TestBuilderMethods(t *testing.T) {
	otlp := NewOTLPReceiverSink()

	withEndpoint := otlp.WithEndpoint("myendpoint")
	require.Equal(t, "myendpoint", withEndpoint.Endpoint)
	require.Empty(t, otlp.Endpoint)

	host := componenttest.NewNopHost()
	withHost := otlp.WithHost(host)
	require.Same(t, host, withHost.Host)
	require.Nil(t, otlp.Host)

	logger := zap.NewNop()
	withLogger := otlp.WithLogger(logger)
	require.Same(t, logger, withLogger.Logger)
	require.Nil(t, otlp.Logger)
}

func TestBuildDefaults(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().Build()
	require.Error(t, err)
	assert.EqualError(t, err, "must provide an Endpoint for OTLPReceiverSink")
	assert.Nil(t, otlp)

	otlp, err = NewOTLPReceiverSink().WithEndpoint("myEndpoint").Build()
	require.NoError(t, err)
	assert.Equal(t, "myEndpoint", otlp.Endpoint)
	assert.NotNil(t, otlp.Host)
	assert.NotNil(t, otlp.Logger)
	assert.NotNil(t, otlp.logsReceiver)
	assert.NotNil(t, otlp.logsSink)
	assert.NotNil(t, otlp.metricsReceiver)
	assert.NotNil(t, otlp.metricsSink)
	assert.NotNil(t, otlp.tracesReceiver)
	assert.NotNil(t, otlp.tracesSink)
}

func TestReceiverMethodsWithoutBuildingDisallowed(t *testing.T) {
	otlp := NewOTLPReceiverSink()

	err := otlp.Start()
	require.Error(t, err)
	require.EqualError(t, err, "cannot invoke Start() on an OTLPReceiverSink that hasn't been built")

	err = otlp.Shutdown()
	require.Error(t, err)
	require.EqualError(t, err, "cannot invoke Shutdown() on an OTLPReceiverSink that hasn't been built")

	metrics := otlp.AllMetrics()
	require.Nil(t, metrics)

	dataPointCount := otlp.DataPointCount()
	require.Zero(t, dataPointCount)

	traces := otlp.AllTraces()
	require.Nil(t, traces)

	spanCount := otlp.SpanCount()
	require.Zero(t, spanCount)

	// doesn't panic
	otlp.Reset()

	err = otlp.AssertAllMetricsReceived(t, telemetry.ResourceMetrics{}, 0)
	require.Error(t, err)
	require.EqualError(t, err, "cannot invoke AssertAllMetricsReceived() on an OTLPReceiverSink that hasn't been built")

	err = otlp.AssertAllTracesReceived(t, telemetry.ResourceTraces{}, 0)
	require.Error(t, err)
	require.EqualError(t, err, "cannot invoke AssertAllTracesReceived() on an OTLPReceiverSink that hasn't been built")
}

func createOTLPFactoryParameters() (otlpexporter.Config, otelcolexporter.CreateSettings) {
	exporterCfg := otlpexporter.Config{
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Endpoint: "localhost:4317",
			TLSSetting: configtls.TLSClientSetting{
				Insecure: true,
			},
		},
	}
	createParams := otelcolexporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.NewNop(),
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  noop.NewMeterProvider(),
		},
	}

	return exporterCfg, createParams
}

func otlpMetricsExporter(t *testing.T) otelcolexporter.Metrics {
	exporterCfg, createParams := createOTLPFactoryParameters()
	otlpExporterFactory := otlpexporter.NewFactory()
	ctx := context.Background()

	exporter, err := otlpExporterFactory.CreateMetricsExporter(ctx, createParams, &exporterCfg)

	require.NoError(t, err)
	require.NotNil(t, exporter)
	err = exporter.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	return exporter
}

func TestOTLPReceiverMetricsAvailableToSink(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().WithEndpoint("localhost:4317").Build()
	require.NoError(t, err)

	err = otlp.Start()
	defer func() {
		require.NoError(t, otlp.Shutdown())
	}()
	require.NoError(t, err)

	exporter := otlpMetricsExporter(t)
	defer func() { require.NoError(t, exporter.Shutdown(context.Background())) }()

	metrics := telemetry.PDataMetrics()
	expectedCount := metrics.DataPointCount()
	err = exporter.ConsumeMetrics(context.Background(), metrics)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return otlp.DataPointCount() == expectedCount
	}, 5*time.Second, 1*time.Millisecond)
}

func TestAssertAllMetricsReceivedHappyPath(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().WithEndpoint("localhost:4317").Build()
	require.NoError(t, err)

	err = otlp.Start()
	defer func() {
		require.NoError(t, otlp.Shutdown())
	}()
	require.NoError(t, err)

	exporter := otlpMetricsExporter(t)
	defer func() { require.NoError(t, exporter.Shutdown(context.Background())) }()

	metrics := telemetry.PDataMetrics()
	err = exporter.ConsumeMetrics(context.Background(), metrics)
	require.NoError(t, err)

	resourceMetrics, err := telemetry.PDataToResourceMetrics(metrics)
	resourceMetrics = telemetry.FlattenResourceMetrics(resourceMetrics)
	require.NoError(t, err)
	require.NoError(t, otlp.AssertAllMetricsReceived(t, resourceMetrics, 100*time.Millisecond))
}

func otlpTracesExporter(t *testing.T) otelcolexporter.Traces {
	exporterCfg, createParams := createOTLPFactoryParameters()
	otlpExporterFactory := otlpexporter.NewFactory()
	ctx := context.Background()

	exporter, err := otlpExporterFactory.CreateTracesExporter(ctx, createParams, &exporterCfg)

	require.NoError(t, err)
	require.NotNil(t, exporter)
	err = exporter.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	return exporter
}

func TestOTLPReceiverTracesAvailableToSink(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().WithEndpoint("localhost:4317").Build()
	require.NoError(t, err)

	err = otlp.Start()
	defer func() {
		require.NoError(t, otlp.Shutdown())
	}()
	require.NoError(t, err)

	exporter := otlpTracesExporter(t)
	defer func() { require.NoError(t, exporter.Shutdown(context.Background())) }()

	traces := telemetry.PDataTraces()
	expectedCount := traces.SpanCount()
	err = exporter.ConsumeTraces(context.Background(), traces)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return otlp.SpanCount() == expectedCount
	}, 5*time.Second, 1*time.Millisecond)
}

func TestAssertAllTracesReceivedHappyPath(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().WithEndpoint("localhost:4317").Build()
	require.NoError(t, err)

	err = otlp.Start()
	defer func() {
		require.NoError(t, otlp.Shutdown())
	}()
	require.NoError(t, err)

	exporter := otlpTracesExporter(t)
	defer func() { require.NoError(t, exporter.Shutdown(context.Background())) }()

	traces := telemetry.PDataTraces()
	err = exporter.ConsumeTraces(context.Background(), traces)
	require.NoError(t, err)

	resourceTraces, err := telemetry.PDataToResourceTraces(traces)
	resourceTraces = telemetry.FlattenResourceTraces(resourceTraces)
	require.NoError(t, err)
	require.NoError(t, otlp.AssertAllTracesReceived(t, resourceTraces, 100*time.Millisecond))
}
