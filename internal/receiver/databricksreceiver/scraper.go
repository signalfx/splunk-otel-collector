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
// timer.
type scraper struct {
	rmp             runMetricsProvider
	semb            sparkExtraMetricsBuilder
	dbrmp           dbrMetricsProvider
	scmb            sparkClusterMetricsBuilder
	dbrsvc          databricksService
	logger          *zap.Logger
	metricsBuilder  *metadata.MetricsBuilder
	dbrInstanceName string
}

func (s scraper) scrape(_ context.Context) (pmetric.Metrics, error) {
	now := pcommon.NewTimestampFromTime(time.Now())

	jobIDs, err := s.dbrmp.addJobStatusMetrics(s.metricsBuilder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("srape: error adding job status metrics: %w", err)
	}

	err = s.dbrmp.addNumActiveRunsMetric(s.metricsBuilder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to add num active runs metric %w", err)
	}

	err = s.rmp.addMultiJobRunMetrics(jobIDs, s.metricsBuilder, now)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to add multi job run metrics: %w", err)
	}

	dbrMetrics := s.metricsBuilder.Emit(metadata.WithDatabricksInstanceName(s.dbrInstanceName))

	// spark metrics
	clusters, err := s.dbrsvc.runningClusters()
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to get running clusters: %w", err)
	}
	s.logger.Debug("found clusters", zap.Any("clusters", clusters))

	pipelines, err := s.dbrsvc.runningPipelines()
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to get pipelines: %w", err)
	}

	allSparkDbrMetrics := newSparkDbrMetrics()

	coreClusterMetrics, err := s.scmb.buildCoreMetrics(clusters, pipelines)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: error building spark metrics: %w", err)
	}
	allSparkDbrMetrics.append(coreClusterMetrics)

	execClusterMetrics, err := s.semb.buildExecutorMetrics(clusters)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to build executor metrics: %w", err)
	}
	allSparkDbrMetrics.append(execClusterMetrics)

	jobClusterMetrics, err := s.semb.buildJobMetrics(clusters)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to build job metrics: %w", err)
	}
	allSparkDbrMetrics.append(jobClusterMetrics)

	stageClusterMetrics, err := s.semb.buildStageMetrics(clusters)
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("scrape: failed to build stage metrics: %w", err)
	}
	allSparkDbrMetrics.append(stageClusterMetrics)

	out := pmetric.NewMetrics()
	dbrMetrics.ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())

	sparkMetrics := allSparkDbrMetrics.build(s.metricsBuilder, now, metadata.WithDatabricksInstanceName(s.dbrInstanceName))
	for _, metric := range sparkMetrics {
		metric.ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	return out, nil
}
