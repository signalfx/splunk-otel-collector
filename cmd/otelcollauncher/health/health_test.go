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

package health

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const verboseCollectorResponse = `{
	"start_time": "2024-01-18T17:27:12.570394-08:00",
	"healthy": true,
	"status": "StatusRecoverableError",
	"error": "rpc error: code = ResourceExhausted desc = resource exhausted",
	"status_time": "2024-01-18T17:27:32.572301-08:00",
	"components": {
		"extensions": {
			"healthy": true,
			"status": "StatusOK",
			"status_time": "2024-01-18T17:27:12.570428-08:00",
			"components": {
				"extension:healthcheckv2": {
					"healthy": true,
					"status": "StatusOK",
					"status_time": "2024-01-18T17:27:12.570428-08:00"
				}
			}
		},
		"pipeline:metrics/grpc": {
			"healthy": true,
			"status": "StatusRecoverableError",
			"error": "rpc error: code = ResourceExhausted desc = resource exhausted",
			"status_time": "2024-01-18T17:27:32.572301-08:00",
			"components": {
				"exporter:otlp_grpc/staging": {
					"healthy": true,
					"status": "StatusRecoverableError",
					"error": "rpc error: code = ResourceExhausted desc = resource exhausted",
					"status_time": "2024-01-18T17:27:32.572301-08:00"
				},
				"receiver:otlp": {
					"healthy": true,
					"status": "StatusOK",
					"status_time": "2024-01-18T17:27:12.571576-08:00"
				}
			}
		}
	}
}`

// newFakeExtension starts a stand-in for the healthcheckv2 extension's HTTP
// status server and returns its host/port. lastRequest, if non-nil, captures
// the most recent request's URL for assertions.
func newFakeExtension(t *testing.T, body string, statusCode int, lastRequest *url.URL) (host string, port int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lastRequest != nil {
			*lastRequest = *r.URL
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	h, p, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)
	portNum, err := strconv.Atoi(p)
	require.NoError(t, err)
	return h, portNum
}

func TestFamilyVerbs(t *testing.T) {
	verbs := New().Verbs()
	require.Len(t, verbs, 1)
	assert.Equal(t, "health", verbs[0].Name)
	assert.NotEmpty(t, verbs[0].Help)
	assert.Contains(t, verbs[0].Usage, "--component")
}

func TestDispatchUnknownVerb(t *testing.T) {
	code := New().Dispatch("bogus", nil, cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})
	assert.Equal(t, cli.ExitUsage, code)
}

func TestHealthDefaultQueryHitsDefaultURLShape(t *testing.T) {
	var gotURL url.URL
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, &gotURL)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	assert.Equal(t, cli.ExitOK, code)
	assert.Equal(t, defaultPath, gotURL.Path)
	assert.True(t, gotURL.Query().Has("verbose"))
	assert.Empty(t, gotURL.Query().Get("pipeline"))
	assert.Contains(t, out.String(), "Overall status: RecoverableError")
	assert.Contains(t, out.String(), "healthy: yes")
}

func TestHealthCustomPathAndPipeline(t *testing.T) {
	var gotURL url.URL
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, &gotURL)

	code := New().Dispatch("health", []string{
		"--host", host,
		"--port", strconv.Itoa(port),
		"--path", "/custom/status",
		"--pipeline", "traces/http",
	}, cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})

	assert.Equal(t, cli.ExitOK, code)
	assert.Equal(t, "/custom/status", gotURL.Path)
	assert.Equal(t, "traces/http", gotURL.Query().Get("pipeline"))
}

func TestHealthComponentFilterNarrowsOutput(t *testing.T) {
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, nil)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{
		"--host", host, "--port", strconv.Itoa(port), "--component", "otlp_grpc",
	}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	assert.Equal(t, cli.ExitOK, code)
	text := out.String()
	assert.Contains(t, text, "exporter:otlp_grpc/staging")
	assert.Contains(t, text, "pipeline:metrics/grpc", "ancestors of a match must still print")
	assert.NotContains(t, text, "receiver:otlp", "siblings that don't match must be pruned")
	assert.NotContains(t, text, "extension:healthcheckv2", "unrelated branches must be pruned")
}

