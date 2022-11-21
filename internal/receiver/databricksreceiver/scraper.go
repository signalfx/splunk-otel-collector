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
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

// scraper provides a scrape method to a scraper controller receiver. The scrape
// method is the entry point into this receiver's functionality, running on a
// timer, and building metrics from metrics providers.
type scraper struct {
	logger      *zap.Logger
	rmp         runMetricsProvider
	dbmp        dbMetricsProvider
	builder     *metadata.MetricsBuilder
	resourceOpt metadata.ResourceMetricsOption
	scmb        sparkCoreMetricsBuilder
	semb        sparkMetricsBuilder
}

func (s scraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	var err error

	now := pcommon.NewTimestampFromTime(time.Now())

	jobIDs, err := s.dbmp.addJobStatusMetrics(s.builder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("srape: error adding job status metrics: %w", err)
	}

	err = s.dbmp.addNumActiveRunsMetric(s.builder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: error adding num active runs metric %w", err)
	}

	err = s.rmp.addMultiJobRunMetrics(jobIDs, s.builder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: error adding multi job run metrics: %w", err)
	}

	histoMetrics, clusterIDs, err := s.scmb.buildCoreMetrics(s.builder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: error building core spark metrics: %w", err)
	}

	s.logger.Debug("found clusters", zap.Strings("cluster-ids", clusterIDs))

	err = s.semb.buildExecutorMetrics(s.builder, now, clusterIDs)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scraper.scrape(): %w", err)
	}

	out := s.builder.Emit(s.resourceOpt)
	scopeMetrics := out.ResourceMetrics().At(0).ScopeMetrics().At(0)
	if histoMetrics != nil {
		for _, histoMetric := range histoMetrics {
			histoMetric.CopyTo(scopeMetrics.Metrics().AppendEmpty())
		}
	}

	return out, nil
}
