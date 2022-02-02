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
	"encoding/json"
	"fmt"
)

// unmarshaller wraps a databricksAPI implementation and unmarshals json byte
// arrays to the types defined in json_types.go. Its methods signatures mirror
// those of the api.
type unmarshaller struct {
	api databricksAPI
}

func (u unmarshaller) jobsList(limit int, offset int) (jobsList, error) {
	bytes, err := u.api.jobsList(limit, offset)
	out := jobsList{}
	if err != nil {
		return out, fmt.Errorf("unmarshaller.jobsList(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (u unmarshaller) activeJobRuns(limit int, offset int) (jobRuns, error) {
	bytes, err := u.api.activeJobRuns(limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("unmarshaller.activeJobRuns(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}

func (u unmarshaller) completedJobRuns(jobID int, limit int, offset int) (jobRuns, error) {
	bytes, err := u.api.completedJobRuns(jobID, limit, offset)
	out := jobRuns{}
	if err != nil {
		return out, fmt.Errorf("unmarshaller.completedJobRuns(): %w", err)
	}
	err = json.Unmarshal(bytes, &out)
	return out, err
}
