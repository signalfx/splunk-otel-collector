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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

func NewResourceMetrics() *ResourceMetrics {
	return &ResourceMetrics{
		m: map[resource][]metricBuilder{},
	}
}

type ResourceMetrics struct {
	m map[resource][]metricBuilder
}

type metricBuilder interface {
	build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp)
}

type resource struct {
	cluster Cluster
	appID   string
}

func (m *ResourceMetrics) Append(other *ResourceMetrics) {
	for k, v := range other.m {
		if _, found := m.m[k]; found {
			m.m[k] = append(m.m[k], v...)
			continue
		}
		m.m[k] = v
	}
}

func (m *ResourceMetrics) addGauge(cluster Cluster, appID string, gauge Gauge, mb sparkMetricBase) {
	rs := resource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &gaugeInfo{
		sparkMetricBase: mb,
		gauge:           gauge,
	})
}

func (m *ResourceMetrics) addCounter(cluster Cluster, appID string, counter Counter, mb sparkMetricBase) {
	rs := resource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &counterInfo{
		sparkMetricBase: mb,
		counter:         counter,
	})
}

func (m *ResourceMetrics) addTimer(cluster Cluster, appID string, timer Timer, mb sparkMetricBase) {
	rs := resource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &timerInfo{
		sparkMetricBase: mb,
		timer:           timer,
	})
}

func (m *ResourceMetrics) addHisto(cluster Cluster, appID string, histo Histogram, mb sparkMetricBase) {
	rs := resource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &histoInfo{
		sparkMetricBase: mb,
		histo:           histo,
	})
}

func (m *ResourceMetrics) addExecInfo(clstr Cluster, appID string, info ExecutorInfo) {
	resrc := resource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &execInfo{
		execInfo: info,
	})
}

func (m *ResourceMetrics) addJobInfos(clstr Cluster, appID string, info JobInfo) {
	resrc := resource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &jobInfo{
		jobInfo: info,
	})
}

func (m *ResourceMetrics) addStageInfo(clstr Cluster, appID string, info StageInfo) {
	resrc := resource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &stageInfo{
		stageInfo: info,
	})
}

func (m *ResourceMetrics) Build(mb *metadata.MetricsBuilder, rb *metadata.ResourceBuilder, now pcommon.Timestamp, instanceName string) []pmetric.Metrics {
	var out []pmetric.Metrics
	for rs, metricInfos := range m.m {
		for _, mi := range metricInfos {
			mi.build(mb, rs, now)
		}
		rb.SetDatabricksInstanceName(instanceName)
		rb.SetSparkClusterID(rs.cluster.ClusterID)
		rb.SetSparkClusterName(rs.cluster.ClusterName)
		out = append(out, mb.Emit(metadata.WithResource(rb.Emit())))
	}
	return out
}

