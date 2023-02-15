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

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

// runMetricsProvider provides metrics for job and task runs. It uses a
// runTracker to extract just the new runs returned from the API.
type runMetricsProvider struct {
	tracker *runTracker
	dbrsvc  databricksService
}

func newRunMetricsProvider(dbrsvc databricksService) runMetricsProvider {
	return runMetricsProvider{
		tracker: newRunTracker(),
		dbrsvc:  dbrsvc,
	}
}

func (p runMetricsProvider) addMultiJobRunMetrics(jobIDs []int, builder *metadata.MetricsBuilder, ts pcommon.Timestamp) error {
	for _, jobID := range jobIDs {
		err := p.addSingleJobRunMetrics(jobID, builder, ts)
		if err != nil {
			return fmt.Errorf("runMetricsProvider.addMultiJobRunMetrics(): aborting: %w", err)
		}
	}
	return nil
}

func (p runMetricsProvider) addSingleJobRunMetrics(jobID int, builder *metadata.MetricsBuilder, ts pcommon.Timestamp) error {
	startTime := p.tracker.getPrevStartTime(jobID)
	runs, err := p.dbrsvc.completedJobRuns(jobID, startTime)
	if err != nil {
		return fmt.Errorf("runMetricsProvider.addSingleJobRunMetrics(): %w", err)
	}
	newRuns := p.tracker.extractNewRuns(runs)
	for _, run := range newRuns {
		// consider skipping run.State.LifeCycleState == "TERMINATED" due to error
		if run.State.LifeCycleState == "SKIPPED" {
			continue
		}
		builder.RecordDatabricksJobsRunDurationDataPoint(ts, int64(run.ExecutionDuration), int64(jobID))
		for _, task := range run.Tasks {
			builder.RecordDatabricksTasksRunDurationDataPoint(ts, int64(task.ExecutionDuration), int64(jobID), task.TaskKey)
		}
	}
	return nil
}
