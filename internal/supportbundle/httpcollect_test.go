// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package supportbundle

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func withHTTPTransport(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	original := httpClient
	httpClient = &http.Client{Transport: transport}
	t.Cleanup(func() { httpClient = original })
}

func TestFetchToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("otelcol_up 1\n"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "metrics.txt")
	require.NoError(t, fetchToFile(t.Context(), srv.URL, dest))

	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "otelcol_up 1\n", string(got))
}

func TestFetchToFile_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "metrics.txt")
	err := fetchToFile(t.Context(), srv.URL, dest)
	require.Error(t, err)
	assert.NoFileExists(t, dest)
}

func TestCollectMetrics_Unreachable(t *testing.T) {
	withHTTPTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, metricsURL, r.URL.String())
		return nil, errors.New("test endpoint unavailable")
	}))
	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, metricsDirName), 0o755))

	logger, buf := newTestLogger()
	collectMetrics(t.Context(), logger, tempDir)

	assert.Contains(t, buf.String(), "WARN")
	assert.Contains(t, buf.String(), "test endpoint unavailable")
	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "collector-metrics.txt"))
}
