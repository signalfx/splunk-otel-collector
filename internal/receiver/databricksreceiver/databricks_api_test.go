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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIClient(t *testing.T) {
	h := &fakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()
	c := apiClient{authClient{
		baseURL: svr.URL,
		tok:     "abc123",
	}}
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

// testdataAPI implements databricksAPI but is backed by json files in testdata.
type testdataAPI struct {
}

func (*testdataAPI) jobsList(limit int, offset int) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("testdata/jobs-list-%d.json", offset/limit))
}

func (*testdataAPI) activeJobRuns(limit int, offset int) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("testdata/active-job-runs-%d.json", offset/limit))
}

func (*testdataAPI) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("testdata/completed-job-runs-%d.json", offset/limit))
}
