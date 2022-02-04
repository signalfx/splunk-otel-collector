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
	"net/http"
)

const (
	jobsListPath         = "/api/2.1/jobs/list?expand_tasks=true&limit=%d&offset=%d"
	activeJobRunsPath    = "/api/2.1/jobs/runs/list?active_only=true&limit=%d&offset=%d"
	completedJobRunsPath = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=%d&limit=%d&offset=%d"
)

// databricksAPI is an interface representing the databricks web-based API. It's
// populated with the parts of the databricks API that we use.
type databricksAPI interface {
	jobsList(limit int, offset int) ([]byte, error)
	activeJobRuns(limit int, offset int) ([]byte, error)
	completedJobRuns(id int, limit int, offset int) ([]byte, error)
}

// apiClient wraps an authClient, encapsulates calls to the databricks API, and
// implements databricksAPI. Its methods return byte arrays to be unmarshalled
// by the caller.
type apiClient struct {
	authClient authClient
}

func newAPIClient(endpoint string, tok string, httpClient *http.Client) databricksAPI {
	return &apiClient{authClient{
		httpClient: httpClient,
		endpoint:   endpoint,
		tok:        tok,
	}}
}

func (c apiClient) jobsList(limit int, offset int) (out []byte, err error) {
	return c.authClient.get(fmt.Sprintf(jobsListPath, limit, offset))
}

func (c apiClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	return c.authClient.get(fmt.Sprintf(activeJobRunsPath, limit, offset))
}

func (c apiClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	return c.authClient.get(fmt.Sprintf(completedJobRunsPath, jobID, limit, offset))
}
