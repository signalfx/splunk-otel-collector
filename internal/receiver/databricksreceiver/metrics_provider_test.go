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

func TestMetricsProvider(t *testing.T) {
	const ignored = 25
	mp := newMetricsProvider(newDatabricksClient(&testdataClient{}, ignored))
	ms := pmetric.NewMetricSlice()
	_, err := mp.addJobStatusMetrics(ms)
	require.NoError(t, err)
	assert.Equal(t, ms.Len(), 3)

	jobTotalMetrics := ms.At(0)
	assert.Equal(t, "databricks.jobs.total", jobTotalMetrics.Name())
	assert.EqualValues(t, 6, jobTotalMetrics.Gauge().DataPoints().At(0).IntValue())

	jobScheduleMetrics := ms.At(1)
	assert.Equal(t, "databricks.jobs.schedule.status", jobScheduleMetrics.Name())
	pts := jobScheduleMetrics.Gauge().DataPoints()
	assert.Equal(t, 6, pts.Len())
	assert.EqualValues(t, 0, pts.At(0).IntValue())

	taskStatusMetric := ms.At(2)
	assert.Equal(t, "databricks.tasks.schedule.status", taskStatusMetric.Name())
	taskPts := taskStatusMetric.Gauge().DataPoints()
	assert.Equal(t, 8, taskPts.Len())

	task0Pt := taskPts.At(0)
	taskAttrs := task0Pt.Attributes()
	jobIDAttr, _ := taskAttrs.Get("job_id")
	assert.EqualValues(t, 7, jobIDAttr.Int())
	taskIDAttr, _ := taskAttrs.Get("task_id")
	assert.EqualValues(t, "user2test", taskIDAttr.Str())

	assertTaskTypeEquals(t, taskPts, 0, "NotebookTask")
	assertTaskTypeEquals(t, taskPts, 1, "SparkPythonTask")
	assertTaskTypeEquals(t, taskPts, 2, "SparkJarTask")
	assertTaskTypeEquals(t, taskPts, 3, "PipelineTask")
	assertTaskTypeEquals(t, taskPts, 4, "PythonWheelTask")
	assertTaskTypeEquals(t, taskPts, 5, "SparkSubmitTask")

	ms = pmetric.NewMetricSlice()
	err = mp.addNumActiveRunsMetric(ms)
	require.NoError(t, err)
	activeRunsMetric := ms.At(0)
	assert.Equal(t, "databricks.jobs.active.total", activeRunsMetric.Name())
	assert.Equal(t, 1, activeRunsMetric.Gauge().DataPoints().Len())
}

func assertTaskTypeEquals(t *testing.T, taskPts pmetric.NumberDataPointSlice, idx int, expected string) {
	tskType, _ := taskPts.At(idx).Attributes().Get("task_type")
	assert.EqualValues(t, expected, tskType.Str())
}
