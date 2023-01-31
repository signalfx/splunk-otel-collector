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

const TestdataJobID = 288

// databricksClient wraps a DatabricksRawClient implementation and unmarshals json byte
// arrays to the types defined in json_types.go. Its method signatures mirror
// those of the rawClient.
type databricksClient struct {
	rawClient DatabricksRawClient
}

func (c databricksClient) jobsList(limit int, offset int) (jobsList, error) {
	bytes, err := c.rawClient.jobsList(limit, offset)
	out := jobsList{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.jobsList(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c databricksClient) activeJobRuns(limit int, offset int) (jobRuns, error) {
	bytes, err := c.rawClient.activeJobRuns(limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.activeJobRuns(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c databricksClient) completedJobRuns(jobID int, limit int, offset int) (jobRuns, error) {
	bytes, err := c.rawClient.completedJobRuns(jobID, limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.completedJobRuns(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c databricksClient) clustersList() (clusterList, error) {
	bytes, err := c.rawClient.clustersList()
	out := clusterList{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.clusterList(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c databricksClient) pipelines() (pipelinesInfo, error) {
	bytes, err := c.rawClient.pipelines()
	out := pipelinesInfo{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.pipelines(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (c databricksClient) pipeline(id string) (pipelineInfo, error) {
	bytes, err := c.rawClient.pipeline(id)
	out := pipelineInfo{}
	if err != nil {
		return out, fmt.Errorf("databricksClient.pipeline(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}
