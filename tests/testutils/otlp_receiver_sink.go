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
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// To be used as a builder whose Build() method provides the actual instance capable of starting the OTLP receiver
// providing received metrics to test cases.
type OTLPMetricsReceiverSink struct {
	receiver *component.MetricsReceiver
	sink     *consumertest.MetricsSink
	Endpoint string
	Logger   *zap.Logger
	Host     component.Host
}

func NewOTLPMetricsReceiverSink() OTLPMetricsReceiverSink {
	return OTLPMetricsReceiverSink{}
}

// Required
func (otlp OTLPMetricsReceiverSink) WithEndpoint(endpoint string) OTLPMetricsReceiverSink {
	otlp.Endpoint = endpoint
	return otlp
}

// If not set will use NopHost
func (otlp OTLPMetricsReceiverSink) WithHost(host component.Host) OTLPMetricsReceiverSink {
	otlp.Host = host
	return otlp
}

// If not set will use nop logger
func (otlp OTLPMetricsReceiverSink) WithLogger(logger *zap.Logger) OTLPMetricsReceiverSink {
	otlp.Logger = logger
	return otlp
}

// Will create, configure, and start an OTLPReceiver with GRPC listener and associated metric sink
func (otlp OTLPMetricsReceiverSink) Build() (*OTLPMetricsReceiverSink, error) {
	if otlp.Endpoint == "" {
		return nil, fmt.Errorf("must provide an Endpoint for OTLPMetricsReceiverSink")
	}
	if otlp.Logger == nil {
		otlp.Logger = zap.NewNop()
	}
	if otlp.Host == nil {
		otlp.Host = componenttest.NewNopHost()
	}

	otlp.sink = new(consumertest.MetricsSink)

	otlpFactory := otlpreceiver.NewFactory()
	otlpConfig := otlpFactory.CreateDefaultConfig().(*otlpreceiver.Config)
	otlpConfig.GRPC.NetAddr = confignet.NetAddr{Endpoint: otlp.Endpoint, Transport: "tcp"}
	otlpConfig.HTTP = nil

	params := component.ReceiverCreateSettings{TelemetrySettings: component.TelemetrySettings{Logger: otlp.Logger, TracerProvider: trace.NewNoopTracerProvider()}}
	receiver, err := otlpFactory.CreateMetricsReceiver(context.Background(), params, otlpConfig, otlp.sink)
	if err != nil {
		return nil, err
	}
	otlp.receiver = &receiver

	return &otlp, nil
}

func (otlp *OTLPMetricsReceiverSink) assertBuilt(operation string) error {
	if otlp.receiver == nil || otlp.sink == nil {
		return fmt.Errorf("cannot invoke %s() on an OTLPMetricsReceiverSink that hasn't been built", operation)
	}
	return nil
}

func (otlp *OTLPMetricsReceiverSink) Start() error {
	if err := otlp.assertBuilt("Start"); err != nil {
		return err
	}
	return (*otlp.receiver).Start(context.Background(), otlp.Host)
}

func (otlp *OTLPMetricsReceiverSink) Shutdown() error {
	if err := otlp.assertBuilt("Shutdown"); err != nil {
		return err
	}
	return (*otlp.receiver).Shutdown(context.Background())
}
func (otlp *OTLPMetricsReceiverSink) AllMetrics() []pdata.Metrics {
	if err := otlp.assertBuilt("AllMetrics"); err != nil {
		return nil
	}
	return otlp.sink.AllMetrics()
}

func (otlp *OTLPMetricsReceiverSink) DataPointCount() int {
	if err := otlp.assertBuilt("DataPointCount"); err != nil {
		return 0
	}
	return otlp.sink.DataPointCount()
}

func (otlp *OTLPMetricsReceiverSink) Reset() {
	if err := otlp.assertBuilt("Reset"); err == nil {
		otlp.sink.Reset()
	}
}

func (otlp *OTLPMetricsReceiverSink) AssertAllMetricsReceived(t *testing.T, expectedResourceMetrics ResourceMetrics, waitTime time.Duration) error {
	if err := otlp.assertBuilt("AssertAllMetricsReceived"); err != nil {
		return err
	}

	if len(expectedResourceMetrics.ResourceMetrics) == 0 {
		return fmt.Errorf("empty ResourceMetrics provided")
	}

	receivedMetrics := ResourceMetrics{}

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

		receivedResourceMetrics, e := PDataToResourceMetrics(receivedOTLPMetrics...)
		require.NoError(t, e)
		require.NotNil(t, receivedResourceMetrics)
		receivedMetrics = FlattenResourceMetrics(receivedMetrics, receivedResourceMetrics)

		var containsAll bool
		containsAll, err = receivedMetrics.ContainsAll(expectedResourceMetrics)
		return containsAll
	}, waitTime, 10*time.Millisecond, "Failed to receive expected metrics")

	//testify won't render exceptionally long errors, so leaving this here for easy debugging
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return err
}
