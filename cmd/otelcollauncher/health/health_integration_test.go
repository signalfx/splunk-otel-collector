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
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/baseline"
	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const configTemplate = `
receivers:
  nop:
exporters:
  nop:
extensions:
  healthcheckv2:
    use_v2: true
    http:
      endpoint: "127.0.0.1:%d"
      status:
        enabled: true
        path: "/status"
service:
  extensions: [healthcheckv2]
  pipelines:
    traces:
      receivers: [nop]
      exporters: [nop]
    metrics:
      receivers: [nop]
      exporters: [nop]
    logs:
      receivers: [nop]
      exporters: [nop]
  telemetry:
    logs:
      level: "error"
`

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startTestCollector(t *testing.T) int {
	t.Helper()

	port := freePort(t)
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(fmt.Sprintf(configTemplate, port)), 0o600))

	factories, err := baseline.NewBaseline().Build()
	require.NoError(t, err)

	col, err := otelcol.NewCollector(otelcol.CollectorSettings{
		BuildInfo:               component.BuildInfo{Command: "health-integration-test", Version: "test"},
		Factories:               func() (otelcol.Factories, error) { return factories, nil },
		DisableGracefulShutdown: true,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs:              []string{"file:" + cfgPath},
				ProviderFactories: []confmap.ProviderFactory{fileprovider.NewFactory()},
			},
		},
	})
	require.NoError(t, err)

	runErrCh := make(chan error, 1)
	go func() { runErrCh <- col.Run(context.Background()) }()
	t.Cleanup(func() {
		col.Shutdown()
		assert.NoError(t, <-runErrCh)
	})

	require.Eventually(t, func() bool {
		return col.GetState() == otelcol.StateRunning
	}, 10*time.Second, 50*time.Millisecond, "collector did not reach the running state")

	require.Eventually(t, func() bool {
		conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 200*time.Millisecond)
		if dialErr != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 5*time.Second, 50*time.Millisecond, "healthcheckv2 HTTP status server never started listening")

	return port
}

func TestHealthAgainstRealCollector(t *testing.T) {
	port := startTestCollector(t)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := New().Dispatch("health", []string{"--host", "127.0.0.1", "--port", strconv.Itoa(port)}, cli.IO{Out: out, Err: errOut})

	t.Logf("otelcollauncher health output:\n%s", out.String())

	assert.Empty(t, errOut.String())
	assert.Equal(t, cli.ExitOK, code)

	text := out.String()
	assert.Contains(t, text, "Overall status: OK")
	assert.Contains(t, text, "healthy: yes")
	assert.Contains(t, text, "extensions")
	assert.Contains(t, text, "extension:healthcheckv2")
	assert.Contains(t, text, "pipeline:traces")
	assert.Contains(t, text, "pipeline:metrics")
	assert.Contains(t, text, "pipeline:logs")
	assert.Contains(t, text, "receiver:nop")
	assert.Contains(t, text, "exporter:nop")
}

func TestHealthAgainstRealCollectorPipelineFilter(t *testing.T) {
	port := startTestCollector(t)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{
		"--host", "127.0.0.1", "--port", strconv.Itoa(port), "--pipeline", "traces",
	}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	t.Logf("otelcollauncher health --pipeline traces output:\n%s", out.String())

	assert.Equal(t, cli.ExitOK, code)
	text := out.String()
	assert.Contains(t, text, "receiver:nop")
	assert.Contains(t, text, "exporter:nop")
	assert.NotContains(t, text, "pipeline:metrics", "a pipeline-scoped query must not include other pipelines")
	assert.NotContains(t, text, "extension:healthcheckv2", "a pipeline-scoped query must not include extensions")
}

func TestHealthAgainstRealCollectorComponentFilter(t *testing.T) {
	port := startTestCollector(t)

	out := &bytes.Buffer{}
	code := New().Dispatch("health", []string{
		"--host", "127.0.0.1", "--port", strconv.Itoa(port), "--component", "healthcheckv2",
	}, cli.IO{Out: out, Err: &bytes.Buffer{}})

	t.Logf("otelcollauncher health --component healthcheckv2 output:\n%s", out.String())

	assert.Equal(t, cli.ExitOK, code)
	text := out.String()
	assert.Contains(t, text, "extension:healthcheckv2")
	assert.NotContains(t, text, "receiver:nop", "unrelated components must be pruned")
	assert.NotContains(t, text, "pipeline:traces", "unrelated pipelines must be pruned")
}
