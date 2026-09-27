// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(metadata.Type, createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	controller := scraperhelper.NewDefaultControllerConfig()
	controller.CollectionInterval = 10 * time.Second
	controller.Timeout = 10 * time.Second
	return &Config{
		ControllerConfig:     controller,
		MetricsBuilderConfig: metadata.NewDefaultMetricsBuilderConfig(),
		SocketPath:           "/var/run/ptp/ptp4l",
		PMCPath:              "pmc",
	}
}

func createMetricsReceiver(_ context.Context, settings receiver.Settings, cfg component.Config, next consumer.Metrics) (receiver.Metrics, error) {
	config := cfg.(*Config)
	s, err := scraper.NewMetrics(newScraper(config, settings).scrape)
	if err != nil {
		return nil, err
	}
	return scraperhelper.NewMetricsController(&config.ControllerConfig, settings, next,
		scraperhelper.AddMetricsScraper(metadata.Type, s))
}
