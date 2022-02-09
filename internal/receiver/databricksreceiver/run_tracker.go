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

// runTracker keeps track of start times by job ID so that it can return just
// the new runs when given a list of them.
type runTracker struct {
	startTimesByJobID map[int]int64
}

func newRunTracker() *runTracker {
	return &runTracker{
		startTimesByJobID: map[int]int64{},
	}
}

func (t *runTracker) extractNewRuns(runs []jobRun) []jobRun {
	if runs == nil {
		return nil
	}
	// runs are sorted with the latest at the top/left/zero position
	latestRun := runs[0]
	jobID := latestRun.JobID
	prev := t.startTimesByJobID[jobID]
	t.startTimesByJobID[jobID] = latestRun.StartTime
	if prev == 0 {
		// we assume that the latest run we receive at startup not current (it's older
		// than our start time), so we discard it
		return nil
	}
	return collectNewRuns(runs, prev)
}

func (t *runTracker) getPrevStartTime(jobID int) int64 {
	return t.startTimesByJobID[jobID]
}

func collectNewRuns(runs []jobRun, prev int64) (out []jobRun) {
	for i := len(runs) - 1; i >= 0; i-- {
		run := runs[i]
		if run.StartTime <= prev {
			continue
		}
		out = append(out, run)
	}
	return out
}
