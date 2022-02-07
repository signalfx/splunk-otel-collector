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

	"go.opentelemetry.io/collector/model/pdata"
)

const (
	jobScheduleStatusMetric  = "databricks.jobs.schedule.status"
	totalJobsMetric          = "databricks.jobs.total"
	totalActiveJobsMetric    = "databricks.jobs.active.total"
	taskScheduleStatusMetric = "databricks.tasks.schedule.status"
)

// metricsProvider wraps a databricksClientInterface and provides metrics for databricks
// endpoints.
type metricsProvider struct {
	dbClient databricksClientInterface
}

func newMetricsProvider(dbClient databricksClientInterface) metricsProvider {
	return metricsProvider{dbClient: dbClient}
}

func (p metricsProvider) addJobStatusMetrics(ms pdata.MetricSlice) ([]int, error) {
	jobs, err := p.dbClient.jobs()
	if err != nil {
		return nil, fmt.Errorf("metricsProvider.addJobStatusMetrics(): %w", err)
	}

	addJobCountMetric(ms, int64(len(jobs)))

	jobPts := mkGauge(ms, jobScheduleStatusMetric)
	taskPts := mkGauge(ms, taskScheduleStatusMetric)

	var jobIDs []int
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.JobID)
		jobPt := jobPts.AppendEmpty()
		pauseStatus := pauseStatusToInt(j.Settings.Schedule.PauseStatus)
		jobPt.SetIntVal(pauseStatus)
		jobIDAttr := pdata.NewAttributeValueInt(int64(j.JobID))
		jobPt.Attributes().Insert("job_id", jobIDAttr)
		for _, task := range j.Settings.Tasks {
			taskPt := taskPts.AppendEmpty()
			taskPt.SetIntVal(pauseStatus)
			taskAttrs := taskPt.Attributes()
			taskAttrs.Insert("job_id", jobIDAttr)
			taskAttrs.Insert("task_id", pdata.NewAttributeValueString(task.TaskKey))
		}
	}
	return jobIDs, nil
}

func (p metricsProvider) addNumActiveRunsMetric(ms pdata.MetricSlice) error {
	runs, err := p.dbClient.activeJobRuns()
	if err != nil {
		return fmt.Errorf("metricsProvider.addNumActiveJobsMetric(): %w", err)
	}
	addActiveRunsMetric(ms, int64(len(runs)))
	return nil
}

func pauseStatusToInt(ps string) int64 {
	switch ps {
	case "PAUSED":
		return 0
	case "UNPAUSED":
		return 1
	default:
		// jobs that aren't scheduled end up here
		return 2
	}
}

func addJobCountMetric(ms pdata.MetricSlice, val int64) {
	mkGauge(ms, totalJobsMetric).AppendEmpty().SetIntVal(val)
}

func addActiveRunsMetric(ms pdata.MetricSlice, val int64) {
	mkGauge(ms, totalActiveJobsMetric).AppendEmpty().SetIntVal(val)
}

func mkGauge(ms pdata.MetricSlice, name string) pdata.NumberDataPointSlice {
	m := ms.AppendEmpty()
	m.SetName(name)
	m.SetDataType(pdata.MetricDataTypeGauge)
	return m.Gauge().DataPoints()
}
