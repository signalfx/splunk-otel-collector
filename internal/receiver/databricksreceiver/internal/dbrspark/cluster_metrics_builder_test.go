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
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/databricks"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

var testdataDir = filepath.Join("..", "..", "testdata")

func TestStripSparkMetricKey(t *testing.T) {
	key := "app-20221117221047-0000.driver.BlockManager.memory.diskSpaceUsed_MB"
	appID, stripped := stripSparkMetricKey(key)
	assert.Equal(t, "app-20221117221047-0000", appID)
	assert.Equal(t, "blockmanager.memory.diskspaceused", stripped)
}

func TestClusterMetricsBuilder_GeneratedMetrics(t *testing.T) {
	mb := ClusterMetricsBuilder{
		Ssvc: NewTestSuccessSparkService(testdataDir),
	}
	coreMetrics, err := mb.BuildCoreMetrics([]spark.Cluster{{
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

	metricMap := databricks.MetricsByName(pm)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.diskspaceused", 42)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxmem", 123)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxoffheapmem", 111)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.maxonheapmem", 222)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.memused", 333)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.offheapmemused", 444)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.onheapmemused", 555)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingmem", 666)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingoffheapmem", 777)
	assertSparkDoubleGaugeEq(t, metricMap, "blockmanager.memory.remainingonheapmem", 888)
	assertSparkDoubleGaugeEq(t, metricMap, "dagscheduler.job.activejobs", 999)
	assertSparkDoubleGaugeEq(t, metricMap, "dagscheduler.job.alljobs", 1111)
	assertSparkDoubleGaugeEq(t, metricMap, "dagscheduler.stage.failedstages", 2222)
	assertSparkDoubleGaugeEq(t, metricMap, "dagscheduler.stage.runningstages", 3333)
	assertSparkDoubleGaugeEq(t, metricMap, "dagscheduler.stage.waitingstages", 4444)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.directpoolmemory", 591058)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.jvmheapmemory", 1748700144)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.jvmoffheapmemory", 269709952)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.majorgccount", 5)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.majorgctime", 748)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.mappedpoolmemory", 5555)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.minorgccount", 5)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.minorgctime", 200)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.offheapexecutionmemory", 6666)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.offheapstoragememory", 7777)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.offheapunifiedmemory", 8888)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.onheapexecutionmemory", 9999)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.onheapstoragememory", 11111)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.onheapunifiedmemory", 22222)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreejvmrssmemory", 33333)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreejvmvmemory", 44444)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreeotherrssmemory", 55555)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreeothervmemory", 66666)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreepythonrssmemory", 77777)
	assertSparkDoubleGaugeEq(t, metricMap, "executormetrics.processtreepythonvmemory", 88888)
	assertSparkDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.appstatus.size", 99999)
	assertSparkDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.executormanagement.size", 111111)
	assertSparkDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.shared.size", 222222)
	assertSparkDoubleGaugeEq(t, metricMap, "livelistenerbus.queue.streams.size", 333333)
	assertSparkDoubleGaugeEq(t, metricMap, "sparksqloperationmanager.numhiveoperations", 444444)

	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.autovacuumcount", 1)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.deletedfilesfiltered", 2)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.filterlistingcount", 3)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.jobcommitcompleted", 4)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.markerreaderrors", 5)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.markerrefreshcount", 6)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.markerrefresherrors", 7)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.markersread", 8)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.repeatedlistcount", 9)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.uncommittedfilesfiltered", 10)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.untrackedfilesfound", 11)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.vacuumcount", 12)
	assertSparkIntSumEq(t, metricMap, "databricks.directorycommit.vacuumerrors", 13)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.numchecks", 14)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.numpoolsautoexpired", 15)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.numtaskspreempted", 16)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.poolstarvationmillis", 17)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.scheduleroverheadnanos", 18)
	assertSparkIntSumEq(t, metricMap, "databricks.preemption.tasktimewastedmillis", 19)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.activepools", 20)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.bypasslaneactivepools", 21)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.fastlaneactivepools", 22)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.finishedqueriestotaltasktimens", 23)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.markedpools", 24)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.twophasepoolscleaned", 25)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.lanecleanup.zombiepoolscleaned", 26)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransfernumsuccessfulpreemptioniterations", 27)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransfernumtaskspreempted", 28)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.preemption.slottransferwastedtasktimens", 29)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numgradualdecrease", 30)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numquickdrop", 31)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.numquickjump", 32)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slotreservation.slotsreserved", 33)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.slowlaneactivepools", 34)
	assertSparkIntSumEq(t, metricMap, "databricks.taskschedulinglanes.totalquerygroupsfinished", 35)
	assertSparkIntSumEq(t, metricMap, "hiveexternalcatalog.filecachehits", 36)
	assertSparkIntSumEq(t, metricMap, "hiveexternalcatalog.filesdiscovered", 37)
	assertSparkIntSumEq(t, metricMap, "hiveexternalcatalog.hiveclientcalls", 38)
	assertSparkIntSumEq(t, metricMap, "hiveexternalcatalog.parallellistingjobcount", 39)
	assertSparkIntSumEq(t, metricMap, "hiveexternalcatalog.partitionsfetched", 40)
	assertSparkIntSumEq(t, metricMap, "livelistenerbus.numeventsposted", 41)
	assertSparkIntSumEq(t, metricMap, "livelistenerbus.queue.appstatus.numdroppedevents", 42)
	assertSparkIntSumEq(t, metricMap, "livelistenerbus.queue.executormanagement.numdroppedevents", 43)
	assertSparkIntSumEq(t, metricMap, "livelistenerbus.queue.shared.numdroppedevents", 44)
	assertSparkIntSumEq(t, metricMap, "livelistenerbus.queue.streams.numdroppedevents", 45)

	assertSparkDoubleSumEq(t, metricMap, "jvmcpu.jvmcputime", 57690000000)
	assertSparkDoubleSumEq(t, metricMap, "timer.dagscheduler.messageprocessingtime", 1.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.backend.daemon.driver.dbceventlogginglistener", 1.2)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.backend.daemon.driver.dataplaneeventlistener", 1.3)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.photon.photoncleanuplistener", 1.4)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.spark.util.executortimelogginglistener", 1.5)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.spark.util.usagelogginglistener", 1.6)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.sql.advice.advisorlistener", 1.7)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.sql.debugger.querywatchdoglistener", 1.8)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.sql.execution.ui.iocachelistener", 1.9)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.databricks.sql.io.caching.repeatedreadsestimator", 2.0)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.sparksession", 2.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.sqlexecution", 2.2)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.streaming.streamingquerylistenerbus", 2.3)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.execution.ui.sqlappstatuslistener", 2.4)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.hive.thriftserver.ui.hivethriftserver2listener", 2.5)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.sql.util.executionlistenerbus", 2.6)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.status.appstatuslistener", 2.7)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.listenerprocessingtime.apache.spark.util.profilerenv", 2.8)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.queue.appstatus.listenerprocessingtime", 2.9)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.queue.executormanagement.listenerprocessingtime", 3.0)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.queue.shared.listenerprocessingtime", 3.1)
	assertSparkDoubleSumEq(t, metricMap, "timer.livelistenerbus.queue.streams.listenerprocessingtime", 3.2)
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
