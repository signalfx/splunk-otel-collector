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

package dbrspark

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/databricks"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func TestSparkMetricsBuilder_Executors(t *testing.T) {
	semb := newTestExtraMetricsBuilder()
	execMetrics, err := semb.BuildExecutorMetrics([]spark.Cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	builder := commontest.NewTestMetricsBuilder()
	built := execMetrics.Build(builder, pcommon.NewTimestampFromTime(time.Now()))
	pm := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(pm.ResourceMetrics())
	}

	metricMap := databricks.MetricsByName(pm)
	assertSparkIntGaugeEq(t, metricMap, "executor.memory_used", 709517, 636121)
	assertSparkIntGaugeEq(t, metricMap, "executor.disk_used", 1, 2)
	assertSparkIntSumEq(t, metricMap, "executor.total_input_bytes", 3, 4)
	assertSparkIntSumEq(t, metricMap, "executor.total_shuffle_read", 5, 6)
	assertSparkIntSumEq(t, metricMap, "executor.total_shuffle_write", 7, 8)
	assertSparkIntGaugeEq(t, metricMap, "executor.max_memory", 4544318668, 4773956812)
}

func TestSparkMetricsBuilder_Jobs(t *testing.T) {
	semb := newTestExtraMetricsBuilder()
	jobMetrics, err := semb.BuildJobMetrics([]spark.Cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	builder := commontest.NewTestMetricsBuilder()
	built := jobMetrics.Build(builder, pcommon.NewTimestampFromTime(time.Now()))
	pm := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(pm.ResourceMetrics())
	}

	metricMap := databricks.MetricsByName(pm)
	assertSparkIntGaugeEq(t, metricMap, "job.num_tasks", 9, 8)
	assertSparkIntGaugeEq(t, metricMap, "job.num_active_tasks", 10, 11)
	assertSparkIntGaugeEq(t, metricMap, "job.num_completed_tasks", 12, 13)
	assertSparkIntGaugeEq(t, metricMap, "job.num_skipped_tasks", 14, 15)
	assertSparkIntGaugeEq(t, metricMap, "job.num_failed_tasks", 16, 17)
	assertSparkIntGaugeEq(t, metricMap, "job.num_active_stages", 18, 19)
	assertSparkIntGaugeEq(t, metricMap, "job.num_completed_stages", 20, 21)
	assertSparkIntGaugeEq(t, metricMap, "job.num_skipped_stages", 22, 23)
	assertSparkIntGaugeEq(t, metricMap, "job.num_failed_stages", 24, 25)
}

func TestSparkMetricsBuilder_Stages(t *testing.T) {
	semb := newTestExtraMetricsBuilder()
	stageMetrics, err := semb.BuildStageMetrics([]spark.Cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	builder := commontest.NewTestMetricsBuilder()
	built := stageMetrics.Build(builder, pcommon.NewTimestampFromTime(time.Now()))
	pm := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(pm.ResourceMetrics())
	}
	metricMap := databricks.MetricsByName(pm)
	assertSparkIntGaugeEq(t, metricMap, "stage.executor_run_time", 1, 2)
	assertSparkIntGaugeEq(t, metricMap, "stage.input_bytes", 3, 4)
	assertSparkIntGaugeEq(t, metricMap, "stage.input_records", 5, 6)
	assertSparkIntGaugeEq(t, metricMap, "stage.output_bytes", 7, 8)
	assertSparkIntGaugeEq(t, metricMap, "stage.output_records", 9, 10)
	assertSparkIntGaugeEq(t, metricMap, "stage.memory_bytes_spilled", 11, 12)
	assertSparkIntGaugeEq(t, metricMap, "stage.disk_bytes_spilled", 13, 14)
}

func newTestExtraMetricsBuilder() ExtraMetricsBuilder {
	return ExtraMetricsBuilder{
		Ssvc:   NewTestSuccessSparkService(testdataDir),
		Logger: zap.NewNop(),
	}
}
