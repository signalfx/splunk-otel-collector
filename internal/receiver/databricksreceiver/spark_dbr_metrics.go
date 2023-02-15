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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func newSparkDbrMetrics() *sparkDbrMetrics {
	return &sparkDbrMetrics{
		m: map[sparkResource][]sparkMetricBuilder{},
	}
}

type sparkDbrMetrics struct {
	m map[sparkResource][]sparkMetricBuilder
}

type sparkMetricBuilder interface {
	build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp)
}

type sparkResource struct {
	cluster cluster
	appID   string
}

func (m *sparkDbrMetrics) append(other *sparkDbrMetrics) {
	for k, v := range other.m {
		if _, found := m.m[k]; found {
			m.m[k] = append(m.m[k], v...)
			continue
		}
		m.m[k] = v
	}
}

func (m *sparkDbrMetrics) addGauge(cluster cluster, appID string, gauge spark.Gauge, mb sparkMetricBase) {
	rs := sparkResource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &gaugeInfo{
		sparkMetricBase: mb,
		gauge:           gauge,
	})
}

func (m *sparkDbrMetrics) addCounter(cluster cluster, appID string, counter spark.Counter, mb sparkMetricBase) {
	rs := sparkResource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &counterInfo{
		sparkMetricBase: mb,
		counter:         counter,
	})
}

func (m *sparkDbrMetrics) addTimer(cluster cluster, appID string, timer spark.Timer, mb sparkMetricBase) {
	rs := sparkResource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &timerInfo{
		sparkMetricBase: mb,
		timer:           timer,
	})
}

func (m *sparkDbrMetrics) addHisto(cluster cluster, appID string, histo spark.Histogram, mb sparkMetricBase) {
	rs := sparkResource{
		cluster: cluster,
		appID:   appID,
	}
	m.m[rs] = append(m.m[rs], &histoInfo{
		sparkMetricBase: mb,
		histo:           histo,
	})
}

func (m *sparkDbrMetrics) addExecInfo(clstr cluster, appID string, info spark.ExecutorInfo) {
	resrc := sparkResource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &execInfo{
		execInfo: info,
	})
}

func (m *sparkDbrMetrics) addJobInfos(clstr cluster, appID string, info spark.JobInfo) {
	resrc := sparkResource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &jobInfo{
		jobInfo: info,
	})
}

func (m *sparkDbrMetrics) addStageInfo(clstr cluster, appID string, info spark.StageInfo) {
	resrc := sparkResource{
		cluster: clstr,
		appID:   appID,
	}
	m.m[resrc] = append(m.m[resrc], &stageInfo{
		stageInfo: info,
	})
}

func (m *sparkDbrMetrics) build(builder *metadata.MetricsBuilder, now pcommon.Timestamp, rmo ...metadata.ResourceMetricsOption) []pmetric.Metrics {
	var out []pmetric.Metrics
	for rs, metricInfos := range m.m {
		for _, mi := range metricInfos {
			mi.build(builder, rs, now)
		}
		rmo = append(rmo, metadata.WithSparkClusterID(rs.cluster.ClusterID))
		rmo = append(rmo, metadata.WithSparkClusterName(rs.cluster.ClusterName))
		out = append(out, builder.Emit(rmo...))
	}
	return out
}

