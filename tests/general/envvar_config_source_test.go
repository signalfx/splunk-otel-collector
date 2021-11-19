// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"context"
	//"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestEffectiveConfigWithEnvVarConfigSource(t *testing.T) {
	image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE")
	if strings.TrimSpace(image) == "" {
		t.Skipf("skipping container-only test")
	}

	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)

	hostMetricsEnvVars := map[string]string{
		"HOST_METRICS_COLLECTION_INTERVAL": "10s",
		"HOST_METRICS_SCRAPERS_TO_EXPAND": `{ cpu: {},
	disk: {},
	filesystem: {},
	memory: {},
	network: {},
	load: {},
	paging: {},
	processes: {},
}`}

	collector, err := testutils.NewCollectorContainer().
		WithImage(image).
		WithEnv(hostMetricsEnvVars).
		WithConfigPath(path.Join(".", "testdata", "config_with_envvars.yaml")).
		WithLogger(logger).
		Build()

	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())
	defer func() { require.NoError(t, collector.Shutdown()) }()

	var collectorContainer interface{} = collector
	container := collectorContainer.(*testutils.CollectorContainer).Container
	// hijacking process stdout until https://github.com/testcontainers/testcontainers-go/issues/126 is resolved
	rc, err := container.Exec(context.Background(), []string{"sh", "-c", "curl http://localhost:55554/debug/configz/effective > /proc/1/fd/1"})
	require.NoError(t, err)
	require.Zero(t, rc)

	require.Eventually(t, func() bool {
		complete := ""
		for _, log := range logs.All() {
			complete += log.Message
		}
		return strings.Contains(complete, `exporters:
  logging:
    logLevel: error
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu: {}
      disk: {}
      filesystem: {}
      load: {}
      memory: {}
      network: {}
      paging: {}
      processes: {}
  hostmetrics/default-env-config-source:
    collection_interval: 10s
    scrapers:
      cpu: null
service:
  pipelines:
    metrics:
      exporters:
      - logging
      receivers:
      - hostmetrics
      - hostmetrics/default-env-config-source`,
		)
	}, 20*time.Second, time.Second)
}
