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

package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

// To be used as a builder whose Build() method provides the actual instance capable of starting the OTLP receiver
// providing received metrics to test cases.
type OTLPReceiverSink struct {
	Host            component.Host
	logsReceiver    *receiver.Logs
	logsSink        *consumertest.LogsSink
	metricsReceiver *receiver.Metrics
	metricsSink     *consumertest.MetricsSink
	tracesReceiver  *receiver.Traces
	tracesSink      *consumertest.TracesSink
	Logger          *zap.Logger
	Endpoint        string
}

func NewOTLPReceiverSink() OTLPReceiverSink {
	return OTLPReceiverSink{}
}

// WithEndpoint is required or Build() will fail
func (otlp OTLPReceiverSink) WithEndpoint(endpoint string) OTLPReceiverSink {
	otlp.Endpoint = endpoint
	return otlp
}

// WithHost is optional: if not set will use NopHost
func (otlp OTLPReceiverSink) WithHost(host component.Host) OTLPReceiverSink {
	otlp.Host = host
	return otlp
}

// WithLogger is optional: if not set will use nop logger
func (otlp OTLPReceiverSink) WithLogger(logger *zap.Logger) OTLPReceiverSink {
	otlp.Logger = logger
	return otlp
}

// Build will create, configure, and start an OTLPReceiver with GRPC listener and associated metric and log sinks
func (otlp OTLPReceiverSink) Build() (*OTLPReceiverSink, error) {
	if otlp.Endpoint == "" {
		return nil, fmt.Errorf("must provide an Endpoint for OTLPReceiverSink")
	}
	if otlp.Logger == nil {
		otlp.Logger = zap.NewNop()
	}
	if otlp.Host == nil {
		otlp.Host = componenttest.NewNopHost()
	}

	otlp.logsSink = new(consumertest.LogsSink)
	otlp.metricsSink = new(consumertest.MetricsSink)
	otlp.tracesSink = new(consumertest.TracesSink)

	otlpFactory := otlpreceiver.NewFactory()
	otlpConfig := otlpFactory.CreateDefaultConfig().(*otlpreceiver.Config)
	otlpConfig.GRPC.NetAddr = confignet.NetAddr{Endpoint: otlp.Endpoint, Transport: "tcp"}
	otlpConfig.HTTP = nil

	params := receiver.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         otlp.Logger,
			MeterProvider:  noop.NewMeterProvider(),
			TracerProvider: trace.NewNoopTracerProvider(),
		},
	}

	logsReceiver, err := otlpFactory.CreateLogsReceiver(context.Background(), params, otlpConfig, otlp.logsSink)
	if err != nil {
		return nil, err
	}
	otlp.logsReceiver = &logsReceiver

	metricsReceiver, err := otlpFactory.CreateMetricsReceiver(context.Background(), params, otlpConfig, otlp.metricsSink)
	if err != nil {
		return nil, err
	}
	otlp.metricsReceiver = &metricsReceiver

	tracesReceiver, err := otlpFactory.CreateTracesReceiver(context.Background(), params, otlpConfig, otlp.tracesSink)
	if err != nil {
		return nil, err
	}
	otlp.tracesReceiver = &tracesReceiver

	return &otlp, nil
}

func (otlp *OTLPReceiverSink) assertBuilt(operation string) error {
	if otlp.logsReceiver == nil || otlp.logsSink == nil ||
		otlp.metricsReceiver == nil || otlp.metricsSink == nil ||
		otlp.tracesReceiver == nil || otlp.tracesSink == nil {
		return fmt.Errorf("cannot invoke %s() on an OTLPReceiverSink that hasn't been built", operation)
	}
	return nil
}

func (otlp *OTLPReceiverSink) Start() error {
	if err := otlp.assertBuilt("Start"); err != nil {
		return err
	}

	return (*otlp.metricsReceiver).Start(context.Background(), otlp.Host)
}

func (otlp *OTLPReceiverSink) Shutdown() error {
	if err := otlp.assertBuilt("Shutdown"); err != nil {
		return err
	}

	return (*otlp.metricsReceiver).Shutdown(context.Background())
}

func (otlp *OTLPReceiverSink) AllLogs() []plog.Logs {
	if err := otlp.assertBuilt("AllLogs"); err != nil {
		return nil
	}
	return otlp.logsSink.AllLogs()
}

func (otlp *OTLPReceiverSink) AllMetrics() []pmetric.Metrics {
	if err := otlp.assertBuilt("AllMetrics"); err != nil {
		return nil
	}
	return otlp.metricsSink.AllMetrics()
}

func (otlp *OTLPReceiverSink) AllTraces() []ptrace.Traces {
	if err := otlp.assertBuilt("AllTraces"); err != nil {
		return nil
	}
	return otlp.tracesSink.AllTraces()
}

