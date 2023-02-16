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
	"encoding/json"
	"fmt"
)

// client wraps a RawClient implementation and unmarshals json byte
// arrays to the types defined in json_types.go. Its method signatures mirror
// those of the rawClient.
type client struct {
	rawClient RawClient
}

func (c client) jobsList(limit int, offset int) (jobsList, error) {
	bytes, err := c.rawClient.jobsList(limit, offset)
	out := jobsList{}
	if err != nil {
		return out, fmt.Errorf("jobsList failed to get jobs list: %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c client) activeJobRuns(limit int, offset int) (jobRuns, error) {
	bytes, err := c.rawClient.activeJobRuns(limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("activeJobRuns failed to get active job runs: %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c client) completedJobRuns(jobID int, limit int, offset int) (jobRuns, error) {
	bytes, err := c.rawClient.completedJobRuns(jobID, limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("completedJobRuns failed to get completed job runs: %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c client) clustersList() (clusterList, error) {
	bytes, err := c.rawClient.clustersList()
	out := clusterList{}
	if err != nil {
		return out, fmt.Errorf("clustersList failed to get cluster list: %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c client) pipelines() (pipelinesInfo, error) {
	bytes, err := c.rawClient.pipelines()
	out := pipelinesInfo{}
	if err != nil {
		return out, fmt.Errorf("pipelines failed to get pipelines: %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c client) pipeline(id string) (pipelineInfo, error) {
	bytes, err := c.rawClient.pipeline(id)
	out := pipelineInfo{}
	if err != nil {
		return out, fmt.Errorf("pipeline failed to get pipeline: id %s: %w", id, err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}
