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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/zap"
)

func TestBearerClient(t *testing.T) {
	h := &FakeHandler{}
	svr := httptest.NewServer(h)
	defer svr.Close()
	s := confighttp.HTTPClientSettings{}
	httpClient, err := s.ToClient(nil, component.TelemetrySettings{})
	require.NoError(t, err)
	ac := NewClient(httpClient, "abc123", zap.NewNop())
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

func TestNoBackoff(t *testing.T) {
	backoff := &fakeBackoff{}
	c := bearerClient{
		httpDoer: &fakeHTTPDoer{},
		ebo:      backoff,
		logger:   zap.NewNop(),
	}
	_, err := c.Get("")
	require.NoError(t, err)
	assert.Equal(t, 0, backoff.backoffCallCount)
}

func TestBackoff(t *testing.T) {
	const numBackoffs = 1
	backoff := &fakeBackoff{}
	c := bearerClient{
		httpDoer: &fakeHTTPDoer{numFailedRequestsBeforeSuccess: numBackoffs},
		ebo:      backoff,
		logger:   zap.NewNop(),
	}
	_, err := c.Get("")
	require.NoError(t, err)
	assert.Equal(t, numBackoffs, backoff.backoffCallCount)
	assert.Equal(t, 1, backoff.resetCallCount)
}

func TestRateLimitingHTTPDoer(t *testing.T) {
	d := &RateLimitingHTTPDoer{
		doer:     &fakeHTTPDoer{},
		duration: time.Second,
	}
	resp, err := d.Do(nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = d.Do(nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

type fakeHTTPDoer struct {
	numFailedRequestsBeforeSuccess int
	currRequestNumber              int
}

func (d *fakeHTTPDoer) Do(*http.Request) (*http.Response, error) {
	out := &http.Response{
		Body: &nopCloser{},
	}
	if d.currRequestNumber < d.numFailedRequestsBeforeSuccess {
		out.StatusCode = http.StatusTooManyRequests
	} else {
		out.StatusCode = http.StatusOK
	}
	d.currRequestNumber++
	return out, nil
}

type fakeBackoff struct {
	backoffCallCount int
	resetCallCount   int
}

func (b *fakeBackoff) NextBackOff() time.Duration {
	b.backoffCallCount++
	return 0
}

func (b *fakeBackoff) Reset() {
	b.resetCallCount++
}

type nopCloser struct {
	bytes.Buffer
}

func (nopCloser) Close() error {
	return nil
}

// RateLimitingHTTPDoer lets you wrap an HTTPDoer (aka http.Client) to simulate
// 429s for manual integration testing (because causing a web server to actually
// send 429s can be tricky).
type RateLimitingHTTPDoer struct {
	doer            HTTPDoer
	lastSuccessTime time.Time
	duration        time.Duration
}

func (rl *RateLimitingHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	now := time.Now()
	if now.Sub(rl.lastSuccessTime) > rl.duration {
		rl.lastSuccessTime = now
		return rl.doer.Do(req)
	}
	return &http.Response{
		StatusCode: http.StatusTooManyRequests,
	}, nil
}
