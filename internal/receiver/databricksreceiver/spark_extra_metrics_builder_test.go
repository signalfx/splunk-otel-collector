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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestSparkMetricsBuilder_Executors(t *testing.T) {
	semb := sparkExtraMetricsBuilder{
		ssvc:   newTestSuccessSparkService(newTestDatabricksSingleClusterService()),
		logger: zap.NewNop(),
	}
	execMetrics, err := semb.buildExecutorMetrics([]cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	testBuilder := newTestMetricsBuilder()
	built := execMetrics.build(testBuilder, pcommon.NewTimestampFromTime(time.Now()))
	emitted := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(emitted.ResourceMetrics())
	}

	metricMap := metricsByName(emitted)
	assertIntGaugeEq(t, metricMap, "executor.memory_used", 709517, 636121)
	assertIntGaugeEq(t, metricMap, "executor.disk_used", 1, 2)
	assertIntSumEq(t, metricMap, "executor.total_input_bytes", 3, 4)
	assertIntSumEq(t, metricMap, "executor.total_shuffle_read", 5, 6)
	assertIntSumEq(t, metricMap, "executor.total_shuffle_write", 7, 8)
	assertIntGaugeEq(t, metricMap, "executor.max_memory", 4544318668, 4773956812)
}

func TestSparkMetricsBuilder_Jobs(t *testing.T) {
	semb := sparkExtraMetricsBuilder{
		ssvc:   newTestSuccessSparkService(newTestDatabricksSingleClusterService()),
		logger: zap.NewNop(),
	}
	jobMetrics, err := semb.buildJobMetrics([]cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	builder := newTestMetricsBuilder()
	built := jobMetrics.build(builder, pcommon.NewTimestampFromTime(time.Now()))
	emitted := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(emitted.ResourceMetrics())
	}

	metricMap := metricsByName(emitted)
	assertIntGaugeEq(t, metricMap, "job.num_tasks", 9, 8)
	assertIntGaugeEq(t, metricMap, "job.num_active_tasks", 10, 11)
	assertIntGaugeEq(t, metricMap, "job.num_completed_tasks", 12, 13)
	assertIntGaugeEq(t, metricMap, "job.num_skipped_tasks", 14, 15)
	assertIntGaugeEq(t, metricMap, "job.num_failed_tasks", 16, 17)
	assertIntGaugeEq(t, metricMap, "job.num_active_stages", 18, 19)
	assertIntGaugeEq(t, metricMap, "job.num_completed_stages", 20, 21)
	assertIntGaugeEq(t, metricMap, "job.num_skipped_stages", 22, 23)
	assertIntGaugeEq(t, metricMap, "job.num_failed_stages", 24, 25)
}

func TestSparkMetricsBuilder_Stages(t *testing.T) {
	semb := sparkExtraMetricsBuilder{
		ssvc:   newTestSuccessSparkService(newTestDatabricksSingleClusterService()),
		logger: zap.NewNop(),
	}
	stageMetrics, err := semb.buildStageMetrics([]cluster{{ClusterID: "foo"}})
	require.NoError(t, err)

	testBuilder := newTestMetricsBuilder()
	built := stageMetrics.build(testBuilder, pcommon.NewTimestampFromTime(time.Now()))
	emitted := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().MoveAndAppendTo(emitted.ResourceMetrics())
	}
	metricMap := metricsByName(emitted)
	assertIntGaugeEq(t, metricMap, "stage.executor_run_time", 1, 2)
	assertIntGaugeEq(t, metricMap, "stage.input_bytes", 3, 4)
	assertIntGaugeEq(t, metricMap, "stage.input_records", 5, 6)
	assertIntGaugeEq(t, metricMap, "stage.output_bytes", 7, 8)
	assertIntGaugeEq(t, metricMap, "stage.output_records", 9, 10)
	assertIntGaugeEq(t, metricMap, "stage.memory_bytes_spilled", 11, 12)
	assertIntGaugeEq(t, metricMap, "stage.disk_bytes_spilled", 13, 14)
}
