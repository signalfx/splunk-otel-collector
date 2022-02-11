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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAPIClient(t *testing.T) {
	h := &fakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()
	c := apiClient{
		authClient: authClient{
			httpClient: http.DefaultClient,
			endpoint:   svr.URL,
			tok:        "abc123",
		},
		logger: zap.NewNop(),
	}
	_, _ = c.jobsList(2, 3)
	path := "/api/2.1/jobs/list?expand_tasks=true&limit=2&offset=3"
	assert.Equal(t, path, h.reqs[0].RequestURI)
	_, _ = c.activeJobRuns(2, 3)
	path = "/api/2.1/jobs/runs/list?active_only=true&limit=2&offset=3"
	assert.Equal(t, path, h.reqs[1].RequestURI)
	_, _ = c.completedJobRuns(42, 2, 3)
	path = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=42&limit=2&offset=3"
	assert.Equal(t, path, h.reqs[2].RequestURI)
}

// testdataClient implements apiClientInterface but is backed by json files in testdata.
type testdataClient struct {
	i int
}

func (*testdataClient) jobsList(limit int, offset int) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("testdata/jobs-list-%d.json", offset/limit))
}

func (*testdataClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("testdata/active-job-runs-%d.json", offset/limit))
}

func (c *testdataClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	if jobID != 288 {
		return []byte("{}"), nil
	}
	if offset == 0 {
		c.i++
	}
	file, err := os.ReadFile(fmt.Sprintf("testdata/completed-job-runs-%d-%d.json", c.i-1, offset/limit))
	return file, err
}
