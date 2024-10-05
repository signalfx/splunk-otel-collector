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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

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
    tls:
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
			cc.Container = cc.Container.WithEntrypoint("sh").WillWaitForLogs("ok")
			return cc
		},
	)

	defer shutdown()

	sc, stdout, _ := c.Container.AssertExec(t, 15*time.Second,
		"sh", "-c", "/otelcol --dry-run 2>/dev/null",
	)
	require.Equal(t, config, stdout)
	require.Zero(t, sc)

	// confirm successful service functionality
	sc, _, _ = c.Container.AssertExec(t, 15*time.Second, "sh", "-c", "/otelcol &")
	require.Zero(t, sc)

	// confirm successful service functionality
	assert.Eventually(t, func() bool {
		if tc.OTLPReceiverSink.DataPointCount() == 0 {
			return false
		}
		receivedOTLPMetrics := tc.OTLPReceiverSink.AllMetrics()
		tc.OTLPReceiverSink.Reset()

		for _, rom := range receivedOTLPMetrics {
			for i := 0; i < rom.ResourceMetrics().Len(); i++ {
				rm := rom.ResourceMetrics().At(i)
				for j := 0; j < rm.ScopeMetrics().Len(); j++ {
					sm := rm.ScopeMetrics().At(j)
					if sm.Scope().Name() == "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper" {
						for k := 0; k < sm.Metrics().Len(); k++ {
							m := sm.Metrics().At(k)
							if m.Name() == "system.cpu.time" && m.Type() == pmetric.MetricTypeSum {
								sum := m.Sum()
								if sum.IsMonotonic() && sum.AggregationTemporality() == pmetric.AggregationTemporalityCumulative {
									return true
								}
							}
						}
					}
				}
			}
		}

		return false
	}, 30*time.Second, 10*time.Millisecond, "Failed to receive expected metrics")
}
