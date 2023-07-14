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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

func TestRawHTTPClient(t *testing.T) {
	h := &httpauth.FakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()
	c := rawHTTPClient{
		authClient: httpauth.NewClient(http.DefaultClient, "abc123", zap.NewNop()),
		endpoint:   svr.URL,
		logger:     zap.NewNop(),
	}
	_, _ = c.jobsList(2, 3)
	path := "/api/2.1/jobs/list?expand_tasks=true&limit=2&offset=3"
	assert.Equal(t, path, h.Reqs[0].RequestURI)
	_, _ = c.activeJobRuns(2, 3)
	path = "/api/2.1/jobs/runs/list?active_only=true&limit=2&offset=3"
	assert.Equal(t, path, h.Reqs[1].RequestURI)
	_, _ = c.completedJobRuns(42, 2, 3)
	path = "/api/2.1/jobs/runs/list?completed_only=true&expand_tasks=true&job_id=42&limit=2&offset=3"
	assert.Equal(t, path, h.Reqs[2].RequestURI)
}
