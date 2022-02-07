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

	"go.uber.org/zap"
)

const (
	jobsListPath         = "/api/2.1/jobs/list?expand_tasks=true&limit=%d&offset=%d"
	activeJobRunsPath    = "/api/2.1/jobs/runs/list?active_only=true&limit=%d&offset=%d"
	completedJobRunsPath = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=%d&limit=%d&offset=%d"
)

// apiClientInterface is extracted from apiClient so that it can be swapped for
// testing.
type apiClientInterface interface {
	jobsList(limit int, offset int) ([]byte, error)
	activeJobRuns(limit int, offset int) ([]byte, error)
	completedJobRuns(id int, limit int, offset int) ([]byte, error)
}

// apiClient wraps an authClient, encapsulates calls to the databricks API, and
// implements apiClientInterface. Its methods return byte arrays to be unmarshalled
// by the caller.
type apiClient struct {
	logger     *zap.Logger
	authClient authClient
}

func newAPIClient(endpoint string, tok string, httpClient *http.Client, logger *zap.Logger) apiClientInterface {
	return &apiClient{
		authClient: authClient{
			httpClient: httpClient,
			endpoint:   endpoint,
			tok:        tok,
		},
		logger: logger,
	}
}

func (c apiClient) jobsList(limit int, offset int) (out []byte, err error) {
	path := fmt.Sprintf(jobsListPath, limit, offset)
	c.logger.Debug("apiClient.jobsList", zap.String("path", path))
	return c.authClient.get(path)
}

func (c apiClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(activeJobRunsPath, limit, offset)
	c.logger.Debug("apiClient.activeJobRuns", zap.String("path", path))
	return c.authClient.get(path)
}

func (c apiClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(completedJobRunsPath, jobID, limit, offset)
	c.logger.Debug("apiClient.completedJobRuns", zap.String("path", path))
	return c.authClient.get(path)
}
