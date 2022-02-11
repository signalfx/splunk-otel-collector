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

// databricksClientInterface is extracted from databricksClient for swapping out in unit tests
type databricksClientInterface interface {
	jobs() (out []job, err error)
	activeJobRuns() (out []jobRun, err error)
	completedJobRuns(jobID int, time int64) (out []jobRun, err error)
}

// databricksClient handles pagination (responses specify hasMore=true/false) and
// combines the returned objects into one array.
type databricksClient struct {
	unmarshaller unmarshaller
	limit        int
}

func newDatabricksClient(api apiClientInterface, limit int) databricksClient {
	return databricksClient{
		unmarshaller: unmarshaller{api: api},
		limit:        limit,
	}
}

func (c databricksClient) jobs() (out []job, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := c.unmarshaller.jobsList(c.limit, c.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksClient.jobs(): %w", err)
		}
		out = append(out, resp.Jobs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (c databricksClient) activeJobRuns() (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := c.unmarshaller.activeJobRuns(c.limit, c.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksClient.activeJobRuns(): %w", err)
		}
		out = append(out, resp.Runs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (c databricksClient) completedJobRuns(jobID int, prevStartTime int64) (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := c.unmarshaller.completedJobRuns(jobID, c.limit, c.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksClient.completedJobRuns(): %w", err)
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
