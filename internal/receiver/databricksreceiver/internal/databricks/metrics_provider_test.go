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

package databricks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
)

func TestDatabricksMetricsProvider(t *testing.T) {
	var dbrsvc = NewService(&testdataRawClient{testDataDir: commontest.TestdataDir}, 25)
	mp := MetricsProvider{Svc: dbrsvc}

	builder := commontest.NewTestMetricsBuilder()
	_, err := mp.AddJobStatusMetrics(builder, 0)
	require.NoError(t, err)
	emitted := builder.Emit()
	assert.Equal(t, 3, emitted.MetricCount())

	metricMap := commontest.MetricsByName(emitted)
	rms := emitted.ResourceMetrics()
	assert.Equal(t, 1, rms.Len())
	rm := rms.At(0)
	sms := rm.ScopeMetrics()

	assert.Equal(t, 1, sms.Len())

	const dbrjt = "databricks.jobs.total"
	jobTotalMetrics := metricMap[dbrjt]
	assert.Equal(t, dbrjt, jobTotalMetrics.Name())
	assert.EqualValues(t, 6, jobTotalMetrics.Gauge().DataPoints().At(0).IntValue())

	const dbrjss = "databricks.jobs.schedule.status"
	jobScheduleMetrics := metricMap[dbrjss]
	assert.Equal(t, dbrjss, jobScheduleMetrics.Name())
	pts := jobScheduleMetrics.Gauge().DataPoints()
	assert.Equal(t, 6, pts.Len())
	assert.EqualValues(t, 0, pts.At(0).IntValue())

	const dbrtss = "databricks.tasks.schedule.status"
	taskStatusMetric := metricMap[dbrtss]
	assert.Equal(t, dbrtss, taskStatusMetric.Name())
	taskPts := taskStatusMetric.Gauge().DataPoints()
	assert.Equal(t, 8, taskPts.Len())

	taskMap := tasksToMap(taskPts)

	job7Tasks := taskMap[7]
	{
		pt := job7Tasks["user2test"]
		taskAttrs := pt.Attributes()
		jobIDAttr, _ := taskAttrs.Get("job.id")
		assert.EqualValues(t, 7, jobIDAttr.Int())
		taskIDAttr, _ := taskAttrs.Get("task.id")
		assert.EqualValues(t, "user2test", taskIDAttr.Str())
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "NotebookTask", taskTypeAttr.Str())
	}
	{
		pt := job7Tasks["multi"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "SparkPythonTask", taskTypeAttr.Str())
	}

	{
		job102Tasks := taskMap[102]
		pt := job102Tasks["test"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "SparkJarTask", taskTypeAttr.Str())
	}

	{
		job179Tasks := taskMap[179]
		pt := job179Tasks["singletask"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "PipelineTask", taskTypeAttr.Str())
	}

	job248Tasks := taskMap[248]
	{
		pt := job248Tasks["dash"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "PythonWheelTask", taskTypeAttr.Str())
	}
	{
		pt := job248Tasks["user2test"]
		taskAttrs := pt.Attributes()
		taskTypeAttr, _ := taskAttrs.Get("task.type")
		assert.Equal(t, "SparkSubmitTask", taskTypeAttr.Str())
	}

	err = mp.AddNumActiveRunsMetric(builder, 0)
	require.NoError(t, err)

	emitted = builder.Emit()
	rms = emitted.ResourceMetrics()
	assert.Equal(t, 1, rms.Len())

	assert.Equal(t, 1, rms.Len())
	rm = rms.At(0)
	sms = rm.ScopeMetrics()
	assert.Equal(t, 1, sms.Len())
	ms := sms.At(0).Metrics()

	activeRunsMetric := ms.At(0)
	assert.Equal(t, "databricks.jobs.active.total", activeRunsMetric.Name())
	assert.Equal(t, 1, activeRunsMetric.Gauge().DataPoints().Len())
}

func tasksToMap(tasks pmetric.NumberDataPointSlice) map[int64]map[string]pmetric.NumberDataPoint {
	jobMap := map[int64]map[string]pmetric.NumberDataPoint{}
	for i := 0; i < tasks.Len(); i++ {
		task := tasks.At(i)
		attrs := task.Attributes()
		jobIDAttr, _ := attrs.Get("job.id")
		jobID := jobIDAttr.Int()
		taskMap, ok := jobMap[jobID]
		if !ok {
			taskMap = map[string]pmetric.NumberDataPoint{}
			jobMap[jobID] = taskMap
		}
		taskIDAttr, _ := attrs.Get("task.id")
		taskMap[taskIDAttr.Str()] = task
	}
	return jobMap
}
