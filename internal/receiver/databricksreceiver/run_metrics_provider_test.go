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
	"go.opentelemetry.io/collector/model/pdata"
)

func TestRunMetricProvider(t *testing.T) {
	p := newRunMetricsProvider(&fakeCompletedJobRunPaginator{})
	jobPts := pdata.NewNumberDataPointSlice()
	taskPts := pdata.NewNumberDataPointSlice()
	err := p.addSingleJobRunMetrics(jobPts, taskPts, 42)
	require.NoError(t, err)
	jobPt := jobPts.At(0)
	v, _ := jobPt.Attributes().Get("job_id")
	assert.EqualValues(t, 42, v.IntVal())
	assert.EqualValues(t, 15000, jobPt.IntVal())
}

func TestRunMetricsProvider_AddJobRunDurationMetrics(t *testing.T) {
	mp := newRunMetricsProvider(newPaginator(&testdataAPI{}))
	ms := pdata.NewMetricSlice()
	err := mp.addMultiJobRunMetrics(ms, []int{288})
	require.NoError(t, err)
	jobMetric := ms.At(0)
	jobPt := jobMetric.Gauge().DataPoints().At(0)
	jobAttrs := jobPt.Attributes()
	jobID, _ := jobAttrs.Get("job_id")
	assert.EqualValues(t, 288, jobID.IntVal())
	assert.EqualValues(t, 15000, jobPt.IntVal())

	taskMetric := ms.At(1)
	taskPt := taskMetric.Gauge().DataPoints().At(0)
	taskAttrs := taskPt.Attributes()
	jobID, _ = taskAttrs.Get("job_id")
	assert.EqualValues(t, 288, jobID.IntVal())
	taskKey, _ := taskAttrs.Get("task_key")
	assert.Equal(t, "user-task", taskKey.StringVal())
	assert.EqualValues(t, 15000, taskPt.IntVal())
}

func TestFakePaginator(t *testing.T) {
	p := &fakeCompletedJobRunPaginator{}
	runs, _ := p.completedJobRuns(42, 0)
	assert.Equal(t, 1, len(runs))
	runs, _ = p.completedJobRuns(42, 0)
	assert.Equal(t, 2, len(runs))
	assert.True(t, runs[0].StartTime > runs[1].StartTime)
	assert.True(t, runs[0].ExecutionDuration > runs[1].ExecutionDuration)
}

type fakeCompletedJobRunPaginator struct {
	runs []jobRun
	i    int
}

func (a *fakeCompletedJobRunPaginator) jobsList() (out []job, err error) {
	return nil, nil
}

func (a *fakeCompletedJobRunPaginator) activeJobRuns() (out []jobRun, err error) {
	return nil, nil
}

func (a *fakeCompletedJobRunPaginator) completedJobRuns(jobID int, _ int64) ([]jobRun, error) {
	a.addCompletedRun(jobID)
	return a.runs, nil
}

func (a *fakeCompletedJobRunPaginator) addCompletedRun(jobID int) {
	a.runs = append([]jobRun{{
		JobID:             jobID,
		StartTime:         1_600_000_000_000 + (1_000_000 * int64(a.i)),
		ExecutionDuration: 15_000 + (1000 * a.i),
	}}, a.runs...)
	a.i++
}
