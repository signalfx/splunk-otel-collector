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

func TestRunMetricProvider(t *testing.T) {
	p := newRunMetricsProvider(&fakeCompletedJobRunClient{})
	jobPts := pmetric.NewNumberDataPointSlice()
	err := p.addSingleJobRunMetrics(jobPts, pmetric.NewNumberDataPointSlice(), 42)
	require.NoError(t, err)
	assert.Equal(t, 0, jobPts.Len())

	jobPts = pmetric.NewNumberDataPointSlice()
	err = p.addSingleJobRunMetrics(jobPts, pmetric.NewNumberDataPointSlice(), 42)
	require.NoError(t, err)
	assert.Equal(t, 1, jobPts.Len())
	pt := jobPts.At(0)
	assert.EqualValues(t, 16000, pt.IntValue())
}

func TestRunMetricsProvider_AddJobRunDurationMetrics(t *testing.T) {
	const ignored = 25
	mp := newRunMetricsProvider(newDatabricksClient(&testdataClient{}, ignored))
	ms := pmetric.NewMetricSlice()
	err := mp.addMultiJobRunMetrics(ms, []int{288})
	require.NoError(t, err)
	jobMetric := ms.At(0)
	assert.Equal(t, 0, jobMetric.Gauge().DataPoints().Len())
	taskMetric := ms.At(1)
	assert.Equal(t, 0, taskMetric.Gauge().DataPoints().Len())

	ms = pmetric.NewMetricSlice()
	err = mp.addMultiJobRunMetrics(ms, []int{288})
	require.NoError(t, err)
	jobMetric = ms.At(0)
	assert.Equal(t, 1, jobMetric.Gauge().DataPoints().Len())
	taskMetric = ms.At(1)
	assert.Equal(t, 1, taskMetric.Gauge().DataPoints().Len())

	jobPt := jobMetric.Gauge().DataPoints().At(0)
	jobAttrs := jobPt.Attributes()
	jobID, _ := jobAttrs.Get("job_id")
	assert.EqualValues(t, 288, jobID.Int())
	assert.EqualValues(t, 15000, jobPt.IntValue())

	taskPt := taskMetric.Gauge().DataPoints().At(0)
	taskAttrs := taskPt.Attributes()
	jobID, _ = taskAttrs.Get("job_id")
	assert.EqualValues(t, 288, jobID.Int())
	taskKey, _ := taskAttrs.Get("task_id")
	assert.Equal(t, "user-task", taskKey.Str())
	assert.EqualValues(t, 15000, taskPt.IntValue())
}

func TestFakeCompletedJobRunClient(t *testing.T) {
	p := &fakeCompletedJobRunClient{}
	runs, _ := p.completedJobRuns(42, 0)
	assert.Equal(t, 1, len(runs))
	runs, _ = p.completedJobRuns(42, 0)
	assert.Equal(t, 2, len(runs))
	assert.True(t, runs[0].StartTime > runs[1].StartTime)
	assert.True(t, runs[0].ExecutionDuration > runs[1].ExecutionDuration)
}

type fakeCompletedJobRunClient struct {
	runs []jobRun
	i    int
}

func (c *fakeCompletedJobRunClient) jobs() (out []job, err error) {
	return nil, nil
}

func (c *fakeCompletedJobRunClient) activeJobRuns() (out []jobRun, err error) {
	return nil, nil
}

func (c *fakeCompletedJobRunClient) completedJobRuns(jobID int, _ int64) ([]jobRun, error) {
	c.addCompletedRun(jobID)
	return c.runs, nil
}

func (c *fakeCompletedJobRunClient) addCompletedRun(jobID int) {
	c.runs = append([]jobRun{{
		JobID:             jobID,
		StartTime:         1_600_000_000_000 + (1_000_000 * int64(c.i)),
		ExecutionDuration: 15_000 + (1000 * c.i),
	}}, c.runs...)
	c.i++
}
