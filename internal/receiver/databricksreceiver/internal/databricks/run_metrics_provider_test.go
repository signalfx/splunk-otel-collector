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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
)

func TestRunMetricProvider(t *testing.T) {
	p := NewRunMetricsProvider(&fakeDatabricksRestService{})
	builder := commontest.NewTestMetricsBuilder()
	err := p.addSingleJobRunMetrics(42, builder, 0)
	require.NoError(t, err)
	emitted := builder.Emit()
	assert.Equal(t, 0, emitted.MetricCount())
	assert.Equal(t, 0, emitted.DataPointCount())

	err = p.addSingleJobRunMetrics(42, builder, 0)
	require.NoError(t, err)
	emitted = builder.Emit()
	assert.Equal(t, 1, emitted.MetricCount())
	assert.Equal(t, 1, emitted.DataPointCount())
	metric := emitted.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	assert.EqualValues(t, 16000, metric.Gauge().DataPoints().At(0).IntValue())
}

func TestRunMetricsProvider_AddJobRunDurationMetrics(t *testing.T) {
	const ignored = 25
	mp := NewRunMetricsProvider(NewDatabricksService(&testdataDBRawClient{testDataDir: testdataDir}, ignored))
	builder := commontest.NewTestMetricsBuilder()
	err := mp.AddMultiJobRunMetrics([]int{TestdataJobID}, builder, 0)
	require.NoError(t, err)

	emitted := builder.Emit()
	assert.Equal(t, 0, emitted.MetricCount())
	assert.Equal(t, 0, emitted.DataPointCount())

	err = mp.AddMultiJobRunMetrics([]int{TestdataJobID}, builder, 0)
	require.NoError(t, err)
	emitted = builder.Emit()
	metricMap := MetricsByName(emitted)
	assert.Equal(t, 2, emitted.MetricCount())
	assert.Equal(t, 2, emitted.DataPointCount())

	jobMetric := metricMap["databricks.jobs.run.duration"]
	assert.Equal(t, 1, jobMetric.Gauge().DataPoints().Len())
	jobPt := jobMetric.Gauge().DataPoints().At(0)
	jobAttrs := jobPt.Attributes()
	jobID, _ := jobAttrs.Get("job_id")
	assert.EqualValues(t, TestdataJobID, jobID.Int())
	assert.EqualValues(t, 15000, jobPt.IntValue())

	taskMetric := metricMap["databricks.tasks.run.duration"]
	assert.Equal(t, 1, taskMetric.Gauge().DataPoints().Len())
	taskPt := taskMetric.Gauge().DataPoints().At(0)
	taskAttrs := taskPt.Attributes()
	jobID, _ = taskAttrs.Get("job_id")
	assert.EqualValues(t, TestdataJobID, jobID.Int())
	taskKey, _ := taskAttrs.Get("task_id")
	assert.Equal(t, "user-task", taskKey.Str())
	assert.EqualValues(t, 15000, taskPt.IntValue())
}

func TestFakeDatabricksRestService(t *testing.T) {
	p := &fakeDatabricksRestService{}
	runs, _ := p.CompletedJobRuns(42, 0)
	assert.Equal(t, 1, len(runs))
	runs, _ = p.CompletedJobRuns(42, 0)
	assert.Equal(t, 2, len(runs))
	assert.True(t, runs[0].StartTime > runs[1].StartTime)
	assert.True(t, runs[0].ExecutionDuration > runs[1].ExecutionDuration)
}
