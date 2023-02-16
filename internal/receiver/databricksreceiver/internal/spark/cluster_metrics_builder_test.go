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

package spark

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
)

func TestStripSparkMetricKey(t *testing.T) {
	key := "app-20221117221047-0000.driver.BlockManager.memory.diskSpaceUsed_MB"
	appID, stripped := stripSparkMetricKey(key)
	assert.Equal(t, "app-20221117221047-0000", appID)
	assert.Equal(t, "blockmanager.memory.diskspaceused", stripped)
}

func TestClusterMetricsBuilder_GeneratedMetrics(t *testing.T) {
	mb := ClusterMetricsBuilder{
		Ssvc: NewTestSuccessSparkService(commontest.TestdataDir),
	}
	coreMetrics, err := mb.BuildCoreMetrics([]Cluster{{
		ClusterID:   "my-cluster-id",
		ClusterName: "my-cluster-name",
		State:       "my-cluster-state",
	}}, nil)
	require.NoError(t, err)

	const expectedCount = 112

	testBuilder := commontest.NewTestMetricsBuilder()
	built := coreMetrics.Build(testBuilder, pcommon.NewTimestampFromTime(time.Now()))
	pm := pmetric.NewMetrics()
	for _, metric := range built {
		metric.ResourceMetrics().MoveAndAppendTo(pm.ResourceMetrics())
	}

	assert.Equal(t, expectedCount, pm.MetricCount())
	assert.Equal(t, expectedCount, pm.DataPointCount())

	metricMap := commontest.MetricsByName(pm)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.disk_space.used", 42)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.max", 123)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.off_heap.max", 111)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.on_heap.max", 222)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.used", 333)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.off_heap.used", 444)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.on_heap.used", 555)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.remaining", 666)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.remaining.off_heap", 777)
	assertSparkDoubleGaugeEq(t, metricMap, "block_manager.memory.remaining.on_heap", 888)
	assertSparkDoubleGaugeEq(t, metricMap, "dag_scheduler.jobs.active", 999)
	assertSparkDoubleGaugeEq(t, metricMap, "dag_scheduler.jobs.all", 1111)
	assertSparkDoubleGaugeEq(t, metricMap, "dag_scheduler.stages.failed", 2222)
	assertSparkDoubleGaugeEq(t, metricMap, "dag_scheduler.stages.running", 3333)
	assertSparkDoubleGaugeEq(t, metricMap, "dag_scheduler.stages.waiting", 4444)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.direct_pool.memory", 591058)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.jvm.heap.memory", 1748700144)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.jvm.off_heap.memory", 269709952)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.major_gc.count", 5)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.major_gc.time", 748)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.mapped_pool.memory", 5555)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.minor_gc.count", 5)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.minor_gc.time", 200)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.off_heap.execution.memory", 6666)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.off_heap.storage.memory", 7777)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.off_heap.unified.memory", 8888)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.on_heap.execution.memory", 9999)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.on_heap.storage.memory", 11111)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.on_heap.unified.memory", 22222)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.jvm_rss.memory", 33333)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.jvm_v.memory", 44444)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.other_rss.memory", 55555)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.other_v.memory", 66666)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.python_rss.memory", 77777)
	assertSparkDoubleGaugeEq(t, metricMap, "executor_metrics.process_tree.python_v.memory", 88888)
	assertSparkDoubleGaugeEq(t, metricMap, "live_listener_bus.queue.appstatus.size", 99999)
	assertSparkDoubleGaugeEq(t, metricMap, "live_listener_bus.queue.executormanagement.size", 111111)
	assertSparkDoubleGaugeEq(t, metricMap, "live_listener_bus.queue.shared.size", 222222)
	assertSparkDoubleGaugeEq(t, metricMap, "live_listener_bus.queue.streams.size", 333333)
	assertSparkDoubleGaugeEq(t, metricMap, "spark_sql_operation_manager.hive_operations.count", 444444)

	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.auto_vacuum.count", 1)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.deleted_files_filtered", 2)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.filter_listing.count", 3)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.job_commit_completed", 4)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.marker_read.errors", 5)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.marker_refresh.count", 6)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.marker_refresh.errors", 7)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.markers.read", 8)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.repeated_list.count", 9)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.uncommitted_files.filtered", 10)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.untracked_files.found", 11)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.vacuum.count", 12)
	assertSparkIntSumEq(t, metricMap, "databricks.directory_commit.vacuum.errors", 13)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.checks.count", 14)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.pools_autoexpired.count", 15)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.tasks_preempted.count", 16)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.poolstarvation.time", 17)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.scheduler_overhead.time", 18)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.task_wasted.time", 19)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.active_pools", 20)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.bypass_lane_active_pools", 21)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.fast_lane_active_pools", 22)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.finished_queries_total_task.time", 23)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.lane_cleanup.marked_pools", 24)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.lane_cleanup.two_phase_pools_cleaned", 25)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.lane_cleanup.zombie_pools_cleaned", 26)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.preemption.slot_transfer_successful_preemption_iterations.count", 27)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.preemption.slot_transfer_tasks_preempted.count", 28)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.preemption.slot_transfer_wasted_task.time", 29)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.slot_reservation.gradual_decrease.count", 30)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.slot_reservation.quick_drop.count", 31)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.slot_reservation.quick_jump.count", 32)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.slot_reservation.slots_reserved", 33)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.slow_lane_active_pools", 34)
	assertSparkIntSumEq(t, metricMap, "databricks.task_scheduling_lanes.totalquerygroupsfinished", 35)
	assertSparkIntSumEq(t, metricMap, "hive_external_catalog.file_cache.hits", 36)
	assertSparkIntSumEq(t, metricMap, "hive_external_catalog.files_discovered", 37)
	assertSparkIntSumEq(t, metricMap, "hive_external_catalog.hive_client_calls", 38)
	assertSparkIntSumEq(t, metricMap, "hive_external_catalog.parallel_listing_jobs.count", 39)
	assertSparkIntSumEq(t, metricMap, "hive_external_catalog.partitions_fetched", 40)
	assertSparkIntSumEq(t, metricMap, "live_listener_bus.events_posted.count", 41)
	assertSparkIntSumEq(t, metricMap, "live_listener_bus.queue.app_status.dropped_events.count", 42)
	assertSparkIntSumEq(t, metricMap, "live_listener_bus.queue.executor_management.dropped_events.count", 43)
	assertSparkIntSumEq(t, metricMap, "live_listener_bus.queue.shared.dropped_events.count", 44)
	assertSparkIntSumEq(t, metricMap, "live_listener_bus.queue.streams.dropped_events.count", 45)

	assertSparkDoubleSumEq(t, metricMap, "jvm.cpu.time", 57690000000)
	assertSparkDoubleSumEq(t, metricMap, "timer.dag_scheduler.message_processing.time", 1.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.backend.daemon.driver.dbc_event_logging_listener.time", 1.2)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.backend.daemon.driver.data_plane_event_listener.time", 1.3)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.photon.photon_cleanup_listener.time", 1.4)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.spark.util.executor_time_logging_listener.time", 1.5)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.spark.util.usage_logging_listener.time", 1.6)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.sql.advice.advisor_listener.time", 1.7)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.sql.debugger.query_watchdog_listener.time", 1.8)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.sql.execution.ui.io_cache_listener.time", 1.9)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.databricks.sql.io.caching.repeated_reads_estimator.time", 2.0)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.spark_session.time", 2.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.execution.time", 2.2)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.execution.streaming.query_listener_bus.time", 2.3)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.execution.ui.sql_app_status_listener.time", 2.4)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.hive.thriftserver.ui.hive_thrift_server2listener.time", 2.5)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.sql.util.execution_listener_bus.time", 2.6)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.status.app_status_listener.time", 2.7)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.listener_processing.apache.spark.util.profiler_env.time", 2.8)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.queue.app_status.listener_processing.time", 2.9)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.queue.executor_management.listener_processing.time", 3.0)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.queue.shared.listener_processing.time", 3.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.live_listener_bus.queue.streams.listener_processing.time", 3.2)
}

const prefix = "databricks.spark."

func assertSparkIntGaugeEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expectedValues ...int) {
	for i, expected := range expectedValues {
		metric, ok := metricMap[prefix+metricName]
		require.True(t, ok, metricName)
		assert.EqualValues(t, expected, metric.Gauge().DataPoints().At(i).IntValue())
	}
}

func assertSparkDoubleGaugeEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expected float64) {
	metric, ok := metricMap[prefix+metricName]
	require.True(t, ok, metricName)
	assert.Equal(t, expected, metric.Gauge().DataPoints().At(0).DoubleValue())
}

func assertSparkDoubleSumEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expected float64) {
	metricName = "databricks.spark." + metricName
	metric, ok := metricMap[metricName]
	require.True(t, ok, metricName)
	assert.EqualValues(t, expected, metric.Sum().DataPoints().At(0).DoubleValue())
}

func assertSparkIntSumEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expectedValues ...int) {
	for i, expected := range expectedValues {
		metric, ok := metricMap[prefix+metricName]
		require.True(t, ok, metricName)
		assert.EqualValues(t, expected, metric.Sum().DataPoints().At(i).IntValue())
	}
}