func buildGauge(
	builder *metadata.MetricsBuilder,
	partialMetricName string,
	now pcommon.Timestamp,
	gauge Gauge,
	rs resource,
	pipelineID string,
	pipelineName string,
) {
	clusterID := rs.cluster.ClusterID
	appID := rs.appID
	switch partialMetricName {
	case "blockmanager.memory.diskspaceused":
		builder.RecordDatabricksSparkBlockManagerMemoryDiskSpaceUsedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxmem":
		builder.RecordDatabricksSparkBlockManagerMemoryMaxDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxoffheapmem":
		builder.RecordDatabricksSparkBlockManagerMemoryOffHeapMaxDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxonheapmem":
		builder.RecordDatabricksSparkBlockManagerMemoryOnHeapMaxDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.memused":
		builder.RecordDatabricksSparkBlockManagerMemoryUsedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.offheapmemused":
		builder.RecordDatabricksSparkBlockManagerMemoryOffHeapUsedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.onheapmemused":
		builder.RecordDatabricksSparkBlockManagerMemoryOnHeapUsedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingmem":
		builder.RecordDatabricksSparkBlockManagerMemoryRemainingDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingoffheapmem":
		builder.RecordDatabricksSparkBlockManagerMemoryRemainingOffHeapDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingonheapmem":
		builder.RecordDatabricksSparkBlockManagerMemoryRemainingOnHeapDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.job.activejobs":
		builder.RecordDatabricksSparkDagSchedulerJobsActiveDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.job.alljobs":
		builder.RecordDatabricksSparkDagSchedulerJobsAllDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.failedstages":
		builder.RecordDatabricksSparkDagSchedulerStagesFailedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.runningstages":
		builder.RecordDatabricksSparkDagSchedulerStagesRunningDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.waitingstages":
		builder.RecordDatabricksSparkDagSchedulerStagesWaitingDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.directpoolmemory":
		builder.RecordDatabricksSparkExecutorMetricsDirectPoolMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.jvmheapmemory":
		builder.RecordDatabricksSparkExecutorMetricsJvmHeapMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.jvmoffheapmemory":
		builder.RecordDatabricksSparkExecutorMetricsJvmOffHeapMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.majorgccount":
		builder.RecordDatabricksSparkExecutorMetricsMajorGcCountDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.majorgctime":
		builder.RecordDatabricksSparkExecutorMetricsMajorGcTimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.mappedpoolmemory":
		builder.RecordDatabricksSparkExecutorMetricsMappedPoolMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.minorgccount":
		builder.RecordDatabricksSparkExecutorMetricsMinorGcCountDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.minorgctime":
		builder.RecordDatabricksSparkExecutorMetricsMinorGcTimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapexecutionmemory":
		builder.RecordDatabricksSparkExecutorMetricsOffHeapExecutionMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapstoragememory":
		builder.RecordDatabricksSparkExecutorMetricsOffHeapStorageMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapunifiedmemory":
		builder.RecordDatabricksSparkExecutorMetricsOffHeapUnifiedMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapexecutionmemory":
		builder.RecordDatabricksSparkExecutorMetricsOnHeapExecutionMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapstoragememory":
		builder.RecordDatabricksSparkExecutorMetricsOnHeapStorageMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapunifiedmemory":
		builder.RecordDatabricksSparkExecutorMetricsOnHeapUnifiedMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreejvmrssmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreeJvmRssMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreejvmvmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreeJvmVMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreeotherrssmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreeOtherRssMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreeothervmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreeOtherVMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreepythonrssmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreePythonRssMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreepythonvmemory":
		builder.RecordDatabricksSparkExecutorMetricsProcessTreePythonVMemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "jvmcpu.jvmcputime":
		builder.RecordDatabricksSparkJvmCPUTimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.appstatus.size":
		builder.RecordDatabricksSparkLiveListenerBusQueueAppstatusSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.executormanagement.size":
		builder.RecordDatabricksSparkLiveListenerBusQueueExecutormanagementSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.shared.size":
		builder.RecordDatabricksSparkLiveListenerBusQueueSharedSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.streams.size":
		builder.RecordDatabricksSparkLiveListenerBusQueueStreamsSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "sparksqloperationmanager.numhiveoperations":
		builder.RecordDatabricksSparkSparkSQLOperationManagerHiveOperationsCountDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	}
}

func buildCounter(
	builder *metadata.MetricsBuilder,
	partialMetricName string,
	now pcommon.Timestamp,
	counter Counter,
	rs resource,
	pipelineID string,
	pipelineName string,
) {
	clusterID := rs.cluster.ClusterID
	appID := rs.appID
	switch partialMetricName {
	case "databricks.directorycommit.autovacuumcount":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitAutoVacuumCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.deletedfilesfiltered":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitDeletedFilesFilteredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.filterlistingcount":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitFilterListingCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.jobcommitcompleted":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitJobCommitCompletedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerreaderrors":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitMarkerReadErrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerrefreshcount":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitMarkerRefreshCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerrefresherrors":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitMarkerRefreshErrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markersread":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitMarkersReadDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.repeatedlistcount":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitRepeatedListCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.uncommittedfilesfiltered":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitUncommittedFilesFilteredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.untrackedfilesfound":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitUntrackedFilesFoundDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.vacuumcount":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitVacuumCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.vacuumerrors":
		builder.RecordDatabricksSparkDatabricksDirectoryCommitVacuumErrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numchecks":
		builder.RecordDatabricksSparkDatabricksPreemptionChecksCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numpoolsautoexpired":
		builder.RecordDatabricksSparkDatabricksPreemptionPoolsAutoexpiredCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numtaskspreempted":
		builder.RecordDatabricksSparkDatabricksPreemptionTasksPreemptedCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.poolstarvationmillis":
		builder.RecordDatabricksSparkDatabricksPreemptionPoolstarvationTimeDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.scheduleroverheadnanos":
		builder.RecordDatabricksSparkDatabricksPreemptionSchedulerOverheadTimeDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.tasktimewastedmillis":
		builder.RecordDatabricksSparkDatabricksPreemptionTaskWastedTimeDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.activepools":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesActivePoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.bypasslaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesBypassLaneActivePoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.fastlaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesFastLaneActivePoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.finishedqueriestotaltasktimens":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesFinishedQueriesTotalTaskTimeDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.markedpools":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesLaneCleanupMarkedPoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.twophasepoolscleaned":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesLaneCleanupTwoPhasePoolsCleanedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.zombiepoolscleaned":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesLaneCleanupZombiePoolsCleanedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransfernumsuccessfulpreemptioniterations":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesPreemptionSlotTransferSuccessfulPreemptionIterationsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransfernumtaskspreempted":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesPreemptionSlotTransferTasksPreemptedCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransferwastedtasktimens":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesPreemptionSlotTransferWastedTaskTimeDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numgradualdecrease":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesSlotReservationGradualDecreaseCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numquickdrop":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesSlotReservationQuickDropCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numquickjump":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesSlotReservationQuickJumpCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.slotsreserved":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesSlotReservationSlotsReservedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slowlaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesSlowLaneActivePoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.totalquerygroupsfinished":
		builder.RecordDatabricksSparkDatabricksTaskSchedulingLanesTotalquerygroupsfinishedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.filecachehits":
		builder.RecordDatabricksSparkHiveExternalCatalogFileCacheHitsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.filesdiscovered":
		builder.RecordDatabricksSparkHiveExternalCatalogFilesDiscoveredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.hiveclientcalls":
		builder.RecordDatabricksSparkHiveExternalCatalogHiveClientCallsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.parallellistingjobcount":
		builder.RecordDatabricksSparkHiveExternalCatalogParallelListingJobsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.partitionsfetched":
		builder.RecordDatabricksSparkHiveExternalCatalogPartitionsFetchedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.numeventsposted":
		builder.RecordDatabricksSparkLiveListenerBusEventsPostedCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.appstatus.numdroppedevents":
		builder.RecordDatabricksSparkLiveListenerBusQueueAppStatusDroppedEventsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.executormanagement.numdroppedevents":
		builder.RecordDatabricksSparkLiveListenerBusQueueExecutorManagementDroppedEventsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.shared.numdroppedevents":
		builder.RecordDatabricksSparkLiveListenerBusQueueSharedDroppedEventsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.streams.numdroppedevents":
		builder.RecordDatabricksSparkLiveListenerBusQueueStreamsDroppedEventsCountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	}
}

