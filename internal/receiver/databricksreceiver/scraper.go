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

package databricksreceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"
)

const instanceNameAttr = "databricks.instance.name"

// scraper provides a scrape method to a scraper controller receiver. The scrape
// method is the entry point into this receiver's functionality, running on a
// timer, and building metrics from metrics providers.
type scraper struct {
	rmp          runMetricsProvider
	mp           metricsProvider
	instanceName string
}

func (s scraper) scrape(_ context.Context) (pdata.Metrics, error) {
	out := pdata.NewMetrics()
	rms := out.ResourceMetrics()
	rm := rms.AppendEmpty()
	rm.Resource().Attributes().Insert(
		instanceNameAttr,
		pdata.NewAttributeValueString(s.instanceName),
	)
	ilms := rm.InstrumentationLibraryMetrics()
	ilm := ilms.AppendEmpty()
	ms := ilm.Metrics()

	const errfmt = "scraper.scrape(): %w"
	var err error

	jobIDs, err := s.mp.addJobStatusMetrics(ms)
	if err != nil {
		return out, fmt.Errorf(errfmt, err)
	}

	err = s.mp.addNumActiveRunsMetric(ms)
	if err != nil {
		return out, fmt.Errorf(errfmt, err)
	}

	err = s.rmp.addMultiJobRunMetrics(ms, jobIDs)
	if err != nil {
		return out, fmt.Errorf(errfmt, err)
	}

	return out, err
}
