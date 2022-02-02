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

func TestMetricsProvider(t *testing.T) {
	mp := newMetricsProvider(newPaginator(&testdataAPI{}))
	ms := pdata.NewMetricSlice()
	_, err := mp.addJobStatusMetrics(ms)
	require.NoError(t, err)
	assert.Equal(t, ms.Len(), 3)

	jobTotalMetrics := ms.At(0)
	assert.Equal(t, "databricks.jobs.total", jobTotalMetrics.Name())
	assert.EqualValues(t, 6, jobTotalMetrics.Gauge().DataPoints().At(0).IntVal())

	jobScheduleMetrics := ms.At(1)
	assert.Equal(t, "databricks.jobs.schedule.status", jobScheduleMetrics.Name())
	pts := jobScheduleMetrics.Gauge().DataPoints()
	assert.Equal(t, 6, pts.Len())
	assert.EqualValues(t, 0, pts.At(0).IntVal())

	taskStatusMetric := ms.At(2)
	assert.Equal(t, "databricks.task.schedule.status", taskStatusMetric.Name())
	assert.Equal(t, 8, taskStatusMetric.Gauge().DataPoints().Len())

	ms = pdata.NewMetricSlice()
	err = mp.addNumActiveRunsMetric(ms)
	require.NoError(t, err)
	activeRunsMetric := ms.At(0)
	assert.Equal(t, "databricks.jobs.active.total", activeRunsMetric.Name())
	assert.Equal(t, 1, activeRunsMetric.Gauge().DataPoints().Len())
}
