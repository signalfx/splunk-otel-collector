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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestAuthClient(t *testing.T) {
	h := &fakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()

	s := confighttp.HTTPClientSettings{}
	httpClient, err := s.ToClient(nil, component.TelemetrySettings{})
	require.NoError(t, err)
	ac := authClient{
		httpClient: httpClient,
		endpoint:   svr.URL,
		tok:        "abc123",
	}
	_, _ = ac.get("/foo")
	req := h.reqs[0]
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "Bearer abc123", req.Header.Get("Authorization"))
	assert.Equal(t, "/foo", req.RequestURI)
}

// fakeHandler handles test reqests to a test server, appending requests to an
// array member for later inspection
type fakeHandler struct {
	reqs []*http.Request
}

func (h *fakeHandler) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	h.reqs = append(h.reqs, req)
}
