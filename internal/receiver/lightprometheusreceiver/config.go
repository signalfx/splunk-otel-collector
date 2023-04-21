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
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/lightprometheusreceiver/internal/metadata"
)

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultScraperControllerSettings(metadata.Type)
	// set the default collection interval to 30 seconds which is half of the
	// lowest job frequency of 1 minute
	scs.CollectionInterval = time.Second * 30
	return &Config{
		ScraperControllerSettings: scs,
		HTTPClientSettings:        confighttp.NewDefaultHTTPClientSettings(),
	}
}

type PrometheusMetricFamily string

const (
	Counter   PrometheusMetricFamily = "counter"
	Gauge                            = "gauge"
	Summary                          = "summary"
	Histogram                        = "histogram"
	Untyped                          = "untyped"
)

type Config struct {
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	MetricFamiliesExcluded                  []PrometheusMetricFamily `mapstructure:"metric_families_excluded"`
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
}