func buildTimers(
	builder *metadata.MetricsBuilder,
	partialMetricName string,
	now pcommon.Timestamp,
	timer Timer,
	rs resource,
) {
	appID := rs.appID
	clusterID := rs.cluster.ClusterID
	switch partialMetricName {
	case "dagscheduler.messageprocessingtime":
		builder.RecordDatabricksSparkTimerDagSchedulerMessageProcessingTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.backend.daemon.driver.dbceventlogginglistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksBackendDaemonDriverDbcEventLoggingListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.backend.daemon.driver.dataplaneeventlistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksBackendDaemonDriverDataPlaneEventListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.photon.photoncleanuplistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksPhotonPhotonCleanupListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.spark.util.executortimelogginglistener$":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSparkUtilExecutorTimeLoggingListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.spark.util.usagelogginglistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSparkUtilUsageLoggingListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.advice.advisorlistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSQLAdviceAdvisorListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.debugger.querywatchdoglistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSQLDebuggerQueryWatchdogListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.execution.ui.iocachelistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSQLExecutionUIIoCacheListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.io.caching.repeatedreadsestimator$":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingDatabricksSQLIoCachingRepeatedReadsEstimatorTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.sparksession$$anon$1":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLSparkSessionTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.sqlexecution$":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLExecutionTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.streaming.streamingquerylistenerbus":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLExecutionStreamingQueryListenerBusTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.ui.sqlappstatuslistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLExecutionUISQLAppStatusListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.hive.thriftserver.ui.hivethriftserver2listener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLHiveThriftserverUIHiveThriftServer2listenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.util.executionlistenerbus":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkSQLUtilExecutionListenerBusTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.status.appstatuslistener":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkStatusAppStatusListenerTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.util.profilerenv$$anon$1":
		builder.RecordDatabricksSparkTimerLiveListenerBusListenerProcessingApacheSparkUtilProfilerEnvTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.appstatus.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLiveListenerBusQueueAppStatusListenerProcessingTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.executormanagement.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLiveListenerBusQueueExecutorManagementListenerProcessingTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.shared.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLiveListenerBusQueueSharedListenerProcessingTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.streams.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLiveListenerBusQueueStreamsListenerProcessingTimeDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	}
}

func buildHistos(
	builder *metadata.MetricsBuilder,
	partialMetricName string,
	now pcommon.Timestamp,
	histo Histogram,
	rs resource,
	pipelineID string,
	pipelineName string,
) {
	appID := rs.appID
	clusterID := rs.cluster.ClusterID
	switch partialMetricName {
	case "codegenerator.compilationtime":
		builder.RecordDatabricksSparkCodeGeneratorCompilationTimeDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.generatedclasssize":
		builder.RecordDatabricksSparkCodeGeneratorGeneratedClassSizeDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.generatedmethodsize":
		builder.RecordDatabricksSparkCodeGeneratorGeneratedMethodSizeDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.sourcecodesize":
		builder.RecordDatabricksSparkCodeGeneratorSourcecodeSizeDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	}
}

