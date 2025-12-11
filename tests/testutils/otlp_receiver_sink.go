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
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	typeStr = "otlp"
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

// Build will create, configure, and start an OTLPReceiver with GRPC listener and associated metric and log sinks
func (otlp OTLPReceiverSink) Build() (*OTLPReceiverSink, error) {
	if otlp.Endpoint == "" {
		return nil, errors.New("must provide an Endpoint for OTLPReceiverSink")
	}
	otlp.Logger = zap.NewNop()
	otlp.Host = componenttest.NewNopHost()

	otlp.logsSink = new(consumertest.LogsSink)
	otlp.metricsSink = new(consumertest.MetricsSink)
	otlp.tracesSink = new(consumertest.TracesSink)

	otlpFactory := otlpreceiver.NewFactory()
	otlpConfig := otlpFactory.CreateDefaultConfig().(*otlpreceiver.Config)
	otlpConfig.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  otlp.Endpoint,
			Transport: "tcp",
		},
	})
	otlpConfig.HTTP = configoptional.None[otlpreceiver.HTTPConfig]()

	params := receiver.Settings{
		ID: component.MustNewID(typeStr),
		TelemetrySettings: component.TelemetrySettings{
			Logger:         otlp.Logger,
			TracerProvider: tnoop.NewTracerProvider(),
			MeterProvider:  mnoop.NewMeterProvider(),
		},
	}

	logsReceiver, err := otlpFactory.CreateLogs(context.Background(), params, otlpConfig, otlp.logsSink)
	if err != nil {
		return nil, err
	}
	otlp.logsReceiver = &logsReceiver

	metricsReceiver, err := otlpFactory.CreateMetrics(context.Background(), params, otlpConfig, otlp.metricsSink)
	if err != nil {
		return nil, err
	}
	otlp.metricsReceiver = &metricsReceiver

	tracesReceiver, err := otlpFactory.CreateTraces(context.Background(), params, otlpConfig, otlp.tracesSink)
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
