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
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

type sparkCoreMetricsBuilder struct {
	ssvc sparkService
}

func (b sparkCoreMetricsBuilder) buildCoreMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp) ([]pmetric.Metric, []string, error) {
	coreClusterMetrics, err := b.ssvc.getSparkCoreMetricsForAllClusters()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting spark metrics for all clusters: %w", err)
	}
	var histoMetrics []pmetric.Metric
	var clusterIDs []string
	for clstr, clusterMetric := range coreClusterMetrics {
		clusterID := clstr.ClusterId
		clusterIDs = append(clusterIDs, clusterID)
		b.buildClusterMetrics(builder, clusterMetric, now, clusterID)
		b.buildClusterTimers(builder, clusterMetric, now, clstr.ClusterName)
		otelHistos := b.sparkClusterHistosToOtelHistos(clusterMetric, now, clusterID)
		histoMetrics = append(histoMetrics, otelHistos...)
	}
	return histoMetrics, clusterIDs, nil
}

func (b sparkCoreMetricsBuilder) buildClusterMetrics(
	builder *metadata.MetricsBuilder,
	m spark.ClusterMetrics,
	now pcommon.Timestamp,
	clusterID string,
) {
	for key, gauge := range m.Gauges {
		appID, stripped := stripSparkMetricKey(key)
		switch stripped {
		case "blockmanager.memory.diskspaceused":
			builder.RecordDatabricksSparkBlockmanagerMemoryDiskspaceusedDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.maxmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryMaxmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.maxoffheapmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryMaxoffheapmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.maxonheapmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryMaxonheapmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.memused":
			builder.RecordDatabricksSparkBlockmanagerMemoryMemusedDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.offheapmemused":
			builder.RecordDatabricksSparkBlockmanagerMemoryOffheapmemusedDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.onheapmemused":
			builder.RecordDatabricksSparkBlockmanagerMemoryOnheapmemusedDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.remainingmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryRemainingmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.remainingoffheapmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryRemainingoffheapmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "blockmanager.memory.remainingonheapmem":
			builder.RecordDatabricksSparkBlockmanagerMemoryRemainingonheapmemDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "dagscheduler.job.activejobs":
			builder.RecordDatabricksSparkDagschedulerJobActivejobsDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "dagscheduler.job.alljobs":
			builder.RecordDatabricksSparkDagschedulerJobAlljobsDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "dagscheduler.stage.failedstages":
			builder.RecordDatabricksSparkDagschedulerStageFailedstagesDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "dagscheduler.stage.runningstages":
			builder.RecordDatabricksSparkDagschedulerStageRunningstagesDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "dagscheduler.stage.waitingstages":
			builder.RecordDatabricksSparkDagschedulerStageWaitingstagesDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.directpoolmemory":
			builder.RecordDatabricksSparkExecutormetricsDirectpoolmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.jvmheapmemory":
			builder.RecordDatabricksSparkExecutormetricsJvmheapmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.jvmoffheapmemory":
			builder.RecordDatabricksSparkExecutormetricsJvmoffheapmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.majorgccount":
			builder.RecordDatabricksSparkExecutormetricsMajorgccountDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.majorgctime":
			builder.RecordDatabricksSparkExecutormetricsMajorgctimeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.mappedpoolmemory":
			builder.RecordDatabricksSparkExecutormetricsMappedpoolmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.minorgccount":
			builder.RecordDatabricksSparkExecutormetricsMinorgccountDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.minorgctime":
			builder.RecordDatabricksSparkExecutormetricsMinorgctimeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.offheapexecutionmemory":
			builder.RecordDatabricksSparkExecutormetricsOffheapexecutionmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.offheapstoragememory":
			builder.RecordDatabricksSparkExecutormetricsOffheapstoragememoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.offheapunifiedmemory":
			builder.RecordDatabricksSparkExecutormetricsOffheapunifiedmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.onheapexecutionmemory":
			builder.RecordDatabricksSparkExecutormetricsOnheapexecutionmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.onheapstoragememory":
			builder.RecordDatabricksSparkExecutormetricsOnheapstoragememoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.onheapunifiedmemory":
			builder.RecordDatabricksSparkExecutormetricsOnheapunifiedmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreejvmrssmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreejvmrssmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreejvmvmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreejvmvmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreeotherrssmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreeotherrssmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreeothervmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreeothervmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreepythonrssmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreepythonrssmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "executormetrics.processtreepythonvmemory":
			builder.RecordDatabricksSparkExecutormetricsProcesstreepythonvmemoryDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "jvmcpu.jvmcputime":
			builder.RecordDatabricksSparkJvmcpuJvmcputimeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.appstatus.size":
			builder.RecordDatabricksSparkLivelistenerbusQueueAppstatusSizeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.executormanagement.size":
			builder.RecordDatabricksSparkLivelistenerbusQueueExecutormanagementSizeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.shared.size":
			builder.RecordDatabricksSparkLivelistenerbusQueueSharedSizeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.streams.size":
			builder.RecordDatabricksSparkLivelistenerbusQueueStreamsSizeDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		case "sparksqloperationmanager.numhiveoperations":
			builder.RecordDatabricksSparkSparksqloperationmanagerNumhiveoperationsDataPoint(
				now,
				gauge.Value,
				clusterID,
				appID,
			)
		}
	}
	for key, counter := range m.Counters {
		appID, stripped := stripSparkMetricKey(key)
		switch stripped {
		case "databricks.directorycommit.autovacuumcount":
			builder.RecordDatabricksSparkDatabricksDirectorycommitAutovacuumcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.deletedfilesfiltered":
			builder.RecordDatabricksSparkDatabricksDirectorycommitDeletedfilesfilteredDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.filterlistingcount":
			builder.RecordDatabricksSparkDatabricksDirectorycommitFilterlistingcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.jobcommitcompleted":
			builder.RecordDatabricksSparkDatabricksDirectorycommitJobcommitcompletedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.markerreaderrors":
			builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerreaderrorsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.markerrefreshcount":
			builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerrefreshcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.markerrefresherrors":
			builder.RecordDatabricksSparkDatabricksDirectorycommitMarkerrefresherrorsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.markersread":
			builder.RecordDatabricksSparkDatabricksDirectorycommitMarkersreadDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.repeatedlistcount":
			builder.RecordDatabricksSparkDatabricksDirectorycommitRepeatedlistcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.uncommittedfilesfiltered":
			builder.RecordDatabricksSparkDatabricksDirectorycommitUncommittedfilesfilteredDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.untrackedfilesfound":
			builder.RecordDatabricksSparkDatabricksDirectorycommitUntrackedfilesfoundDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.vacuumcount":
			builder.RecordDatabricksSparkDatabricksDirectorycommitVacuumcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.directorycommit.vacuumerrors":
			builder.RecordDatabricksSparkDatabricksDirectorycommitVacuumerrorsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.numchecks":
			builder.RecordDatabricksSparkDatabricksPreemptionNumchecksDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.numpoolsautoexpired":
			builder.RecordDatabricksSparkDatabricksPreemptionNumpoolsautoexpiredDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.numtaskspreempted":
			builder.RecordDatabricksSparkDatabricksPreemptionNumtaskspreemptedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.poolstarvationmillis":
			builder.RecordDatabricksSparkDatabricksPreemptionPoolstarvationmillisDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.scheduleroverheadnanos":
			builder.RecordDatabricksSparkDatabricksPreemptionScheduleroverheadnanosDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.preemption.tasktimewastedmillis":
			builder.RecordDatabricksSparkDatabricksPreemptionTasktimewastedmillisDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.activepools":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesActivepoolsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.bypasslaneactivepools":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesBypasslaneactivepoolsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.fastlaneactivepools":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesFastlaneactivepoolsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.finishedqueriestotaltasktimens":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesFinishedqueriestotaltasktimensDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.lanecleanup.markedpools":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupMarkedpoolsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.lanecleanup.twophasepoolscleaned":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupTwophasepoolscleanedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.lanecleanup.zombiepoolscleaned":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesLanecleanupZombiepoolscleanedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.preemption.slottransfernumsuccessfulpreemptioniterations":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransfernumsuccessfulpreemptioniterationsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.preemption.slottransfernumtaskspreempted":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransfernumtaskspreemptedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.preemption.slottransferwastedtasktimens":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesPreemptionSlottransferwastedtasktimensDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.slotreservation.numgradualdecrease":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumgradualdecreaseDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.slotreservation.numquickdrop":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumquickdropDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.slotreservation.numquickjump":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationNumquickjumpDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.slotreservation.slotsreserved":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlotreservationSlotsreservedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.slowlaneactivepools":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesSlowlaneactivepoolsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "databricks.taskschedulinglanes.totalquerygroupsfinished":
			builder.RecordDatabricksSparkDatabricksTaskschedulinglanesTotalquerygroupsfinishedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "hiveexternalcatalog.filecachehits":
			builder.RecordDatabricksSparkHiveexternalcatalogFilecachehitsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "hiveexternalcatalog.filesdiscovered":
			builder.RecordDatabricksSparkHiveexternalcatalogFilesdiscoveredDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "hiveexternalcatalog.hiveclientcalls":
			builder.RecordDatabricksSparkHiveexternalcatalogHiveclientcallsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "hiveexternalcatalog.parallellistingjobcount":
			builder.RecordDatabricksSparkHiveexternalcatalogParallellistingjobcountDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "hiveexternalcatalog.partitionsfetched":
			builder.RecordDatabricksSparkHiveexternalcatalogPartitionsfetchedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "livelistenerbus.numeventsposted":
			builder.RecordDatabricksSparkLivelistenerbusNumeventspostedDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.appstatus.numdroppedevents":
			builder.RecordDatabricksSparkLivelistenerbusQueueAppstatusNumdroppedeventsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.executormanagement.numdroppedevents":
			builder.RecordDatabricksSparkLivelistenerbusQueueExecutormanagementNumdroppedeventsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.shared.numdroppedevents":
			builder.RecordDatabricksSparkLivelistenerbusQueueSharedNumdroppedeventsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		case "livelistenerbus.queue.streams.numdroppedevents":
			builder.RecordDatabricksSparkLivelistenerbusQueueStreamsNumdroppedeventsDataPoint(
				now,
				int64(counter.Count),
				clusterID,
				appID,
			)
		}
	}
}

