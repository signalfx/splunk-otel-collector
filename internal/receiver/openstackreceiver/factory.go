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

package openstackreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	scraperpkg "go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

const typeStr = "openstack"

// NewFactory creates a new factory for the OpenStack receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	cs := scraperhelper.NewDefaultControllerConfig()
	cs.CollectionInterval = 10 * time.Second
	return &Config{
		ControllerConfig:       cs,
		ProjectDomainID:        "default",
		UserDomainID:           "default",
		HTTPTimeout:            30 * time.Second,
		QueryHypervisorMetrics: true,
	}
}

func createMetricsReceiver(
	_ context.Context,
	params receiver.Settings,
	rConf component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errInvalidConfig
	}

	s := newScraper(params.Logger, cfg)

	scraper, err := scraperpkg.NewMetrics(
		s.scrape,
		scraperpkg.WithStart(s.start),
		scraperpkg.WithShutdown(s.shutdown),
	)
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(
		&cfg.ControllerConfig,
		params,
		consumer,
		scraperhelper.AddMetricsScraper(component.MustNewType(typeStr), scraper),
	)
}
