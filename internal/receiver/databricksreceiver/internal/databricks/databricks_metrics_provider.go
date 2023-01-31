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
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

// DbrMetricsProvider wraps a Service and provides metrics for databricks
// endpoints.
type DbrMetricsProvider struct {
	Dbrsvc Service
}

func (p DbrMetricsProvider) AddJobStatusMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp) ([]int, error) {
	jobs, err := p.Dbrsvc.jobs()
	if err != nil {
		return nil, fmt.Errorf("dbrMetricsProvider.addJobStatusMetrics(): %w", err)
	}
	builder.RecordDatabricksJobsTotalDataPoint(now, int64(len(jobs)))

	var jobIDs []int
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.JobID)
		pauseStatus := pauseStatusToInt(j.Settings.Schedule.PauseStatus)
		builder.RecordDatabricksJobsScheduleStatusDataPoint(now, pauseStatus, int64(j.JobID))
		for _, task := range j.Settings.Tasks {
			builder.RecordDatabricksTasksScheduleStatusDataPoint(
				now,
				pauseStatus,
				int64(j.JobID),
				task.TaskKey,
				taskType(task),
			)
		}
	}
	return jobIDs, nil
}

func (p DbrMetricsProvider) AddNumActiveRunsMetric(builder *metadata.MetricsBuilder, now pcommon.Timestamp) error {
	runs, err := p.Dbrsvc.activeJobRuns()
	if err != nil {
		return fmt.Errorf("dbrMetricsProvider.addNumActiveJobsMetric(): %w", err)
	}
	builder.RecordDatabricksJobsActiveTotalDataPoint(now, int64(len(runs)))
	return nil
}

func taskType(task jobTask) metadata.AttributeTaskType {
	switch {
	case task.NotebookTask != nil:
		return metadata.AttributeTaskTypeNotebookTask
	case task.SparkJarTask != nil:
		return metadata.AttributeTaskTypeSparkJarTask
	case task.SparkPythonTask != nil:
		return metadata.AttributeTaskTypeSparkPythonTask
	case task.PipelineTask != nil:
		return metadata.AttributeTaskTypePipelineTask
	case task.PythonWheelTask != nil:
		return metadata.AttributeTaskTypePythonWheelTask
	case task.SparkSubmitTask != nil:
		return metadata.AttributeTaskTypeSparkSubmitTask
	}
	return 0
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
