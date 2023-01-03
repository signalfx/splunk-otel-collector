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
	"go.opentelemetry.io/collector/consumer"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
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

func getOrCreateReceiver(cfg component.Config, params otelcolreceiver.CreateSettings) (*receiver, error) {
	receiverStoreLock.Lock()
	defer receiverStoreLock.Unlock()
	receiverConfig := cfg.(*Config)

	err := receiverConfig.validate()
	if err != nil {
		return nil, err
	}

	receiverInst, ok := receiverStore[receiverConfig]
	if !ok {
		receiverInst = newReceiver(params, *receiverConfig)
		receiverStore[receiverConfig] = receiverInst
	}

	return receiverInst, nil
}

func NewFactory() otelcolreceiver.Factory {
	return otelcolreceiver.NewFactory(
		typeStr,
		CreateDefaultConfig,
		otelcolreceiver.WithMetrics(createMetricsReceiver, component.StabilityLevelBeta),
		otelcolreceiver.WithLogs(createLogsReceiver, component.StabilityLevelBeta),
		otelcolreceiver.WithTraces(createTracesReceiver, component.StabilityLevelBeta),
	)
}

func CreateDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsReceiver(
	_ context.Context,
	params otelcolreceiver.CreateSettings,
	cfg component.Config,
	metricsConsumer consumer.Metrics,
) (otelcolreceiver.Metrics, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerMetricsConsumer(metricsConsumer)
	return receiver, nil
}

func createLogsReceiver(
	_ context.Context,
	params otelcolreceiver.CreateSettings,
	cfg component.Config,
	logsConsumer consumer.Logs,
) (otelcolreceiver.Logs, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerLogsConsumer(logsConsumer)
	return receiver, nil
}

func createTracesReceiver(
	_ context.Context,
	params otelcolreceiver.CreateSettings,
	cfg component.Config,
	tracesConsumer consumer.Traces,
) (otelcolreceiver.Traces, error) {
	receiver, err := getOrCreateReceiver(cfg, params)
	if err != nil {
		return nil, err
	}

	receiver.registerTracesConsumer(tracesConsumer)
	return receiver, nil
}
