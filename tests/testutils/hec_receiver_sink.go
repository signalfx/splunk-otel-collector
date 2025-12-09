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

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	splunkhectypeStr = "splunk_hec"
)

// To be used as a builder whose Build() method provides the actual instance capable of starting the HEC receiver
// providing received metrics to test cases.
type HECReceiverSink struct {
	Host         component.Host
	logsReceiver *receiver.Logs
	logsSink     *consumertest.LogsSink
	Logger       *zap.Logger
	Endpoint     string
}

func NewHECReceiverSink() HECReceiverSink {
	return HECReceiverSink{}
}

// WithEndpoint is required or Build() will fail
func (hec HECReceiverSink) WithEndpoint(endpoint string) HECReceiverSink {
	hec.Endpoint = endpoint
	return hec
}

// Build will create, configure, and start an HECReceiver with GRPC listener and associated metric and log sinks
func (hec HECReceiverSink) Build() (*HECReceiverSink, error) {
	if hec.Endpoint == "" {
		return nil, errors.New("must provide an Endpoint for HECReceiverSink")
	}
	hec.Logger = zap.NewNop()
	hec.Host = componenttest.NewNopHost()

	hec.logsSink = new(consumertest.LogsSink)

	hecFactory := splunkhecreceiver.NewFactory()
	hecConfig := hecFactory.CreateDefaultConfig().(*splunkhecreceiver.Config)
	hecConfig.Endpoint = hec.Endpoint

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

	return &hec, nil
}

func (hec *HECReceiverSink) assertBuilt(operation string) error {
	if hec.logsReceiver == nil || hec.logsSink == nil {
		return fmt.Errorf("cannot invoke %s() on an HECReceiverSink that hasn't been built", operation)
	}
	return nil
}

func (hec *HECReceiverSink) Start() error {
	if err := hec.assertBuilt("Start"); err != nil {
		return err
	}
	return (*hec.logsReceiver).Start(context.Background(), hec.Host)
}

func (hec *HECReceiverSink) Shutdown() error {
	if err := hec.assertBuilt("Shutdown"); err != nil {
		return err
	}
	return (*hec.logsReceiver).Shutdown(context.Background())
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
