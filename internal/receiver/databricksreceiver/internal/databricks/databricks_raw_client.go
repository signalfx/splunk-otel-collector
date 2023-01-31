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
	"net/http"

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

const (
	jobsListPath         = "/api/2.1/jobs/list?expand_tasks=true&limit=%d&offset=%d"
	activeJobRunsPath    = "/api/2.1/jobs/runs/list?active_only=true&limit=%d&offset=%d"
	completedJobRunsPath = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=%d&limit=%d&offset=%d"
	clustersListPath     = "/api/2.0/clusters/list"
	pipelinesPath        = "/api/2.0/pipelines"
	pipelinePath         = "/api/2.0/pipelines/%s"
)

// DatabricksRawClient defines methods extracted from databricksRawHTTPClient so
// that it can be swapped for testing.
type DatabricksRawClient interface {
	jobsList(limit int, offset int) ([]byte, error)
	activeJobRuns(limit int, offset int) ([]byte, error)
	completedJobRuns(id int, limit int, offset int) ([]byte, error)
	clustersList() ([]byte, error)
	pipelines() ([]byte, error)
	pipeline(string) ([]byte, error)
}

// databricksRawHTTPClient wraps an authClient, encapsulates calls to the databricks API, and
// implements DatabricksRawClient. Its methods return byte arrays to be unmarshalled
// by the caller.
type databricksRawHTTPClient struct {
	logger     *zap.Logger
	authClient httpauth.Client
	endpoint   string
}

func NewDatabricksRawClient(tok, endpoint string, httpClient *http.Client, logger *zap.Logger) DatabricksRawClient {
	return &databricksRawHTTPClient{
		authClient: httpauth.NewClient(httpClient, tok),
		endpoint:   endpoint,
		logger:     logger,
	}
}

func (c databricksRawHTTPClient) jobsList(limit int, offset int) (out []byte, err error) {
	path := fmt.Sprintf(jobsListPath, limit, offset)
	c.logger.Debug("databricksRawHTTPClient.jobsList", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c databricksRawHTTPClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(activeJobRunsPath, limit, offset)
	c.logger.Debug("databricksRawHTTPClient.activeJobRuns", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c databricksRawHTTPClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(completedJobRunsPath, jobID, limit, offset)
	c.logger.Debug("databricksRawHTTPClient.completedJobRuns", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c databricksRawHTTPClient) clustersList() ([]byte, error) {
	c.logger.Debug("databricksRawHTTPClient.clustersList", zap.String("path", clustersListPath))
	return c.authClient.Get(c.endpoint + clustersListPath)
}

func (c databricksRawHTTPClient) pipelines() ([]byte, error) {
	c.logger.Debug("databricksRawHTTPClient.pipelines", zap.String("path", pipelinesPath))
	return c.authClient.Get(c.endpoint + pipelinesPath)
}

func (c databricksRawHTTPClient) pipeline(s string) ([]byte, error) {
	path := fmt.Sprintf(pipelinePath, s)
	c.logger.Debug("databricksRawHTTPClient.pipeline", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}