func TestHealthComponentFilterNoMatch(t *testing.T) {
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, nil)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{
		"--host", host, "--port", strconv.Itoa(port), "--component", "nonexistent",
	}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	assert.Equal(t, cli.ExitOK, code)
	assert.Contains(t, out.String(), `No components matched --component "nonexistent"`)
}

func TestHealthUnhealthyOverallExitsFailure(t *testing.T) {
	const unhealthy = `{"healthy": false, "status": "StatusFatalError", "status_time": "2024-01-18T17:27:32Z"}`
	host, port := newFakeExtension(t, unhealthy, http.StatusInternalServerError, nil)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	assert.Equal(t, cli.ExitFailure, code)
	assert.Contains(t, out.String(), "healthy: no")
}

func TestHealthUnreachableEndpointPrintsHint(t *testing.T) {
	// Reserve a port and immediately close it so nothing is listening.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().(*net.TCPAddr)
	require.NoError(t, l.Close())

	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", "127.0.0.1", "--port", strconv.Itoa(addr.Port)}, cli.IO{Out: &bytes.Buffer{}, Err: errBuf})

	assert.Equal(t, cli.ExitNotRunning, code)
	assert.Contains(t, errBuf.String(), "healthcheckv2 extension is enabled")
	assert.Contains(t, errBuf.String(), "--host")
	assert.Contains(t, errBuf.String(), "--port")
	assert.Contains(t, errBuf.String(), "--path")
}

func TestHealthMalformedResponsePrintsHint(t *testing.T) {
	host, port := newFakeExtension(t, "not json", http.StatusOK, nil)

	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: &bytes.Buffer{}, Err: errBuf})

	assert.Equal(t, cli.ExitFailure, code)
	assert.Contains(t, errBuf.String(), "did not return a valid healthcheckv2 status response")
	assert.Contains(t, errBuf.String(), "healthcheckv2 extension is enabled")
}

func TestHealthRejectsResponseMissingRequiredFields(t *testing.T) {
	host, port := newFakeExtension(t, `{"healthy": true}`, http.StatusOK, nil)

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: errBuf})

	assert.Equal(t, cli.ExitFailure, code)
	assert.Empty(t, out.String(), "must not print a fabricated status table for an invalid response")
	assert.Contains(t, errBuf.String(), "did not return a valid healthcheckv2 status response")
	assert.Contains(t, errBuf.String(), "healthcheckv2 extension is enabled")
}

func TestHealthRejectsNullNestedComponent(t *testing.T) {
	const withNullChild = `{
		"healthy": true,
		"status": "StatusOK",
		"status_time": "2024-01-18T17:27:12Z",
		"components": {
			"extensions": {
				"healthy": true,
				"status": "StatusOK",
				"status_time": "2024-01-18T17:27:12Z",
				"components": {
					"foo": null
				}
			}
		}
	}`

	for _, args := range [][]string{
		{},
		{"--component", "foo"},
		{"--component", "nonexistent"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			host, port := newFakeExtension(t, withNullChild, http.StatusOK, nil)

			out := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			fullArgs := append([]string{"--host", host, "--port", strconv.Itoa(port)}, args...)

			var code int
			require.NotPanics(t, func() {
				code = New().Dispatch("health", fullArgs, cli.IO{Out: out, Err: errBuf})
			})

			assert.Equal(t, cli.ExitFailure, code)
			assert.Empty(t, out.String(), "must not print an incomplete status table as a successful check")
			assert.Contains(t, errBuf.String(), "did not return a valid healthcheckv2 status response")
			assert.Contains(t, errBuf.String(), `component "foo" is null`)
		})
	}
}

