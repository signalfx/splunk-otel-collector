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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

func TestScraper_Success(t *testing.T) {
	dbsvc := newTestDatabricksSingleClusterService()
	ssvc := newTestSuccessSparkService(dbsvc)
	nopLogger := zap.NewNop()
	scrpr := scraper{
		logger:      nopLogger,
		resourceOpt: metadata.WithDatabricksInstanceName("my-instance"),
		builder:     newTestMetricsBuilder(),
		rmp:         newRunMetricsProvider(dbsvc),
		dbmp:        dbMetricsProvider{dbsvc},
		scmb:        sparkClusterMetricsBuilder{ssvc},
		semb: sparkExtraMetricsBuilder{
			ssvc:   ssvc,
			logger: nopLogger,
		},
		dbsvc: dbsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 138, metrics.MetricCount())
	rms := metrics.ResourceMetrics().At(0)
	attrs := rms.Resource().Attributes()
	v, _ := attrs.Get("databricks.instance.name")
	assert.Equal(t, "my-instance", v.Str())
	metricMap := metricsByName(rms.ScopeMetrics().At(0).Metrics())
	histoMetricName := "databricks.spark.codegenerator.compilationtime"
	metric := metricMap[histoMetricName]
	// spot check that a histogram made it through
	assert.Equal(t, histoMetricName, metric.Name())
}

func TestScraper_Forbidden(t *testing.T) {
	// make sure an http 403 doesn't produce an error (possible if the token owner
	// doesn't own a cluster)
	dbsvc := newTestDatabricksSingleClusterService()
	ssvc := newTestForbiddenSparkService(dbsvc)
	nopLogger := zap.NewNop()
	scrpr := scraper{
		logger:      nopLogger,
		resourceOpt: metadata.WithDatabricksInstanceName("my-instance"),
		builder:     newTestMetricsBuilder(),
		rmp:         newRunMetricsProvider(dbsvc),
		dbmp:        dbMetricsProvider{dbsvc},
		scmb:        sparkClusterMetricsBuilder{ssvc},
		semb: sparkExtraMetricsBuilder{
			ssvc:   ssvc,
			logger: nopLogger,
		},
		dbsvc: dbsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 26, metrics.MetricCount())
}

func TestScraper_MutliCluster_Forbidden(t *testing.T) {
	dbsvc := newTestDatabricksMultiClusterService()
	ssvc := newTestForbiddenSparkService(dbsvc)
	nopLogger := zap.NewNop()
	builder := newTestMetricsBuilder()
	scrpr := scraper{
		logger:      nopLogger,
		resourceOpt: metadata.WithDatabricksInstanceName("my-instance"),
		builder:     builder,
		rmp:         newRunMetricsProvider(dbsvc),
		dbmp:        dbMetricsProvider{dbsvc},
		scmb:        sparkClusterMetricsBuilder{ssvc},
		semb: sparkExtraMetricsBuilder{
			ssvc:   ssvc,
			logger: nopLogger,
		},
		dbsvc: dbsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 138, metrics.MetricCount())
}
