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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

const (
	jobsListPath         = "/api/2.1/jobs/list?expand_tasks=true&limit=%d&offset=%d"
	activeJobRunsPath    = "/api/2.1/jobs/runs/list?active_only=true&limit=%d&offset=%d"
	completedJobRunsPath = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=%d&limit=%d&offset=%d"
	clustersListPath     = "/api/2.0/clusters/list"
)

// databricksClientIntf is extracted from databricksClient so that it can be swapped for
// testing.
type databricksClientIntf interface {
	jobsList(limit int, offset int) ([]byte, error)
	activeJobRuns(limit int, offset int) ([]byte, error)
	completedJobRuns(id int, limit int, offset int) ([]byte, error)
	clustersList() ([]byte, error)
}

// databricksClient wraps an authClient, encapsulates calls to the databricks API, and
// implements databricksClientIntf. Its methods return byte arrays to be unmarshalled
// by the caller.
type databricksClient struct {
	logger     *zap.Logger
	authClient httpauth.ClientIntf
}

func newDatabricksClient(endpoint string, tok string, httpClient *http.Client, logger *zap.Logger) databricksClientIntf {
	return &databricksClient{
		authClient: httpauth.NewClient(httpClient, endpoint, tok),
		logger:     logger,
	}
}

func (c databricksClient) jobsList(limit int, offset int) (out []byte, err error) {
	path := fmt.Sprintf(jobsListPath, limit, offset)
	c.logger.Debug("databricksClient.jobsList", zap.String("path", path))
	return c.authClient.Get(path)
}

func (c databricksClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(activeJobRunsPath, limit, offset)
	c.logger.Debug("databricksClient.activeJobRuns", zap.String("path", path))
	return c.authClient.Get(path)
}

func (c databricksClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf(completedJobRunsPath, jobID, limit, offset)
	c.logger.Debug("databricksClient.completedJobRuns", zap.String("path", path))
	return c.authClient.Get(path)
}

func (c databricksClient) clustersList() ([]byte, error) {
	c.logger.Debug("databricksClient.clustersList", zap.String("path", clustersListPath))
	return c.authClient.Get(clustersListPath)
}
