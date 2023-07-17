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
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/exp/slices"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestHostObserver verifies basic discovery mode functionality within the collector container by
// launching two collector processes:
// The first is the main collector process without internal prometheus metrics and `--discovery`
// w/ a config dir containing a host discovery observer and simple prometheus discovery config whose
// rule is for an "otelcol" process using a port of $INTERNAL_PROMETHEUS_PORT.
// The second process is exec'ed after the first successfully starts and is a collector process using
// a noop otlp receiver and logging exporter from another config.d with an internal prometheus server
// bound to $INTERNAL_PROMETHEUS_PORT.
// The test verifies that the second collector process's internal metrics contain a logging exporter
// one and that the first collector process's initial and effective configs from the config server
// are as expected.
func TestHostObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if testutils.CollectorImageIsForArm(t) {
		t.Skip("host_observer missing process info on arm")
	}
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	promPort := testutils.GetAvailablePort(t)

	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"host-otlp-exporter-no-internal-prometheus.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "host-observer-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
			cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			// We need to run another collector process in the container so
			// use a noop config.d w/ otlp receiver and logging exporter
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "logging-exporter-internal-prometheus-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/internal-prometheus-config.d"))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				"INTERNAL_PROMETHEUS_PORT": fmt.Sprintf("%d", promPort),
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
				"LABEL_ONE_VALUE":            "actual.label.one.value.from.env.var",
				"LABEL_TWO_VALUE":            "actual.label.two.value.from.env.var",
			}).WithArgs(
				"--discovery",
				"--config-dir", "/opt/config.d",
				"--set", "splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value.from.cmdline.property",
				"--set", "splunk.discovery.extensions.k8s_observer.enabled=false",
				"--set", "splunk.discovery.extensions.host_observer.config.refresh_interval=1s",
			)
		},
	)
	defer shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	sc, r, err := cc.Container.Exec(ctx, []string{
		// no config server to prevent port collisions
		"bash", "-c", "SPLUNK_DEBUG_CONFIG_SERVER=false /otelcol --configd --config-dir /opt/internal-prometheus-config.d &",
	})
	cancel()
	if r != nil {
		defer func() {
			if t.Failed() {
				out, readErr := io.ReadAll(r)
				require.NoError(t, readErr)
				fmt.Printf("exec'ed otelcol:\n%s\n", string(out))
			}
		}()
	}
	require.NoError(t, err)
	require.Zero(t, sc)

	// get the pid of the collector for endpoint ID verification
	sc, stdout, stderr := cc.Container.AssertExec(
		t, 5*time.Second, "bash", "-c", "ps -C otelcol | tail -n 1 | grep -oE '^\\s*[0-9]+'",
	)
	promPid := strings.TrimSpace(stdout)
	require.Zero(t, sc, stderr)

	expectedResourceMetrics := tc.ResourceMetrics("host-observer-internal-prometheus.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))

	expectedInitial := map[string]any{
		"file": map[string]any{
			"exporters": map[string]any{
				"otlp": map[string]any{
					"endpoint": "${OTLP_ENDPOINT}",
					"tls": map[string]any{
						"insecure": true,
					},
				},
			},
			"processors": map[string]any{
				"filter": map[string]any{
					"metrics": map[string]any{
						"include": map[string]any{
							"match_type": "strict",
							"metric_names": []any{
								"otelcol_exporter_enqueue_failed_log_records",
							},
						},
					},
				},
			},
			"service": map[string]any{
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"exporters":  []any{"otlp"},
						"processors": []any{"filter"},
					},
				},
				"telemetry": map[string]any{
					"metrics": map[string]any{
						"address": "",
						"level":   "none",
					},
				},
			},
		},
		"splunk.discovery": map[string]any{
			"extensions": map[string]any{
				"host_observer": map[string]any{
					"refresh_interval": "1s",
				},
				"host_observer/with-name": map[string]any{
					"refresh_interval": "1s",
				},
			},
			"receivers": map[string]any{
				"receiver_creator/discovery": map[string]any{
					"receivers": map[string]any{
						"prometheus_simple": map[string]any{
							"config": map[string]any{
								"collection_interval": "1s",
								"labels": map[string]any{
									"label_one":   "${LABEL_ONE_VALUE}",
									"label_two":   "${LABEL_TWO_VALUE}",
									"label_three": "actual.label.three.value.from.cmdline.property",
								},
							},
							"resource_attributes": map[string]any{},
							"rule":                `type == "hostport" and command contains "otelcol" and port == ${INTERNAL_PROMETHEUS_PORT}`,
						},
					},
					"watch_observers": []any{"host_observer", "host_observer/with-name"},
				},
			},
			"service": map[string]any{
				"extensions/splunk.discovery": []any{"host_observer", "host_observer/with-name"},
				"receivers/splunk.discovery":  []any{"receiver_creator/discovery"},
			},
		},
		"splunk.property": map[string]any{},
	}
	require.Equal(t, expectedInitial, cc.InitialConfig(t, 55554))

	if runtime.GOOS == "darwin" {
		t.Skip("docker for mac")
	}

	expectedEffective := map[string]any{
		"exporters": map[string]any{
			"otlp": map[string]any{
				"endpoint": tc.OTLPEndpointForCollector,
				"tls": map[string]any{
					"insecure": true,
				},
			},
		},
		"processors": map[string]any{
			"filter": map[string]any{
				"metrics": map[string]any{
					"include": map[string]any{
						"match_type": "strict",
						"metric_names": []any{
							"otelcol_exporter_enqueue_failed_log_records",
						},
					},
				},
			},
		},
		"service": map[string]any{
			"extensions": []any{"host_observer", "host_observer/with-name"},
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"receivers":  []any{"receiver_creator/discovery"},
					"exporters":  []any{"otlp"},
					"processors": []any{"filter"},
				},
			},
			"telemetry": map[string]any{
				"metrics": map[string]any{
					"address": "",
					"level":   "none",
				},
			},
		},
		"extensions": map[string]any{
			"host_observer": map[string]any{
				"refresh_interval": "1s",
			},
			"host_observer/with-name": map[string]any{
				"refresh_interval": "1s",
			},
		},
		"receivers": map[string]any{
			"receiver_creator/discovery": map[string]any{
				"receivers": map[string]any{
					"prometheus_simple": map[string]any{
						"config": map[string]any{
							"collection_interval": "1s",
							"labels": map[string]any{
								"label_one":   "actual.label.one.value.from.env.var",
								"label_two":   "actual.label.two.value.from.env.var",
								"label_three": "actual.label.three.value.from.cmdline.property",
							},
						},
						"resource_attributes": map[string]any{},
						"rule":                fmt.Sprintf(`type == "hostport" and command contains "otelcol" and port == %d`, promPort),
					},
				},
				"watch_observers": []any{"host_observer", "host_observer/with-name"},
			},
		},
	}
	require.Equal(t, expectedEffective, cc.EffectiveConfig(t, 55554))

	sc, stdout, stderr = cc.Container.AssertExec(t, 15*time.Second,
		"bash", "-c", `SPLUNK_DISCOVERY_LOG_LEVEL=error SPLUNK_DEBUG_CONFIG_SERVER=false \
REFRESH_INTERVAL=1s \
SPLUNK_DISCOVERY_DURATION=9s \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_simple_CONFIG_labels_x3a__x3a_label_three=actual.label.three.value.from.env.var.property \
SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_host_observer_CONFIG_refresh_interval=\$REFRESH_INTERVAL \
/otelcol --config-dir /opt/config.d --discovery --dry-run`)

	errorContent := fmt.Sprintf("unexpected --dry-run: %s", stderr)
	require.Equal(t, `exporters:
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    tls:
      insecure: true
extensions:
  host_observer:
    refresh_interval: $REFRESH_INTERVAL
  host_observer/with-name:
    refresh_interval: 1s
processors:
  filter:
    metrics:
      include:
        match_type: strict
        metric_names:
        - otelcol_exporter_enqueue_failed_log_records
receivers:
  receiver_creator/discovery:
    receivers:
      prometheus_simple:
        config:
          collection_interval: 1s
          labels:
            label_one: ${LABEL_ONE_VALUE}
            label_three: actual.label.three.value.from.env.var.property
            label_two: ${LABEL_TWO_VALUE}
        resource_attributes: {}
        rule: type == "hostport" and command contains "otelcol" and port == ${INTERNAL_PROMETHEUS_PORT}
    watch_observers:
    - host_observer
    - host_observer/with-name
service:
  extensions:
  - host_observer
  - host_observer/with-name
  pipelines:
    metrics:
      exporters:
      - otlp
      processors:
      - filter
      receivers:
      - receiver_creator/discovery
  telemetry:
    metrics:
      address: ""
      level: none
`, stdout, errorContent)

	split := strings.Split(stderr, "\n")
	start := slices.Index(split, "Discovering for next 9s...")
	require.GreaterOrEqual(t, start, 0, errorContent)
	require.GreaterOrEqual(t, len(split), start+3, errorContent)
	assert.Equal(t, split[start+3], "Discovery complete.", errorContent)

	statuses := split[start+1 : start+3]
	sort.Strings(statuses)
	var expectedStatuses []string
	for _, obs := range []string{"host_observer", "host_observer/with-name"} {
		expectedStatuses = append(expectedStatuses, fmt.Sprintf(`Successfully discovered "prometheus_simple" using "%s" endpoint "(%s)127.0.0.1-%d-TCP-%s".`, obs, obs, promPort, promPid))
	}
	require.Equal(t, expectedStatuses, statuses, errorContent)
	require.Zero(t, sc)
}