func (otlp *OTLPReceiverSink) DataPointCount() int {
	if err := otlp.assertBuilt("DataPointCount"); err != nil {
		return 0
	}
	return otlp.metricsSink.DataPointCount()
}

func (otlp *OTLPReceiverSink) LogRecordCount() int {
	if err := otlp.assertBuilt("LogRecordCount"); err != nil {
		return 0
	}
	return otlp.logsSink.LogRecordCount()
}

func (otlp *OTLPReceiverSink) SpanCount() int {
	if err := otlp.assertBuilt("SpanCount"); err != nil {
		return 0
	}
	return otlp.tracesSink.SpanCount()
}

func (otlp *OTLPReceiverSink) Reset() {
	if err := otlp.assertBuilt("Reset"); err == nil {
		otlp.metricsSink.Reset()
		otlp.logsSink.Reset()
		otlp.tracesSink.Reset()
	}
}

func (otlp *OTLPReceiverSink) AssertAllLogsReceived(t testing.TB, expectedResourceLogs telemetry.ResourceLogs, waitTime time.Duration) error {
	if err := otlp.assertBuilt("AssertAllLogsReceived"); err != nil {
		return err
	}

	if len(expectedResourceLogs.ResourceLogs) == 0 {
		return fmt.Errorf("empty ResourceLogs provided")
	}

	receivedLogs := telemetry.ResourceLogs{}

	var err error
	assert.Eventually(t, func() bool {
		if otlp.LogRecordCount() == 0 {
			if err == nil {
				err = fmt.Errorf("no logs received")
			}
			return false
		}
		receivedOTLPLogs := otlp.AllLogs()
		otlp.Reset()

		receivedResourceLogs, e := telemetry.PDataToResourceLogs(receivedOTLPLogs...)
		require.NoError(t, e)
		require.NotNil(t, receivedResourceLogs)
		receivedLogs = telemetry.FlattenResourceLogs(receivedLogs, receivedResourceLogs)

		var containsAll bool
		containsAll, err = receivedLogs.ContainsAll(expectedResourceLogs)
		return containsAll
	}, waitTime, 10*time.Millisecond, "Failed to receive expected logs")

	// testify won't render exceptionally long errors, so leaving this here for easy debugging
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return err
}

func (otlp *OTLPReceiverSink) AssertAllMetricsReceived(t testing.TB, expectedResourceMetrics telemetry.ResourceMetrics, waitTime time.Duration) error {
	if err := otlp.assertBuilt("AssertAllMetricsReceived"); err != nil {
		return err
	}

	if len(expectedResourceMetrics.ResourceMetrics) == 0 {
		return fmt.Errorf("empty ResourceMetrics provided")
	}

	receivedMetrics := telemetry.ResourceMetrics{}

	var err error
	assert.Eventually(t, func() bool {
		if otlp.DataPointCount() == 0 {
			if err == nil {
				err = fmt.Errorf("no metrics received")
			}
			return false
		}
		receivedOTLPMetrics := otlp.AllMetrics()
		otlp.Reset()

		receivedResourceMetrics, e := telemetry.PDataToResourceMetrics(receivedOTLPMetrics...)
		require.NoError(t, e)
		require.NotNil(t, receivedResourceMetrics)
		receivedMetrics = telemetry.FlattenResourceMetrics(receivedMetrics, receivedResourceMetrics)

		var containsOnly bool
		containsOnly, err = receivedMetrics.ContainsOnly(expectedResourceMetrics)
		return containsOnly
	}, waitTime, 10*time.Millisecond, "Failed to receive expected metrics")

	// testify won't render exceptionally long errors, so leaving this here for easy debugging
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return err
}

func (otlp *OTLPReceiverSink) AssertAllTracesReceived(t testing.TB, expectedResourceTraces telemetry.ResourceTraces, waitTime time.Duration) error {
	if err := otlp.assertBuilt("AssertAllTracesReceived"); err != nil {
		return err
	}

	if len(expectedResourceTraces.ResourceSpans) == 0 {
		return fmt.Errorf("empty ResourceTraces provided")
	}

	receivedTraces := telemetry.ResourceTraces{}

	var err error
	assert.Eventually(t, func() bool {
		if otlp.SpanCount() == 0 {
			if err == nil {
				err = fmt.Errorf("no traces received")
			}
			return false
		}
		receivedOTLPTraces := otlp.AllTraces()
		otlp.Reset()

		receivedResourceTraces, e := telemetry.PDataToResourceTraces(receivedOTLPTraces...)
		require.NoError(t, e)
		require.NotNil(t, receivedResourceTraces)
		receivedTraces = telemetry.FlattenResourceTraces(receivedResourceTraces)

		var containsAll bool
		containsAll, err = receivedTraces.ContainsAll(expectedResourceTraces)
		return containsAll
	}, waitTime, 10*time.Millisecond, "Failed to receive expected traces")

	// testify won't render exceptionally long errors, so leaving this here for easy debugging
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return err
}
