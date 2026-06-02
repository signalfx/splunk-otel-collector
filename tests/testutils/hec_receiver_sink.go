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

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	splunkhectypeStr = "splunk_hec"
)

// To be used as a builder whose Build() method provides the actual instance capable of starting the HEC receiver
// providing received logs and metrics to test cases.
type HECReceiverSink struct {
	Host            component.Host
	logsReceiver    *receiver.Logs
	logsSink        *consumertest.LogsSink
	metricsReceiver *receiver.Metrics
	metricsSink     *consumertest.MetricsSink
	Logger          *zap.Logger
	Endpoint        string
}

func NewHECReceiverSink() HECReceiverSink {
	return HECReceiverSink{}
}

// WithEndpoint is required or Build() will fail
func (hec HECReceiverSink) WithEndpoint(endpoint string) HECReceiverSink {
	hec.Endpoint = endpoint
	return hec
}

// Build will create and configure an HECReceiver with associated log and metric sinks.
func (hec HECReceiverSink) Build() (*HECReceiverSink, error) {
	if hec.Endpoint == "" {
		return nil, errors.New("must provide an Endpoint for HECReceiverSink")
	}
	hec.Logger = zap.NewNop()
	hec.Host = componenttest.NewNopHost()

	hec.logsSink = new(consumertest.LogsSink)
	hec.metricsSink = new(consumertest.MetricsSink)

	hecFactory := splunkhecreceiver.NewFactory()
	hecConfig := hecFactory.CreateDefaultConfig().(*splunkhecreceiver.Config)
	hecConfig.NetAddr.Endpoint = hec.Endpoint

	params := receiver.Settings{
		ID: component.MustNewID(splunkhectypeStr),
		TelemetrySettings: component.TelemetrySettings{
			Logger:         hec.Logger,
			TracerProvider: tnoop.NewTracerProvider(),
			MeterProvider:  mnoop.NewMeterProvider(),
		},
	}

	logsReceiver, err := hecFactory.CreateLogs(context.Background(), params, hecConfig, hec.logsSink)
	if err != nil {
		return nil, err
	}
	hec.logsReceiver = &logsReceiver

	metricsReceiver, err := hecFactory.CreateMetrics(context.Background(), params, hecConfig, hec.metricsSink)
	if err != nil {
		return nil, err
	}
	hec.metricsReceiver = &metricsReceiver

	return &hec, nil
}

func (hec *HECReceiverSink) assertBuilt(operation string) error {
	if hec.logsReceiver == nil || hec.logsSink == nil || hec.metricsReceiver == nil || hec.metricsSink == nil {
		return fmt.Errorf("cannot invoke %s() on an HECReceiverSink that hasn't been built", operation)
	}
	return nil
}

func (hec *HECReceiverSink) Start() error {
	if err := hec.assertBuilt("Start"); err != nil {
		return err
	}
	logsErr := (*hec.logsReceiver).Start(context.Background(), hec.Host)
	metricsErr := (*hec.metricsReceiver).Start(context.Background(), hec.Host)
	if logsErr != nil || metricsErr != nil {
		return errors.Join(logsErr, metricsErr)
	}
	return nil
}

func (hec *HECReceiverSink) Shutdown() error {
	if err := hec.assertBuilt("Shutdown"); err != nil {
		return err
	}
	logsErr := (*hec.logsReceiver).Shutdown(context.Background())
	metricsErr := (*hec.metricsReceiver).Shutdown(context.Background())
	return errors.Join(logsErr, metricsErr)
}

func (hec *HECReceiverSink) LogRecordCount() int {
	if err := hec.assertBuilt("LogRecordCount"); err != nil {
		return 0
	}
	return hec.logsSink.LogRecordCount()
}

func (hec *HECReceiverSink) AllLogs() []plog.Logs {
	if err := hec.assertBuilt("AllLogs"); err != nil {
		return nil
	}
	return hec.logsSink.AllLogs()
}

func (hec *HECReceiverSink) AllMetrics() []pmetric.Metrics {
	if err := hec.assertBuilt("AllMetrics"); err != nil {
		return nil
	}
	return hec.metricsSink.AllMetrics()
}

func (hec *HECReceiverSink) DataPointCount() int {
	if err := hec.assertBuilt("DataPointCount"); err != nil {
		return 0
	}
	return hec.metricsSink.DataPointCount()
}
