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

package httpauth

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestBearerClient(t *testing.T) {
	h := &FakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()
	s := confighttp.HTTPClientSettings{}
	httpClient, err := s.ToClient(nil, component.TelemetrySettings{})
	require.NoError(t, err)
	ac := NewClient(httpClient, "abc123")
	_, _ = ac.Get(svr.URL + "/foo")
	req := h.Reqs[0]
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "Bearer abc123", req.Header.Get("Authorization"))
	assert.Equal(t, "/foo", req.RequestURI)
}

func TestIsForbidden(t *testing.T) {
	wrapper := fmt.Errorf("wrapping this error: %w", ErrForbidden)
	assert.True(t, IsForbidden(wrapper))
}
