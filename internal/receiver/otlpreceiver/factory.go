// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package otlpreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/receiver"
	upstreamotlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/xreceiver"
)

const typeStr = "otlp"

var upstreamFactory = upstreamotlpreceiver.NewFactory()

// NewFactory creates a new OTLP receiver factory.
func NewFactory() receiver.Factory {
	upstreamProfilesFactory := upstreamFactory.(xreceiver.Factory)
	return xreceiver.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		xreceiver.WithTraces(createTraces, upstreamFactory.TracesStability()),
		xreceiver.WithMetrics(createMetrics, upstreamFactory.MetricsStability()),
		xreceiver.WithLogs(createLogs, upstreamFactory.LogsStability()),
		xreceiver.WithProfiles(createProfiles, upstreamProfilesFactory.ProfilesStability()),
	)
}

func createDefaultConfig() component.Config {
	upstreamConfig := upstreamFactory.CreateDefaultConfig().(*upstreamotlpreceiver.Config)
	return &Config{Config: *upstreamConfig}
}

func createTraces(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (receiver.Traces, error) {
	oCfg := cfg.(*Config)
	r, err := upstreamFactory.CreateTraces(ctx, set, &oCfg.Config, nextConsumer)
	if err != nil {
		return nil, err
	}
	return receivers.LoadOrStore(oCfg, set, r), nil
}

func createMetrics(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	oCfg := cfg.(*Config)
	r, err := upstreamFactory.CreateMetrics(ctx, set, &oCfg.Config, nextConsumer)
	if err != nil {
		return nil, err
	}
	return receivers.LoadOrStore(oCfg, set, r), nil
}

func createLogs(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (receiver.Logs, error) {
	oCfg := cfg.(*Config)
	r, err := upstreamFactory.CreateLogs(ctx, set, &oCfg.Config, nextConsumer)
	if err != nil {
		return nil, err
	}
	return receivers.LoadOrStore(oCfg, set, r), nil
}

func createProfiles(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer xconsumer.Profiles,
) (xreceiver.Profiles, error) {
	oCfg := cfg.(*Config)
	r, err := upstreamFactory.(xreceiver.Factory).CreateProfiles(ctx, set, &oCfg.Config, nextConsumer)
	if err != nil {
		return nil, err
	}
	return receivers.LoadOrStore(oCfg, set, r), nil
}
