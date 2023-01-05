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

//go:build integration

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDryRunDoesntExpandEnvVars(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	config := `config_sources:
  env:
    defaults:
      COLLECTION_INTERVAL: 1s
      OTLP_EXPORTER: otlp
exporters:
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    insecure: true
receivers:
  hostmetrics:
    collection_interval: ${env:COLLECTION_INTERVAL}
    scrapers:
      cpu: null
service:
  pipelines:
    metrics:
      exporters:
      - ${env:OTLP_EXPORTER}
      receivers:
      - ${HOST_METRICS_RECEIVER}
`

	c, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			// deferring running service for exec
			c := collector.WithEnv(
				map[string]string{
					"HOST_METRICS_RECEIVER": "hostmetrics",
					"SPLUNK_CONFIG":         "",
					"SPLUNK_CONFIG_YAML":    config,
				},
			).WithArgs("-c", "trap exit SIGTERM ; echo ok ; while true; do : ; done")
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithEntrypoint("bash").WillWaitForLogs("ok")
			return cc
		},
	)

	defer shutdown()

	sc, stdout, _ := c.Container.AssertExec(t, 15*time.Second,
		"bash", "-c", "/otelcol --dry-run 2>/dev/null",
	)
	require.Equal(t, config, stdout)
	require.Zero(t, sc)

	// confirm successful service functionality
	sc, _, _ = c.Container.AssertExec(t, 15*time.Second, "bash", "-c", "/otelcol &")
	require.Zero(t, sc)

	expectedResourceMetrics := tc.ResourceMetrics("cpu.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
