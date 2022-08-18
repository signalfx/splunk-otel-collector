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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "discovery"
)

func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithLogsReceiver(createLogsReceiver, component.StabilityLevelInDevelopment))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings:    config.NewReceiverSettings(config.NewComponentID(typeStr)),
		LogEndpoints:        false,
		EmbedReceiverConfig: false,
	}
}

func createLogsReceiver(
	_ context.Context,
	settings component.ReceiverCreateSettings,
	cfg config.Receiver,
	consumer consumer.Logs,
) (component.LogsReceiver, error) {
	dCfg := cfg.(*Config)
	if err := dCfg.Validate(); err != nil {
		return nil, err
	}
	return newDiscoveryReceiver(settings, dCfg, consumer)
}
