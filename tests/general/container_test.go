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

//go:build smartagent_integration

package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDefaultContainerConfigRequiresEnvVars(t *testing.T) {
	image := testutils.GetCollectorImageOrSkipTest(t)
	coverDest := os.Getenv("CONTAINER_COVER_DEST")
	coverSrc := os.Getenv("CONTAINER_COVER_SRC")

	tests := []struct {
		name    string
		env     map[string]string
		missing string
	}{
		{"missing realm", map[string]string{
			"SPLUNK_REALM":        "",
			"SPLUNK_ACCESS_TOKEN": "some_token",
			"GOCOVERDIR":          coverDest,
		}, "SPLUNK_REALM"},
		{"missing token", map[string]string{
			"SPLUNK_REALM":        "some_realm",
			"SPLUNK_ACCESS_TOKEN": "",
			"GOCOVERDIR":          coverDest,
		}, "SPLUNK_ACCESS_TOKEN"},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(tt *testing.T) {
			logCore, logs := observer.New(zap.DebugLevel)
			logger := zap.New(logCore)

			collector := testutils.NewCollectorContainer().WithImage(image).WithEnv(testcase.env).WithLogger(logger).WillFail(true)

			if coverSrc != "" && coverDest != "" {
				collector = collector.WithMount(coverSrc, coverDest)
			}

			var err error
			collector, err = collector.Build()
			require.NoError(t, err)
			require.NotNil(t, collector)
			defer collector.Shutdown()
			require.NoError(t, collector.Start())

			expectedError := fmt.Sprintf("ERROR: Missing required environment variable %s with default config path /etc/otel/collector/gateway_config.yaml", testcase.missing)
			require.Eventually(t, func() bool {
				for _, log := range logs.All() {
					if strings.Contains(log.Message, expectedError) {
						return true
					}
				}
				return false
			}, 30*time.Second, time.Second)
		})
	}
}

func TestSpecifiedContainerConfigDefaultsToCmdLineArgIfEnvVarConflict(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer(
		"hostmetrics_cpu.yaml",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs("--config", "/etc/config.yaml")
		},
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(
				map[string]string{
					"SPLUNK_CONFIG": "/not/a/real/path",
				},
			)
		},
	)
	defer shutdown()

	require.Eventually(t, func() bool {
		for _, log := range tc.ObservedLogs.All() {
			if strings.Contains(
				log.Message,
				`Both environment variable SPLUNK_CONFIG and flag '--config' were specified. `+
					`Using the flag values and ignoring the environment variable value `+
					`/not/a/real/path in this session`,
			) {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	// confirm successful service functionality
	assert.Eventually(t, func() bool {
		if tc.OTLPReceiverSink.DataPointCount() == 0 {
			return false
		}
		receivedOTLPMetrics := tc.OTLPReceiverSink.AllMetrics()

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

func TestConfigYamlEnvVar(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(
				map[string]string{
					"SPLUNK_CONFIG": "",
					"SPLUNK_CONFIG_YAML": `receivers:
  hostmetrics:
    collection_interval: 1s
    scrapers:
      cpu:

exporters:
  otlp:
    endpoint: "${OTLP_ENDPOINT}"
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [otlp]
`,
				},
			)
		},
	)

	defer shutdown()

	require.Eventually(t, func() bool {
		for _, log := range tc.ObservedLogs.All() {
			if strings.Contains(
				log.Message,
				`Using environment variable SPLUNK_CONFIG_YAML for configuration`,
			) {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	// confirm successful service functionality
	assert.Eventually(t, func() bool {
		if tc.OTLPReceiverSink.DataPointCount() == 0 {
			return false
		}
		receivedOTLPMetrics := tc.OTLPReceiverSink.AllMetrics()

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

func TestNonDefaultGIDCanAccessPythonInAgentBundle(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("couchbase_config.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithUser("splunk-otel-collector:234567890")
			return cc
		},
	)
	defer shutdown()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		metricsFound := map[string]struct{}{}
		m := tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1]
		for i := 0; i < m.ResourceMetrics().Len(); i++ {
			rm := m.ResourceMetrics().At(i)
			for j := 0; j < rm.ScopeMetrics().Len(); j++ {
				sm := rm.ScopeMetrics().At(j)
				for k := 0; k < sm.Metrics().Len(); k++ {
					metric := sm.Metrics().At(k)
					if metric.Name() == "gauge.storage.ram.quotaUsed" {
						metricsFound[metric.Name()] = struct{}{}
					}
				}
			}
		}
		assert.Equal(tt, 1, len(metricsFound))
	}, 30*time.Second, 1*time.Second)
}