func buildExecMetrics(builder *metadata.MetricsBuilder, execInfo ExecutorInfo, now pcommon.Timestamp, rs resource) {
	builder.RecordDatabricksSparkExecutorMemoryUsedDataPoint(now, int64(execInfo.MemoryUsed), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorDiskUsedDataPoint(now, int64(execInfo.DiskUsed), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalInputBytesDataPoint(now, execInfo.TotalInputBytes, rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalShuffleReadDataPoint(now, int64(execInfo.TotalShuffleRead), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalShuffleWriteDataPoint(now, int64(execInfo.TotalShuffleWrite), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorMaxMemoryDataPoint(now, execInfo.MaxMemory, rs.cluster.ClusterID, rs.appID, execInfo.ID)
}

func buildJobMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, jobInfo JobInfo, rs resource) {
	builder.RecordDatabricksSparkJobNumTasksDataPoint(now, int64(jobInfo.NumTasks), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumActiveTasksDataPoint(now, int64(jobInfo.NumActiveTasks), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumCompletedTasksDataPoint(now, int64(jobInfo.NumCompletedTasks), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumSkippedTasksDataPoint(now, int64(jobInfo.NumSkippedTasks), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumFailedTasksDataPoint(now, int64(jobInfo.NumFailedTasks), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumActiveStagesDataPoint(now, int64(jobInfo.NumActiveStages), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumCompletedStagesDataPoint(now, int64(jobInfo.NumCompletedStages), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumSkippedStagesDataPoint(now, int64(jobInfo.NumSkippedStages), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
	builder.RecordDatabricksSparkJobNumFailedStagesDataPoint(now, int64(jobInfo.NumFailedStages), rs.cluster.ClusterID, rs.appID, int64(jobInfo.JobID))
}

func buildStageMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, stageInfo StageInfo, rs resource) {
	builder.RecordDatabricksSparkStageExecutorRunTimeDataPoint(now, int64(stageInfo.ExecutorRunTime), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageInputBytesDataPoint(now, int64(stageInfo.InputBytes), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageInputRecordsDataPoint(now, int64(stageInfo.InputRecords), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageOutputBytesDataPoint(now, int64(stageInfo.OutputBytes), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageOutputRecordsDataPoint(now, int64(stageInfo.OutputRecords), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageMemoryBytesSpilledDataPoint(now, int64(stageInfo.MemoryBytesSpilled), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageDiskBytesSpilledDataPoint(now, int64(stageInfo.DiskBytesSpilled), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
}

func newSparkMetricBase(partialMetricName string, pipeline *PipelineSummary) sparkMetricBase {
	pipelineID := ""
	pipelineName := ""
	if pipeline != nil {
		pipelineID = pipeline.ID
		pipelineName = pipeline.Name
	}
	return sparkMetricBase{
		partialMetricName: partialMetricName,
		pipelineID:        pipelineID,
		pipelineName:      pipelineName,
	}
}

type sparkMetricBase struct {
	partialMetricName string
	pipelineID        string
	pipelineName      string
}

type gaugeInfo struct {
	sparkMetricBase
	gauge Gauge
}

func (i gaugeInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildGauge(builder, i.partialMetricName, now, i.gauge, rs, i.pipelineID, i.pipelineName)
}

type counterInfo struct {
	sparkMetricBase
	counter Counter
}

func (i counterInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildCounter(builder, i.partialMetricName, now, i.counter, rs, i.pipelineID, i.pipelineName)
}

type timerInfo struct {
	sparkMetricBase
	timer Timer
}

func (i timerInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildTimers(builder, i.partialMetricName, now, i.timer, rs)
}

type histoInfo struct {
	sparkMetricBase
	histo Histogram
}

func (i histoInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildHistos(builder, i.partialMetricName, now, i.histo, rs, i.pipelineID, i.pipelineName)
}

type execInfo struct {
	execInfo ExecutorInfo
}

func (i execInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildExecMetrics(builder, i.execInfo, now, rs)
}

type jobInfo struct {
	jobInfo JobInfo
}

func (i jobInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildJobMetrics(builder, now, i.jobInfo, rs)
}

type stageInfo struct {
	stageInfo StageInfo
}

func (i stageInfo) build(builder *metadata.MetricsBuilder, rs resource, now pcommon.Timestamp) {
	buildStageMetrics(builder, now, i.stageInfo, rs)
}
