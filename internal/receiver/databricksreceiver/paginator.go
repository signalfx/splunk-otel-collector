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

import "fmt"

// paginator is extracted from apiPaginator for swapping out in unit tests
type paginator interface {
	jobs() (out []job, err error)
	activeJobRuns() (out []jobRun, err error)
	completedJobRuns(jobID int, time int64) (out []jobRun, err error)
}

// apiPaginator handles pagination (responses specify hasMore=true/false) and
// combines the returned objects into one array.
type apiPaginator struct {
	unmarshaller unmarshaller
	limit        int
}

func newPaginator(api databricksAPI, limit int) apiPaginator {
	return apiPaginator{
		unmarshaller: unmarshaller{api: api},
		limit:        limit,
	}
}

func (p apiPaginator) jobs() (out []job, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := p.unmarshaller.jobsList(p.limit, p.limit*i)
		if err != nil {
			return nil, fmt.Errorf("apiPaginator.jobs(): %w", err)
		}
		out = append(out, resp.Jobs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (p apiPaginator) activeJobRuns() (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := p.unmarshaller.activeJobRuns(p.limit, p.limit*i)
		if err != nil {
			return nil, fmt.Errorf("apiPaginator.activeJobRuns(): %w", err)
		}
		out = append(out, resp.Runs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (p apiPaginator) completedJobRuns(jobID int, prevStartTime int64) (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := p.unmarshaller.completedJobRuns(jobID, p.limit, p.limit*i)
		if err != nil {
			return nil, fmt.Errorf("apiPaginator.completedJobRuns(): %w", err)
		}
		out = append(out, resp.Runs...)
		if prevStartTime == 0 || resp.Runs == nil || resp.Runs[len(resp.Runs)-1].StartTime < prevStartTime {
			// Don't do another api request if this is the first time through (time == 0) or
			// if the bottom/earliest run in the response is older than our previous startTime
			// for this job id.
			break
		}
		hasMore = resp.HasMore
	}
	return out, nil
}
