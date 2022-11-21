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

// databricksServiceIntf is extracted from databricksService for swapping out in unit tests
type databricksServiceIntf interface {
	jobs() ([]job, error)
	activeJobRuns() ([]jobRun, error)
	completedJobRuns(int, int64) ([]jobRun, error)
	runningClusters() ([]cluster, error)
}

// databricksService handles pagination (responses specify hasMore=true/false) and
// combines the returned objects into one array.
type databricksService struct {
	unmarshaller databricksUnmarshaller
	limit        int
}

func newDatabricksService(dbc databricksClientIntf, limit int) databricksService {
	return databricksService{
		unmarshaller: databricksUnmarshaller{dbc: dbc},
		limit:        limit,
	}
}

func (s databricksService) jobs() (out []job, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.unmarshaller.jobsList(s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksService.jobs(): %w", err)
		}
		out = append(out, resp.Jobs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (s databricksService) activeJobRuns() (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.unmarshaller.activeJobRuns(s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksService.activeJobRuns(): %w", err)
		}
		out = append(out, resp.Runs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (s databricksService) completedJobRuns(jobID int, prevStartTime int64) (out []jobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.unmarshaller.completedJobRuns(jobID, s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("databricksService.completedJobRuns(): %w", err)
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

func (s databricksService) runningClusters() ([]cluster, error) {
	cl, err := s.unmarshaller.clusterList()
	if err != nil {
		return nil, fmt.Errorf("databricksService.runningClusterIDs(): %w", err)
	}
	var out []cluster
	for _, c := range cl.Clusters {
		if c.State != "RUNNING" {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}
