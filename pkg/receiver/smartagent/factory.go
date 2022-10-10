// Copyright OpenTelemetry Authors
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

package smartagentreceiver

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "smartagent"
)

var (
	// Smart Agent receivers can be for metrics or logs (events).
	// We keep store of them to ensure the same instance is used for a given config.
	receiverStoreLock = sync.Mutex{}
	receiverStore     = map[*Config]*receiver{}
)

func getOrCreateReceiver(cfg config.Receiver, params component.ReceiverCreateSettings) (*receiver, error) {
	receiverStoreLock.Lock()
	defer receiverStoreLock.Unlock()
	receiverConfig := cfg.(*Config)

	err := receiverConfig.validate()
	if err != nil {
		return nil, err
	}

	receiver, ok := receiverStore[receiverConfig]
	if !ok {
		receiver = newReceiver(params, *receiverConfig)
		receiverStore[receiverConfig] = receiver
	}

	return receiver, nil
}

func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		CreateDefaultConfig,
		component.WithMetricsReceiver(createMetricsReceiver, component.StabilityLevelBeta),
		component.WithLogsReceiver(createLogsReceiver, component.StabilityLevelBeta),
		component.WithTracesReceiver(createTracesReceiver, component.StabilityLevelBeta),
	)
}

func CreateDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
	}
}

func createMetricsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	metricsConsumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerMetricsConsumer(metricsConsumer)
	return receiver, nil
}

func createLogsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	logsConsumer consumer.Logs,
) (component.LogsReceiver, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerLogsConsumer(logsConsumer)
	return receiver, nil
}

func createTracesReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	tracesConsumer consumer.Traces,
) (component.TracesReceiver, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerTracesConsumer(tracesConsumer)
	return receiver, nil
}
