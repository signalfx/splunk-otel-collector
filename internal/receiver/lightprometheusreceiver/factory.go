// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lightprometheusreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	scraperpkg "go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

const typeStr = "lightprometheus"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDeprecated),
	)
}

// createMetricsReceiver creates a metrics receiver for scraping Prometheus metrics.
func createMetricsReceiver(
	_ context.Context,
	params receiver.Settings,
	rConf component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	c, _ := rConf.(*Config)
	s := newScraper(params, c)

	scraper, err := scraperpkg.NewMetrics(s.scrape, scraperpkg.WithStart(s.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(
		&c.ControllerConfig,
		params,
		consumer,
		scraperhelper.AddMetricsScraper(component.MustNewType(typeStr), scraper),
	)
}
