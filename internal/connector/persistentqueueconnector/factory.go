// Copyright Splunk, Inc.
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

package persistentqueueconnector

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/connector/xconnector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
)

func NewFactory() connector.Factory {
	return xconnector.NewFactory(
		component.MustNewType("persistent_queue"),
		createDefaultConfig,
		xconnector.WithLogsToLogs(createLogs, component.StabilityLevelDevelopment),
		xconnector.WithMetricsToMetrics(createMetrics, component.StabilityLevelDevelopment),
		xconnector.WithTracesToTraces(createTraces, component.StabilityLevelDevelopment),
		xconnector.WithProfilesToProfiles(createProfiles, component.StabilityLevelDevelopment),
	)
}

func createProfiles(_ context.Context, settings connector.Settings, config component.Config, profiles xconsumer.Profiles) (xconnector.Profiles, error) {
	cfg := config.(*Config)
	return &persistentqueue{
		settings:     settings,
		config:       cfg,
		nextProfiles: profiles,
	}, nil
}

func createMetrics(_ context.Context, settings connector.Settings, config component.Config, metrics consumer.Metrics) (connector.Metrics, error) {
	cfg := config.(*Config)
	return &persistentqueue{
		settings:    settings,
		config:      cfg,
		nextMetrics: metrics,
	}, nil
}

func createLogs(_ context.Context, settings connector.Settings, config component.Config, logs consumer.Logs) (connector.Logs, error) {
	cfg := config.(*Config)
	return &persistentqueue{
		settings: settings,
		config:   cfg,
		nextLogs: logs,
	}, nil
}

func createTraces(_ context.Context, settings connector.Settings, config component.Config, traces consumer.Traces) (connector.Traces, error) {
	cfg := config.(*Config)
	return &persistentqueue{
		settings:   settings,
		config:     cfg,
		nextTraces: traces,
	}, nil
}

func createDefaultConfig() component.Config {
	return &Config{
		ThroughputLimit: 10000,
	}
}
