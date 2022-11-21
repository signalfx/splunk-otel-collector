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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestStripSparkMetricKey(t *testing.T) {
	key := "app-20221117221047-0000.driver.BlockManager.memory.diskSpaceUsed_MB"
	appID, stripped := stripSparkMetricKey(key)
	assert.Equal(t, "app-20221117221047-0000", appID)
	assert.Equal(t, "blockmanager.memory.diskspaceused", stripped)
}

func TestSparkMetricsBuilder_GeneratedMetrics(t *testing.T) {
	mp := sparkCoreMetricsBuilder{newTestSparkService()}
	builder := newTestMetricsBuilder()
	_, _, err := mp.buildCoreMetrics(builder, 0)
	require.NoError(t, err)
	emitted := builder.Emit()
	ms := emitted.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	metricMap := metricsByName(ms)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.diskspaceused", 42)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxmem", 123)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxoffheapmem", 111)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxonheapmem", 222)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.memused", 333)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.offheapmemused", 444)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.onheapmemused", 555)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingmem", 666)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingoffheapmem", 777)
	assertDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingonheapmem", 888)
	assertDoubleGaugeEq(t, metricMap, "dagscheduler.job.activejobs", 999)
	assertDoubleGaugeEq(t, metricMap, "dagscheduler.job.alljobs", 1111)
	assertDoubleGaugeEq(t, metricMap, "dagscheduler.stage.failedstages", 2222)
	assertDoubleGaugeEq(t, metricMap, "dagscheduler.stage.runningstages", 3333)
	assertDoubleGaugeEq(t, metricMap, "dagscheduler.stage.waitingstages", 4444)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.directpoolmemory", 591058)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.jvmheapmemory", 1748700144)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.jvmoffheapmemory", 269709952)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.majorgccount", 5)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.majorgctime", 748)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.mappedpoolmemory", 5555)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.minorgccount", 5)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.minorgctime", 200)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.offheapexecutionmemory", 6666)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.offheapstoragememory", 7777)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.offheapunifiedmemory", 8888)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.onheapexecutionmemory", 9999)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.onheapstoragememory", 11111)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.onheapunifiedmemory", 22222)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreejvmrssmemory", 33333)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreejvmvmemory", 44444)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreeotherrssmemory", 55555)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreeothervmemory", 66666)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreepythonrssmemory", 77777)
	assertDoubleGaugeEq(t, metricMap, "executormetrics.processtreepythonvmemory", 88888)
	assertDoubleGaugeEq(t, metricMap, "jvmcpu.jvmcputime", 57690000000)
	assertDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.appstatus.size", 99999)
	assertDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.executormanagement.size", 111111)
	assertDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.shared.size", 222222)
	assertDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.streams.size", 333333)
	assertDoubleGaugeEq(t, metricMap, "sparksqloperationmanager.numhiveoperations", 444444)

	assertIntSumEq(t, metricMap, "databricks.directorycommit.autovacuumcount", 1)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.deletedfilesfiltered", 2)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.filterlistingcount", 3)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.jobcommitcompleted", 4)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.markerreaderrors", 5)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.markerrefreshcount", 6)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.markerrefresherrors", 7)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.markersread", 8)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.repeatedlistcount", 9)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.uncommittedfilesfiltered", 10)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.untrackedfilesfound", 11)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.vacuumcount", 12)
	assertIntSumEq(t, metricMap, "databricks.directorycommit.vacuumerrors", 13)
	assertIntSumEq(t, metricMap, "databricks.preemption.numchecks", 14)
	assertIntSumEq(t, metricMap, "databricks.preemption.numpoolsautoexpired", 15)
	assertIntSumEq(t, metricMap, "databricks.preemption.numtaskspreempted", 16)
	assertIntSumEq(t, metricMap, "databricks.preemption.poolstarvationmillis", 17)
	assertIntSumEq(t, metricMap, "databricks.preemption.scheduleroverheadnanos", 18)
	assertIntSumEq(t, metricMap, "databricks.preemption.tasktimewastedmillis", 19)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.activepools", 20)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.bypasslaneactivepools", 21)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.fastlaneactivepools", 22)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.finishedqueriestotaltasktimens", 23)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.markedpools", 24)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.twophasepoolscleaned", 25)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.zombiepoolscleaned", 26)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransfernumsuccessfulpreemptioniterations", 27)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransfernumtaskspreempted", 28)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransferwastedtasktimens", 29)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numgradualdecrease", 30)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numquickdrop", 31)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numquickjump", 32)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.slotsreserved", 33)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slowlaneactivepools", 34)
	assertIntSumEq(t, metricMap, "databricks.taskschedulinglanes.totalquerygroupsfinished", 35)
	assertIntSumEq(t, metricMap, "hiveexternalcatalog.filecachehits", 36)
	assertIntSumEq(t, metricMap, "hiveexternalcatalog.filesdiscovered", 37)
	assertIntSumEq(t, metricMap, "hiveexternalcatalog.hiveclientcalls", 38)
	assertIntSumEq(t, metricMap, "hiveexternalcatalog.parallellistingjobcount", 39)
	assertIntSumEq(t, metricMap, "hiveexternalcatalog.partitionsfetched", 40)
	assertIntSumEq(t, metricMap, "livelistenerbus.numeventsposted", 41)
	assertIntSumEq(t, metricMap, "livelistenerbus.queue.appstatus.numdroppedevents", 42)
	assertIntSumEq(t, metricMap, "livelistenerbus.queue.executormanagement.numdroppedevents", 43)
	assertIntSumEq(t, metricMap, "livelistenerbus.queue.shared.numdroppedevents", 44)
	assertIntSumEq(t, metricMap, "livelistenerbus.queue.streams.numdroppedevents", 45)

	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.dagscheduler.messageprocessingtime.mean", 1.1)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.backend.daemon.driver.dbceventlogginglistener.mean", 1.2)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.backend.daemon.driver.dataplaneeventlistener.mean", 1.3)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.photon.photoncleanuplistener.mean", 1.4)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.spark.util.executortimelogginglistener.mean", 1.5)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.spark.util.usagelogginglistener.mean", 1.6)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.sql.advice.advisorlistener.mean", 1.7)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.sql.debugger.querywatchdoglistener.mean", 1.8)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.sql.execution.ui.iocachelistener.mean", 1.9)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.databricks.sql.io.caching.repeatedreadsestimator.mean", 2.0)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.sparksession.mean", 2.1)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.sqlexecution.mean", 2.2)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.streaming.streamingquerylistenerbus.mean", 2.3)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.ui.sqlappstatuslistener.mean", 2.4)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.hive.thriftserver.ui.hivethriftserver2listener.mean", 2.5)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.util.executionlistenerbus.mean", 2.6)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.status.appstatuslistener.mean", 2.7)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.listenerprocessingtime.apache.spark.util.profilerenv.mean", 2.8)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.queue.appstatus.listenerprocessingtime.mean", 2.9)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.queue.executormanagement.listenerprocessingtime.mean", 3.0)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.queue.shared.listenerprocessingtime.mean", 3.1)
	assertDoubleSumEq(t, metricMap, "databricks.spark.timer.livelistenerbus.queue.streams.listenerprocessingtime.mean", 3.2)
}

func assertIntGaugeEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expectedSlice ...int) {
	for i, expected := range expectedSlice {
		assert.EqualValues(t, expected, metricMap["databricks.spark."+metricName].Gauge().DataPoints().At(i).IntValue())
	}
}

func assertDoubleGaugeEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expected float64) {
	assert.Equal(t, expected, metricMap["databricks.spark."+metricName].Gauge().DataPoints().At(0).DoubleValue())
}

func assertDoubleSumEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expected float64) {
	metric, ok := metricMap[metricName]
	if !ok {
		t.Errorf("metric not found: %q", metricName)
		return
	}
	assert.EqualValues(t, expected, metric.Sum().DataPoints().At(0).DoubleValue())
}

func assertIntSumEq(t *testing.T, metricMap map[string]pmetric.Metric, metricName string, expectedSlice ...int) {
	for i, expected := range expectedSlice {
		assert.EqualValues(t, expected, metricMap["databricks.spark."+metricName].Sum().DataPoints().At(i).IntValue())
	}
}

func TestSparkMetricsBuilder_Histograms(t *testing.T) {
	mp := sparkCoreMetricsBuilder{newTestSparkService()}
	builder := newTestMetricsBuilder()
	histoMetrics, _, err := mp.buildCoreMetrics(builder, 0)
	require.NoError(t, err)
	ms := pmetric.NewMetricSlice()
	for _, metric := range histoMetrics {
		metric.CopyTo(ms.AppendEmpty())
	}
	metricMap := metricsByName(ms)
	metric := metricMap["databricks.spark.codegenerator.compilationtime"]
	histogram := metric.Histogram()
	dps := histogram.DataPoints()
	pt := dps.At(0)
	assert.EqualValues(t, 3, pt.Count())
	assert.EqualValues(t, 14, pt.Min())
	assert.EqualValues(t, 288, pt.Max())
	assert.EqualValues(t, 3*106, pt.Sum())
	bounds := pt.ExplicitBounds().AsRaw()
	assert.Equal(t, []float64{50, 75, 95, 98, 99, 999}, bounds)
	bucketCounts := pt.BucketCounts().AsRaw()
	assert.Equal(t, []uint64{22, 284, 285, 286, 287, 288}, bucketCounts)
}
