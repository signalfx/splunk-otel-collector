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

package discoveryreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/common/sharedcomponent"
)

const (
	typeStr = "discovery"
)

// This is the map of already created discovery receivers for particular configurations.
// We maintain this map because the Factory is asked log and metric receivers separately
// when it gets CreateLogs() and CreateMetrics() but they must not
// create separate objects, they must use one receiver object per configuration.
var receivers = sharedcomponent.NewSharedComponents()

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelDevelopment),
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment))
}

func createDefaultConfig() component.Config {
	return &Config{
		EmbedReceiverConfig: false,
		CorrelationTTL:      10 * time.Minute,
	}
}

func createLogsReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	var err error
	r := receivers.GetOrAdd(
		cfg, func() component.Component {
			var rcv component.Component
			rcv, err = newDiscoveryReceiver(settings, cfg.(*Config))
			return rcv
		},
	)
	if err != nil {
		return nil, err
	}
	r.Unwrap().(*discoveryReceiver).nextLogsConsumer = consumer
	return r, nil
}

func createMetricsReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	var err error
	r := receivers.GetOrAdd(
		cfg, func() component.Component {
			var rcv component.Component
			rcv, err = newDiscoveryReceiver(settings, cfg.(*Config))
			return rcv
		},
	)
	if err != nil {
		return nil, err
	}
	r.Unwrap().(*discoveryReceiver).nextMetricsConsumer = consumer
	return r, nil
}
