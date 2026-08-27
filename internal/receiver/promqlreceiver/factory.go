package promqlreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("promql"),
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		ControllerConfig: scraperhelper.NewDefaultControllerConfig(),
	}
}

func createMetricsReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	rCfg := cfg.(*Config)
	return scraperhelper.NewMetricsController(&rCfg.ControllerConfig, settings, nextConsumer, scraperhelper.AddMetricsScraper(component.MustNewType("promql"), &scraper{queries: rCfg.Queries, cfg: rCfg, settings: settings}))
}