func TestHealthUnknownPipelineReturns404Hint(t *testing.T) {
	host, port := newFakeExtension(t, "", http.StatusNotFound, nil)

	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{
		"--host", host, "--port", strconv.Itoa(port), "--pipeline", "bogus",
	}, cli.IO{Out: &bytes.Buffer{}, Err: errBuf})

	assert.Equal(t, cli.ExitFailure, code)
	assert.Contains(t, errBuf.String(), "404 Not Found")
}

func TestHealthUnexpectedArgument(t *testing.T) {
	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"extra"}, cli.IO{Out: &bytes.Buffer{}, Err: errBuf})

	assert.Equal(t, cli.ExitUsage, code)
	assert.Contains(t, errBuf.String(), `unexpected argument "extra"`)
}

func TestHealthInvalidFlag(t *testing.T) {
	errBuf := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--port", "not-a-number"}, cli.IO{Out: &bytes.Buffer{}, Err: errBuf})

	assert.Equal(t, cli.ExitUsage, code)
	assert.True(t, strings.Contains(errBuf.String(), "invalid value") || strings.Contains(errBuf.String(), "Usage"))
}

func TestRenderExactFormat(t *testing.T) {
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, nil)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: &bytes.Buffer{}})
	require.Equal(t, cli.ExitOK, code)

	target := fmt.Sprintf("http://%s:%d/status?verbose", host, port)
	want := fmt.Sprintf(`Queried: %s
Overall status: RecoverableError (healthy: yes)
Error: rpc error: code = ResourceExhausted desc = resource exhausted
Started: 2024-01-18T17:27:12-08:00
As of: 2024-01-18T17:27:32-08:00

COMPONENT                     STATUS            HEALTHY
extensions                    OK                yes
  extension:healthcheckv2     OK                yes
pipeline:metrics/grpc         RecoverableError  yes
  error: rpc error: code = ResourceExhausted desc = resource exhausted
  exporter:otlp_grpc/staging  RecoverableError  yes
    error: rpc error: code = ResourceExhausted desc = resource exhausted
  receiver:otlp               OK                yes
`, target)

	assert.Equal(t, want, out.String())
}

func TestRenderNoTrailingWhitespaceAndConsistentAlignment(t *testing.T) {
	host, port := newFakeExtension(t, verboseCollectorResponse, http.StatusOK, nil)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", host, "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: &bytes.Buffer{}})
	require.Equal(t, cli.ExitOK, code)

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")

	var headerStatusCol int
	var sawDataRow bool
	for _, line := range lines {
		assert.Equal(t, line, strings.TrimRight(line, " \t"), "line has trailing whitespace: %q", line)

		switch {
		case strings.HasPrefix(line, "COMPONENT"):
			headerStatusCol = strings.Index(line, "STATUS")
			require.Positive(t, headerStatusCol)
		case headerStatusCol == 0, strings.Contains(line, "error:"):
			// Not in the table yet, or a detail line with no columns to check.
		default:
			sawDataRow = true
			require.Greater(t, len(line), headerStatusCol, "row shorter than the header: %q", line)
			assert.NotEqualf(t, byte(' '), line[headerStatusCol],
				"row's STATUS column doesn't line up with the header: %q", line)
		}
	}
	assert.True(t, sawDataRow, "expected at least one component row after the header")
}

func TestBuildURLEncodesVerboseAndPipeline(t *testing.T) {
	got := buildURL(options{host: "example.com", port: 9999, path: "/status", pipeline: "traces"})
	assert.Equal(t, "http://example.com:9999/status?pipeline=traces&verbose", got)
}

func TestBuildURLDefaultsOmitPipeline(t *testing.T) {
	got := buildURL(options{host: defaultHost, port: defaultPort, path: defaultPath})
	assert.Equal(t, "http://localhost:13133/status?verbose", got)
}