func (b sparkCoreMetricsBuilder) sparkClusterHistosToOtelHistos(
	m spark.ClusterMetrics,
	now pcommon.Timestamp,
	clusterID string,
) []pmetric.Metric {
	var histoMetrics []pmetric.Metric
	for key, sparkHisto := range m.Histograms {
		appID, metricShortname := stripSparkMetricKey(key)
		histoMetrics = append(histoMetrics, sparkToOtelHisto(now, sparkHisto, appID, metricShortname))
	}
	return histoMetrics
}

func (b sparkCoreMetricsBuilder) buildClusterTimers(
	builder *metadata.MetricsBuilder,
	m spark.ClusterMetrics,
	now pcommon.Timestamp,
	clusterID string,
) {
	for key, timer := range m.Timers {
		appID, metricShortname := stripSparkMetricKey(key)
		switch metricShortname {
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
}

const sparkMetricPrefix = "databricks.spark."

func sparkToOtelHisto(ts pcommon.Timestamp, h spark.Histogram, appID string, name string) pmetric.Metric {
	metric := pmetric.NewMetric()
	metric.SetName(sparkMetricPrefix + name)
	histo := metric.SetEmptyHistogram()
	histo.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dps := histo.DataPoints()
	pt := dps.AppendEmpty()
	pt.SetTimestamp(ts)
	pt.SetSum(float64(h.Count) * h.Mean)
	pt.SetCount(uint64(h.Count))
	pt.SetMin(float64(h.Min))
	pt.SetMax(float64(h.Max))
	pt.ExplicitBounds().FromRaw([]float64{50, 75, 95, 98, 99, 999})
	pt.BucketCounts().FromRaw([]uint64{
		uint64(h.P50),
		uint64(h.P75),
		uint64(h.P95),
		uint64(h.P98),
		uint64(h.P99),
		uint64(h.P999),
	})
	attrs := pt.Attributes()
	attrs.PutStr("app_id", appID)
	attrs.PutStr("cluster_id", appID)
	return metric
}

func stripSparkMetricKey(s string) (string, string) {
	parts := strings.Split(s, ".")
	if len(parts) <= 2 || parts[1] != "driver" {
		return "", ""
	}
	metricParts := parts[2:]
	lastPart := metricParts[len(metricParts)-1]
	if strings.HasSuffix(lastPart, "_MB") {
		metricParts[len(metricParts)-1] = lastPart[:len(lastPart)-3]
	}
	joined := strings.Join(metricParts, ".")
	return parts[0], strings.ToLower(joined)
}
