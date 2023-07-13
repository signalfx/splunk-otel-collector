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

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

// RawClient defines methods extracted from rawHTTPClient so
// that it can be swapped for testing.
type RawClient interface {
	jobsList(limit int, offset int) ([]byte, error)
	activeJobRuns(limit int, offset int) ([]byte, error)
	completedJobRuns(id int, limit int, offset int) ([]byte, error)
	clustersList() ([]byte, error)
	pipelines() ([]byte, error)
	pipeline(string) ([]byte, error)
}

// rawHTTPClient wraps an authClient, encapsulates calls to the databricks API, and
// implements RawClient. Its methods return byte arrays to be unmarshalled
// by the caller.
type rawHTTPClient struct {
	logger     *zap.Logger
	authClient httpauth.Client
	endpoint   string
}

func NewRawClient(tok, endpoint string, httpDoer httpauth.HTTPDoer, logger *zap.Logger) RawClient {
	return rawHTTPClient{
		authClient: httpauth.NewClient(httpDoer, tok, logger),
		endpoint:   endpoint,
		logger:     logger,
	}
}

func (c rawHTTPClient) jobsList(limit int, offset int) (out []byte, err error) {
	path := fmt.Sprintf("/api/2.1/jobs/list?expand_tasks=true&limit=%d&offset=%d", limit, offset)
	c.logger.Debug("rawHTTPClient.jobsList", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c rawHTTPClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf("/api/2.1/jobs/runs/list?active_only=true&limit=%d&offset=%d", limit, offset)
	c.logger.Debug("rawHTTPClient.activeJobRuns", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c rawHTTPClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	path := fmt.Sprintf("/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=%d&limit=%d&offset=%d", jobID, limit, offset)
	c.logger.Debug("rawHTTPClient.completedJobRuns", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}

func (c rawHTTPClient) clustersList() ([]byte, error) {
	const clustersListPath = "/api/2.0/clusters/list"
	c.logger.Debug("rawHTTPClient.clustersList", zap.String("path", clustersListPath))
	return c.authClient.Get(c.endpoint + clustersListPath)
}

func (c rawHTTPClient) pipelines() ([]byte, error) {
	const pipelinesPath = "/api/2.0/pipelines"
	c.logger.Debug("rawHTTPClient.pipelines", zap.String("path", pipelinesPath))
	return c.authClient.Get(c.endpoint + pipelinesPath)
}

func (c rawHTTPClient) pipeline(s string) ([]byte, error) {
	path := fmt.Sprintf("/api/2.0/pipelines/%s", s)
	c.logger.Debug("rawHTTPClient.pipeline", zap.String("path", path))
	return c.authClient.Get(c.endpoint + path)
}
