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
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

func TestMetricsProvider(t *testing.T) {
	const ignored = 25
	var dbClient databricksServiceIntf = newDatabricksService(&testdataDBClient{}, ignored)
	mp := dbMetricsProvider{dbsvc: dbClient}

	builder := newTestMetricsBuilder()
	_, err := mp.addJobStatusMetrics(builder, 0)
	require.NoError(t, err)
	emitted := builder.Emit()
	assert.Equal(t, 3, emitted.MetricCount())

	rms := emitted.ResourceMetrics()
	assert.Equal(t, 1, rms.Len())
	rm := rms.At(0)
	sms := rm.ScopeMetrics()
	assert.Equal(t, 1, sms.Len())
	ms := sms.At(0).Metrics()

	metricMap := metricsByName(ms)

	const dbjt = "databricks.jobs.total"
	jobTotalMetrics := metricMap[dbjt]
	assert.Equal(t, dbjt, jobTotalMetrics.Name())
	assert.EqualValues(t, 6, jobTotalMetrics.Gauge().DataPoints().At(0).IntValue())

	const dbjss = "databricks.jobs.schedule.status"
	jobScheduleMetrics := metricMap[dbjss]
	assert.Equal(t, dbjss, jobScheduleMetrics.Name())
	pts := jobScheduleMetrics.Gauge().DataPoints()
	assert.Equal(t, 6, pts.Len())
	assert.EqualValues(t, 0, pts.At(0).IntValue())

	const dbtss = "databricks.tasks.schedule.status"
	taskStatusMetric := metricMap[dbtss]
	assert.Equal(t, dbtss, taskStatusMetric.Name())
	taskPts := taskStatusMetric.Gauge().DataPoints()
	assert.Equal(t, 8, taskPts.Len())

	taskMap := tasksToMap(taskPts)

	job7Tasks := taskMap[7]
	{
		pt := job7Tasks["user2test"]
		taskAttrs := pt.Attributes()
		jobIDAttr, _ := taskAttrs.Get("job_id")
		assert.EqualValues(t, 7, jobIDAttr.Int())
		taskIDAttr, _ := taskAttrs.Get("task_id")
		assert.EqualValues(t, "user2test", taskIDAttr.Str())
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "NotebookTask", taskTypeAttr.Str())
	}
	{
		pt := job7Tasks["multi"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "SparkPythonTask", taskTypeAttr.Str())
	}

	{
		job102Tasks := taskMap[102]
		pt := job102Tasks["test"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "SparkJarTask", taskTypeAttr.Str())
	}

	{
		job179Tasks := taskMap[179]
		pt := job179Tasks["singletask"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "PipelineTask", taskTypeAttr.Str())
	}

	job248Tasks := taskMap[248]
	{
		pt := job248Tasks["dash"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "PythonWheelTask", taskTypeAttr.Str())
	}
	{
		pt := job248Tasks["user2test"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task_type")
		assert.Equal(t, "SparkSubmitTask", taskTypeAttr.Str())
	}

	err = mp.addNumActiveRunsMetric(builder, 0)
	require.NoError(t, err)

	emitted = builder.Emit()
	rms = emitted.ResourceMetrics()
	assert.Equal(t, 1, rms.Len())

	assert.Equal(t, 1, rms.Len())
	rm = rms.At(0)
	sms = rm.ScopeMetrics()
	assert.Equal(t, 1, sms.Len())
	ms = sms.At(0).Metrics()

	activeRunsMetric := ms.At(0)
	assert.Equal(t, "databricks.jobs.active.total", activeRunsMetric.Name())
	assert.Equal(t, 1, activeRunsMetric.Gauge().DataPoints().Len())
}

func newTestMetricsBuilder() *metadata.MetricsBuilder {
	return metadata.NewMetricsBuilder(metadata.DefaultMetricsSettings(), component.BuildInfo{})
}

func metricsByName(ms pmetric.MetricSlice) map[string]pmetric.Metric {
	out := map[string]pmetric.Metric{}
	for i := 0; i < ms.Len(); i++ {
		metric := ms.At(i)
		out[metric.Name()] = metric
	}
	return out
}

func tasksToMap(tasks pmetric.NumberDataPointSlice) map[int64]map[string]pmetric.NumberDataPoint {
	jobMap := map[int64]map[string]pmetric.NumberDataPoint{}
	for i := 0; i < tasks.Len(); i++ {
		task := tasks.At(i)
		attrs := task.Attributes()
		jobIDAttr, _ := attrs.Get("job_id")
		jobID := jobIDAttr.Int()
		taskMap, ok := jobMap[jobID]
		if !ok {
			taskMap = map[string]pmetric.NumberDataPoint{}
			jobMap[jobID] = taskMap
		}
		taskIDAttr, _ := attrs.Get("task_id")
		taskMap[taskIDAttr.Str()] = task
	}
	return jobMap
}
