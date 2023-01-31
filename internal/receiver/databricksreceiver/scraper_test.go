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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/databricks"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

const testdataMetricCount = 138
const testdataDatapointCount = 172
const testdataDir = "testdata"

func TestScraper_Success(t *testing.T) {
	dbrsvc := databricks.NewTestSingleClusterService(testdataDir)
	ssvc := spark.NewTestSuccessSparkService(testdataDir)
	nopLogger := zap.NewNop()
	scrpr := scraper{
		logger:          nopLogger,
		dbrInstanceName: "my-instance",
		metricsBuilder:  commontest.NewTestMetricsBuilder(),
		rmp:             databricks.NewRunMetricsProvider(dbrsvc),
		dbrmp:           databricks.MetricsProvider{Svc: dbrsvc},
		scmb:            spark.ClusterMetricsBuilder{Ssvc: ssvc},
		semb: spark.ExtraMetricsBuilder{
			Ssvc:   ssvc,
			Logger: nopLogger,
		},
		dbrsvc: dbrsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, testdataMetricCount, metrics.MetricCount())
	assert.Equal(t, testdataDatapointCount, metrics.DataPointCount())

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		dbrResourceMetrics := metrics.ResourceMetrics().At(i)
		attrs := dbrResourceMetrics.Resource().Attributes()
		v, _ := attrs.Get("databricks.instance.name")
		assert.Equal(t, "my-instance", v.Str())
	}
}

func TestScraper_Forbidden(t *testing.T) {
	// make sure an http 403 doesn't produce an error (possible if the token owner
	// doesn't own a cluster)
	dbrsvc := databricks.NewTestSingleClusterService(testdataDir)
	ssvc := spark.NewTestForbiddenSparkService(testdataDir, "11111")
	nopLogger := zap.NewNop()
	scrpr := scraper{
		logger:          nopLogger,
		dbrInstanceName: "my-instance",
		metricsBuilder:  commontest.NewTestMetricsBuilder(),
		rmp:             databricks.NewRunMetricsProvider(dbrsvc),
		dbrmp:           databricks.MetricsProvider{Svc: dbrsvc},
		scmb:            spark.ClusterMetricsBuilder{Ssvc: ssvc},
		semb: spark.ExtraMetricsBuilder{
			Ssvc:   ssvc,
			Logger: nopLogger,
		},
		dbrsvc: dbrsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	const dbrOnlyMetricCount = 4
	assert.Equal(t, dbrOnlyMetricCount, metrics.MetricCount())
}

func TestScraper_MultiCluster_Forbidden(t *testing.T) {
	dbrsvc := databricks.NewTestMultiClusterService(testdataDir)
	ssvc := spark.NewTestForbiddenSparkService(testdataDir, "22222")
	nopLogger := zap.NewNop()
	scrpr := scraper{
		logger:          nopLogger,
		dbrInstanceName: "my-instance",
		metricsBuilder:  commontest.NewTestMetricsBuilder(),
		rmp:             databricks.NewRunMetricsProvider(dbrsvc),
		dbrmp:           databricks.MetricsProvider{Svc: dbrsvc},
		scmb:            spark.ClusterMetricsBuilder{Ssvc: ssvc},
		semb: spark.ExtraMetricsBuilder{
			Ssvc:   ssvc,
			Logger: nopLogger,
		},
		dbrsvc: dbrsvc,
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, testdataMetricCount, metrics.MetricCount())
	assert.Equal(t, testdataDatapointCount, metrics.DataPointCount())
}