func buildGauge(
	builder *metadata.MetricsBuilder,
	partialMetricName string,
	now pcommon.Timestamp,
	gauge spark.Gauge,
	rs sparkResource,
	pipelineID string,
	pipelineName string,
) {
	clusterID := rs.cluster.ClusterID
	appID := rs.appID
	switch partialMetricName {
	case "blockmanager.memory.diskspaceused":
		builder.RecordDatabricksSparkBlockmanagerMemoryDiskspaceusedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryMaxmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxoffheapmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryMaxoffheapmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.maxonheapmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryMaxonheapmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.memused":
		builder.RecordDatabricksSparkBlockmanagerMemoryMemusedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.offheapmemused":
		builder.RecordDatabricksSparkBlockmanagerMemoryOffheapmemusedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.onheapmemused":
		builder.RecordDatabricksSparkBlockmanagerMemoryOnheapmemusedDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryRemainingmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingoffheapmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryRemainingoffheapmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "blockmanager.memory.remainingonheapmem":
		builder.RecordDatabricksSparkBlockmanagerMemoryRemainingonheapmemDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.job.activejobs":
		builder.RecordDatabricksSparkDagschedulerJobActivejobsDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.job.alljobs":
		builder.RecordDatabricksSparkDagschedulerJobAlljobsDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.failedstages":
		builder.RecordDatabricksSparkDagschedulerStageFailedstagesDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.runningstages":
		builder.RecordDatabricksSparkDagschedulerStageRunningstagesDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "dagscheduler.stage.waitingstages":
		builder.RecordDatabricksSparkDagschedulerStageWaitingstagesDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.directpoolmemory":
		builder.RecordDatabricksSparkExecutormetricsDirectpoolmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.jvmheapmemory":
		builder.RecordDatabricksSparkExecutormetricsJvmheapmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.jvmoffheapmemory":
		builder.RecordDatabricksSparkExecutormetricsJvmoffheapmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.majorgccount":
		builder.RecordDatabricksSparkExecutormetricsMajorgccountDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.majorgctime":
		builder.RecordDatabricksSparkExecutormetricsMajorgctimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.mappedpoolmemory":
		builder.RecordDatabricksSparkExecutormetricsMappedpoolmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.minorgccount":
		builder.RecordDatabricksSparkExecutormetricsMinorgccountDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.minorgctime":
		builder.RecordDatabricksSparkExecutormetricsMinorgctimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapexecutionmemory":
		builder.RecordDatabricksSparkExecutormetricsOffheapexecutionmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapstoragememory":
		builder.RecordDatabricksSparkExecutormetricsOffheapstoragememoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.offheapunifiedmemory":
		builder.RecordDatabricksSparkExecutormetricsOffheapunifiedmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapexecutionmemory":
		builder.RecordDatabricksSparkExecutormetricsOnheapexecutionmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapstoragememory":
		builder.RecordDatabricksSparkExecutormetricsOnheapstoragememoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.onheapunifiedmemory":
		builder.RecordDatabricksSparkExecutormetricsOnheapunifiedmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreejvmrssmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreejvmrssmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreejvmvmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreejvmvmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreeotherrssmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreeotherrssmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreeothervmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreeothervmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreepythonrssmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreepythonrssmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "executormetrics.processtreepythonvmemory":
		builder.RecordDatabricksSparkExecutormetricsProcesstreepythonvmemoryDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "jvmcpu.jvmcputime":
		builder.RecordDatabricksSparkJvmcpuJvmcputimeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.appstatus.size":
		builder.RecordDatabricksSparkLivelistenerbusQueueAppstatusSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.executormanagement.size":
		builder.RecordDatabricksSparkLivelistenerbusQueueExecutormanagementSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.shared.size":
		builder.RecordDatabricksSparkLivelistenerbusQueueSharedSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.streams.size":
		builder.RecordDatabricksSparkLivelistenerbusQueueStreamsSizeDataPoint(
			now,
			gauge.Value,
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "sparksqloperationmanager.numhiveoperations":
		builder.RecordDatabricksSparkSparksqloperationmanagerNumhiveoperationsDataPoint(
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
	counter spark.Counter,
	rs sparkResource,
	pipelineID string,
	pipelineName string,
) {
	clusterID := rs.cluster.ClusterID
	appID := rs.appID
	switch partialMetricName {
	case "databricks.directorycommit.autovacuumcount":
		builder.RecordDatabricksSparkDatabricksDirectorycommitAutovacuumcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.deletedfilesfiltered":
		builder.RecordDatabricksSparkDatabricksDirectorycommitDeletedfilesfilteredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.filterlistingcount":
		builder.RecordDatabricksSparkDatabricksDirectorycommitFilterlistingcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.jobcommitcompleted":
		builder.RecordDatabricksSparkDatabricksDirectorycommitJobcommitcompletedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerreaderrors":
		builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerreaderrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerrefreshcount":
		builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerrefreshcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markerrefresherrors":
		builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerrefresherrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.markersread":
		builder.RecordDatabricksSparkDatabricksDirectorycommitMarkersreadDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.repeatedlistcount":
		builder.RecordDatabricksSparkDatabricksDirectorycommitRepeatedlistcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.uncommittedfilesfiltered":
		builder.RecordDatabricksSparkDatabricksDirectorycommitUncommittedfilesfilteredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.untrackedfilesfound":
		builder.RecordDatabricksSparkDatabricksDirectorycommitUntrackedfilesfoundDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.vacuumcount":
		builder.RecordDatabricksSparkDatabricksDirectorycommitVacuumcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.directorycommit.vacuumerrors":
		builder.RecordDatabricksSparkDatabricksDirectorycommitVacuumerrorsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numchecks":
		builder.RecordDatabricksSparkDatabricksPreemptionNumchecksDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numpoolsautoexpired":
		builder.RecordDatabricksSparkDatabricksPreemptionNumpoolsautoexpiredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.numtaskspreempted":
		builder.RecordDatabricksSparkDatabricksPreemptionNumtaskspreemptedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.poolstarvationmillis":
		builder.RecordDatabricksSparkDatabricksPreemptionPoolstarvationmillisDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.scheduleroverheadnanos":
		builder.RecordDatabricksSparkDatabricksPreemptionScheduleroverheadnanosDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.preemption.tasktimewastedmillis":
		builder.RecordDatabricksSparkDatabricksPreemptionTasktimewastedmillisDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.activepools":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesActivepoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.bypasslaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesBypasslaneactivepoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.fastlaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesFastlaneactivepoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.finishedqueriestotaltasktimens":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesFinishedqueriestotaltasktimensDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.markedpools":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupMarkedpoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.twophasepoolscleaned":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupTwophasepoolscleanedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.lanecleanup.zombiepoolscleaned":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupZombiepoolscleanedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransfernumsuccessfulpreemptioniterations":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransfernumsuccessfulpreemptioniterationsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransfernumtaskspreempted":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransfernumtaskspreemptedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.preemption.slottransferwastedtasktimens":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransferwastedtasktimensDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numgradualdecrease":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumgradualdecreaseDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numquickdrop":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumquickdropDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.numquickjump":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumquickjumpDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slotreservation.slotsreserved":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationSlotsreservedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.slowlaneactivepools":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlowlaneactivepoolsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "databricks.taskschedulinglanes.totalquerygroupsfinished":
		builder.RecordDatabricksSparkDatabricksTaskschedulinglanesTotalquerygroupsfinishedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.filecachehits":
		builder.RecordDatabricksSparkHiveexternalcatalogFilecachehitsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.filesdiscovered":
		builder.RecordDatabricksSparkHiveexternalcatalogFilesdiscoveredDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.hiveclientcalls":
		builder.RecordDatabricksSparkHiveexternalcatalogHiveclientcallsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.parallellistingjobcount":
		builder.RecordDatabricksSparkHiveexternalcatalogParallellistingjobcountDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "hiveexternalcatalog.partitionsfetched":
		builder.RecordDatabricksSparkHiveexternalcatalogPartitionsfetchedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.numeventsposted":
		builder.RecordDatabricksSparkLivelistenerbusNumeventspostedDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.appstatus.numdroppedevents":
		builder.RecordDatabricksSparkLivelistenerbusQueueAppstatusNumdroppedeventsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.executormanagement.numdroppedevents":
		builder.RecordDatabricksSparkLivelistenerbusQueueExecutormanagementNumdroppedeventsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.shared.numdroppedevents":
		builder.RecordDatabricksSparkLivelistenerbusQueueSharedNumdroppedeventsDataPoint(
			now,
			int64(counter.Count),
			clusterID,
			appID,
			pipelineID,
			pipelineName,
		)
	case "livelistenerbus.queue.streams.numdroppedevents":
		builder.RecordDatabricksSparkLivelistenerbusQueueStreamsNumdroppedeventsDataPoint(
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
	timer spark.Timer,
	rs sparkResource,
) {
	appID := rs.appID
	clusterID := rs.cluster.ClusterID
	switch partialMetricName {
	case "dagscheduler.messageprocessingtime":
		builder.RecordDatabricksSparkTimerDagschedulerMessageprocessingtimeMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.backend.daemon.driver.dbceventlogginglistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksBackendDaemonDriverDbceventlogginglistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.backend.daemon.driver.dataplaneeventlistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksBackendDaemonDriverDataplaneeventlistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.photon.photoncleanuplistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksPhotonPhotoncleanuplistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.spark.util.executortimelogginglistener$":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSparkUtilExecutortimelogginglistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.spark.util.usagelogginglistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSparkUtilUsagelogginglistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.advice.advisorlistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSQLAdviceAdvisorlistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.debugger.querywatchdoglistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSQLDebuggerQuerywatchdoglistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.execution.ui.iocachelistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSQLExecutionUIIocachelistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.com.databricks.sql.io.caching.repeatedreadsestimator$":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeDatabricksSQLIoCachingRepeatedreadsestimatorMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.sparksession$$anon$1":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLSparksessionMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.sqlexecution$":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLExecutionSqlexecutionMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.streaming.streamingquerylistenerbus":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLExecutionStreamingStreamingquerylistenerbusMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.execution.ui.sqlappstatuslistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLExecutionUISqlappstatuslistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.hive.thriftserver.ui.hivethriftserver2listener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLHiveThriftserverUIHivethriftserver2listenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.sql.util.executionlistenerbus":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkSQLUtilExecutionlistenerbusMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.status.appstatuslistener":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkStatusAppstatuslistenerMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.listenerprocessingtime.org.apache.spark.util.profilerenv$$anon$1":
		builder.RecordDatabricksSparkTimerLivelistenerbusListenerprocessingtimeApacheSparkUtilProfilerenvMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.appstatus.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLivelistenerbusQueueAppstatusListenerprocessingtimeMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.executormanagement.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLivelistenerbusQueueExecutormanagementListenerprocessingtimeMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.shared.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLivelistenerbusQueueSharedListenerprocessingtimeMeanDataPoint(
			now,
			timer.Mean,
			clusterID,
			appID,
		)
	case "livelistenerbus.queue.streams.listenerprocessingtime":
		builder.RecordDatabricksSparkTimerLivelistenerbusQueueStreamsListenerprocessingtimeMeanDataPoint(
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
	histo spark.Histogram,
	rs sparkResource,
	pipelineID string,
	pipelineName string,
) {
	appID := rs.appID
	clusterID := rs.cluster.ClusterID
	switch partialMetricName {
	case "codegenerator.compilationtime":
		builder.RecordDatabricksSparkCodegeneratorCompilationtimeMeanDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.generatedclasssize":
		builder.RecordDatabricksSparkCodegeneratorGeneratedclasssizeMeanDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.generatedmethodsize":
		builder.RecordDatabricksSparkCodegeneratorGeneratedmethodsizeMeanDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	case "codegenerator.sourcecodesize":
		builder.RecordDatabricksSparkCodegeneratorSourcecodesizeMeanDataPoint(now, histo.Mean, clusterID, appID, pipelineID, pipelineName)
	}
}

func buildExecMetrics(builder *metadata.MetricsBuilder, execInfo spark.ExecutorInfo, now pcommon.Timestamp, rs sparkResource) {
	builder.RecordDatabricksSparkExecutorMemoryUsedDataPoint(now, int64(execInfo.MemoryUsed), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorDiskUsedDataPoint(now, int64(execInfo.DiskUsed), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalInputBytesDataPoint(now, execInfo.TotalInputBytes, rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalShuffleReadDataPoint(now, int64(execInfo.TotalShuffleRead), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorTotalShuffleWriteDataPoint(now, int64(execInfo.TotalShuffleWrite), rs.cluster.ClusterID, rs.appID, execInfo.ID)
	builder.RecordDatabricksSparkExecutorMaxMemoryDataPoint(now, execInfo.MaxMemory, rs.cluster.ClusterID, rs.appID, execInfo.ID)
}

func buildJobMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, jobInfo spark.JobInfo, rs sparkResource) {
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

func buildStageMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, stageInfo spark.StageInfo, rs sparkResource) {
	builder.RecordDatabricksSparkStageExecutorRunTimeDataPoint(now, int64(stageInfo.ExecutorRunTime), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageInputBytesDataPoint(now, int64(stageInfo.InputBytes), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageInputRecordsDataPoint(now, int64(stageInfo.InputRecords), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageOutputBytesDataPoint(now, int64(stageInfo.OutputBytes), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageOutputRecordsDataPoint(now, int64(stageInfo.OutputRecords), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageMemoryBytesSpilledDataPoint(now, int64(stageInfo.MemoryBytesSpilled), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
	builder.RecordDatabricksSparkStageDiskBytesSpilledDataPoint(now, int64(stageInfo.DiskBytesSpilled), rs.cluster.ClusterID, rs.appID, int64(stageInfo.StageID))
}

func newSparkMetricBase(partialMetricName string, pipeline *pipelineSummary) sparkMetricBase {
	pipelineID := ""
	pipelineName := ""
	if pipeline != nil {
		pipelineID = pipeline.id
		pipelineName = pipeline.name
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
	gauge spark.Gauge
}

func (i gaugeInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildGauge(builder, i.partialMetricName, now, i.gauge, rs, i.pipelineID, i.pipelineName)
}

type counterInfo struct {
	sparkMetricBase
	counter spark.Counter
}

func (i counterInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildCounter(builder, i.partialMetricName, now, i.counter, rs, i.pipelineID, i.pipelineName)
}

type timerInfo struct {
	sparkMetricBase
	timer spark.Timer
}

func (i timerInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildTimers(builder, i.partialMetricName, now, i.timer, rs)
}

type histoInfo struct {
	sparkMetricBase
	histo spark.Histogram
}

func (i histoInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildHistos(builder, i.partialMetricName, now, i.histo, rs, i.pipelineID, i.pipelineName)
}

type execInfo struct {
	execInfo spark.ExecutorInfo
}

func (i execInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildExecMetrics(builder, i.execInfo, now, rs)
}

type jobInfo struct {
	jobInfo spark.JobInfo
}

func (i jobInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildJobMetrics(builder, now, i.jobInfo, rs)
}

type stageInfo struct {
	stageInfo spark.StageInfo
}

func (i stageInfo) build(builder *metadata.MetricsBuilder, rs sparkResource, now pcommon.Timestamp) {
	buildStageMetrics(builder, now, i.stageInfo, rs)
}
