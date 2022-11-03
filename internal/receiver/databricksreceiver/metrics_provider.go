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

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

// metricsProvider wraps a databricksClientInterface and provides metrics for databricks
// endpoints.
type metricsProvider struct {
	dbClient databricksClientInterface
}

func newMetricsProvider(dbClient databricksClientInterface) metricsProvider {
	return metricsProvider{dbClient: dbClient}
}

func (p metricsProvider) addJobStatusMetrics(ms pmetric.MetricSlice) ([]int, error) {
	jobs, err := p.dbClient.jobs()
	if err != nil {
		return nil, fmt.Errorf("metricsProvider.addJobStatusMetrics(): %w", err)
	}

	initGauge(ms, metadata.M.DatabricksJobsTotal).AppendEmpty().SetIntValue(int64(len(jobs)))

	jobPts := initGauge(ms, metadata.M.DatabricksJobsScheduleStatus)
	taskPts := initGauge(ms, metadata.M.DatabricksTasksScheduleStatus)

	var jobIDs []int
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.JobID)
		jobPt := jobPts.AppendEmpty()
		pauseStatus := pauseStatusToInt(j.Settings.Schedule.PauseStatus)
		jobPt.SetIntValue(pauseStatus)
		jobPt.Attributes().PutInt(metadata.A.JobID, int64(j.JobID))
		for _, task := range j.Settings.Tasks {
			taskPt := taskPts.AppendEmpty()
			taskPt.SetIntValue(pauseStatus)
			taskAttrs := taskPt.Attributes()
			taskAttrs.PutInt(metadata.A.JobID, int64(j.JobID))
			taskAttrs.PutStr(metadata.A.TaskID, task.TaskKey)
			taskAttrs.PutStr(metadata.A.TaskType, taskType(task))
		}
	}
	return jobIDs, nil
}

func taskType(task jobTask) string {
	switch {
	case task.NotebookTask != nil:
		return metadata.AttributeTaskType.NotebookTask
	case task.SparkJarTask != nil:
		return metadata.AttributeTaskType.SparkJarTask
	case task.SparkPythonTask != nil:
		return metadata.AttributeTaskType.SparkPythonTask
	case task.PipelineTask != nil:
		return metadata.AttributeTaskType.PipelineTask
	case task.PythonWheelTask != nil:
		return metadata.AttributeTaskType.PythonWheelTask
	case task.SparkSubmitTask != nil:
		return metadata.AttributeTaskType.SparkSubmitTask
	}
	return ""
}

func (p metricsProvider) addNumActiveRunsMetric(ms pmetric.MetricSlice) error {
	runs, err := p.dbClient.activeJobRuns()
	if err != nil {
		return fmt.Errorf("metricsProvider.addNumActiveJobsMetric(): %w", err)
	}
	pts := initGauge(ms, metadata.M.DatabricksJobsActiveTotal)
	pts.AppendEmpty().SetIntValue(int64(len(runs)))
	return nil
}

func initGauge(ms pmetric.MetricSlice, mi metadata.MetricIntf) pmetric.NumberDataPointSlice {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m.Gauge().DataPoints()
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
