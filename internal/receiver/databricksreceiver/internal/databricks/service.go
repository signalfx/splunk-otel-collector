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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

// Service is extracted from restService for swapping out in unit tests
type Service interface {
	jobs() ([]Job, error)
	activeJobRuns() ([]JobRun, error)
	CompletedJobRuns(int, int64) ([]JobRun, error)
	RunningClusters() ([]spark.Cluster, error)
	RunningPipelines() ([]spark.PipelineSummary, error)
}

// restService handles pagination (responses specify hasMore=true/false) and
// combines the returned objects into one array.
type restService struct {
	client client
	limit  int
}

func NewService(dbrc RawClient, limit int) Service {
	return restService{
		client: client{rawClient: dbrc},
		limit:  limit,
	}
}

func (s restService) CompletedJobRuns(jobID int, prevStartTime int64) (out []JobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.client.completedJobRuns(jobID, s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("CompletedJobRuns failed to get completedJobRuns: %w", err)
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

func (s restService) RunningClusters() ([]spark.Cluster, error) {
	cl, err := s.client.clustersList()
	if err != nil {
		return nil, fmt.Errorf("RunningClusters failed to get runningClusterIDs: %w", err)
	}
	var out []spark.Cluster
	for _, c := range cl.Clusters {
		if c.State != "RUNNING" {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (s restService) RunningPipelines() ([]spark.PipelineSummary, error) {
	pipelines, err := s.client.pipelines()
	if err != nil {
		return nil, fmt.Errorf("RunningPipelines failed to get pipelines: %w", err)
	}
	var out []spark.PipelineSummary
	for _, status := range pipelines.Statuses {
		if status.State != "RUNNING" {
			continue
		}
		pipeline, err := s.client.pipeline(status.PipelineID)
		if err != nil {
			return nil, fmt.Errorf(
				"RunningPipelines failed to get single pipeline: id: %s: %w",
				status.PipelineID,
				err,
			)
		}
		out = append(out, spark.PipelineSummary{
			ID:        status.PipelineID,
			Name:      status.Name,
			ClusterID: pipeline.ClusterID,
		})
	}
	return out, nil
}

func (s restService) jobs() (out []Job, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.client.jobsList(s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("jobs failed to get job list: %w", err)
		}
		out = append(out, resp.Jobs...)
		hasMore = resp.HasMore
	}
	return out, nil
}

func (s restService) activeJobRuns() (out []JobRun, err error) {
	hasMore := true
	for i := 0; hasMore; i++ {
		resp, err := s.client.activeJobRuns(s.limit, s.limit*i)
		if err != nil {
			return nil, fmt.Errorf("activeJobRuns failed to get activeJobRuns: %w", err)
		}
		out = append(out, resp.Runs...)
		hasMore = resp.HasMore
	}
	return out, nil
}
