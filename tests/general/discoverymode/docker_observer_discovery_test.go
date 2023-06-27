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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestDockerObserver verifies basic discovery mode functionality within the collector container by
// starting a collector with the daemon domain socket mounted and the container running with its group id
// before starting a prometheus container with a label the receiver creator rule matches against.
func TestDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if runtime.GOOS == "darwin" {
		t.Skip("unable to share sockets between mac and d4m vm: https://github.com/docker/for-mac/issues/483#issuecomment-758836836")
	}
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"docker-otlp-exporter-no-internal-prometheus.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "docker-observer-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
			cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/dock.e.r.sock:ro")
			cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
			// uid check is for basic collector functionality not using the splunk-otel-collector user
			// but the docker gid is required to reach the daemon
			cc.Container = cc.Container.WithUser(fmt.Sprintf("%d:%d", os.Getuid(), testutils.GetDockerGID(t)))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				// runner seems to be slow
				"SPLUNK_DISCOVERY_DURATION": "20s",
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
				"DOCKER_DOMAIN_SOCKET":       "unix:///var/run/dock.e.r.sock",
				"LABEL_ONE_VALUE":            "actual.label.one.value",
				"LABEL_TWO_VALUE":            "actual.label.two.value",
				"SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_three": "overwritten by --set property",
				"SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_four":  "actual.label.four.value",
			}).WithArgs(
				"--discovery", "--config-dir", "/opt/config.d",
				"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
				"--set", `splunk.discovery.extensions.docker_observer.enabled=true`,
				"--set", `splunk.discovery.extensions.docker_observer.config.endpoint=${DOCKER_DOMAIN_SOCKET}`,
				"--set", `splunk.discovery.receivers.prometheus_simple.enabled=true`,
				"--set", `splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value`,
			)
		},
	)
	defer shutdown()

	cntrs, shutdownPrometheus := tc.Containers(
		testutils.NewContainer().WithImage("bitnami/prometheus").WithLabel("test.id", tc.ID).WillWaitForLogs("Server is ready to receive web requests."),
	)
	defer shutdownPrometheus()
	prometheus := cntrs[0]

	expectedResourceMetrics := tc.ResourceMetrics("docker-observer-internal-prometheus.yaml")
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
								"prometheus_tsdb_exemplar_exemplars_in_storage",
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
				"docker_observer": map[string]any{
					"endpoint": "${DOCKER_DOMAIN_SOCKET}",
				},
			},
			"receivers": map[string]any{
				"receiver_creator/discovery": map[string]any{
					"receivers": map[string]any{
						"prometheus_simple": map[string]any{
							"config": map[string]any{
								"collection_interval": "1s",
								"labels": map[string]any{
									"label_five":  "actual.label.five.value",
									"label_four":  "actual.label.four.value",
									"label_one":   "${LABEL_ONE_VALUE}",
									"label_three": "actual.label.three.value",
									"label_two":   "${LABEL_TWO_VALUE}",
								},
							},
							"resource_attributes": map[string]any{},
							"rule":                `type == "container" and labels['test.id'] == '${SPLUNK_TEST_ID}' and port == 9090`,
						},
					},
					"watch_observers": []any{"docker_observer"},
				},
			},
			"service": map[string]any{
				"extensions/splunk.discovery": []any{"docker_observer"},
				"receivers/splunk.discovery":  []any{"receiver_creator/discovery"},
			},
		},
		"splunk.property": map[string]any{},
	}
	require.Equal(t, expectedInitial, cc.InitialConfig(t, 55554))

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
							"prometheus_tsdb_exemplar_exemplars_in_storage",
						},
					},
				},
			},
		},
		"service": map[string]any{
			"extensions": []any{"docker_observer"},
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
			"docker_observer": map[string]any{
				"endpoint": "unix:///var/run/dock.e.r.sock",
			},
		},
		"receivers": map[string]any{
			"receiver_creator/discovery": map[string]any{
				"receivers": map[string]any{
					"prometheus_simple": map[string]any{
						"config": map[string]any{
							"collection_interval": "1s",
							"labels": map[string]any{
								"label_one":   "actual.label.one.value",
								"label_two":   "actual.label.two.value",
								"label_three": "actual.label.three.value",
								"label_four":  "actual.label.four.value",
								"label_five":  "actual.label.five.value",
							},
						},
						"resource_attributes": map[string]any{},
						"rule":                fmt.Sprintf(`type == "container" and labels['test.id'] == '%s' and port == 9090`, tc.ID),
					},
				},
				"watch_observers": []any{"docker_observer"},
			},
		},
	}
	require.Equal(t, expectedEffective, cc.EffectiveConfig(t, 55554))

	sc, stdout, stderr := cc.Container.AssertExec(t, 25*time.Second,
		"bash", "-c", `SPLUNK_DISCOVERY_LOG_LEVEL=error SPLUNK_DEBUG_CONFIG_SERVER=false \
SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_ENABLED=true \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_CONFIG_endpoint=\${DOCKER_DOMAIN_SOCKET} \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_ENABLED=true \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_three=="overwritten by --set property" \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_four="actual.label.four.value" \
/otelcol --config-dir /opt/config.d --discovery --dry-run \
--set splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value`)
	require.Equal(t, `exporters:
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    tls:
      insecure: true
extensions:
  docker_observer:
    endpoint: ${DOCKER_DOMAIN_SOCKET}
processors:
  filter:
    metrics:
      include:
        match_type: strict
        metric_names:
        - prometheus_tsdb_exemplar_exemplars_in_storage
receivers:
  receiver_creator/discovery:
    receivers:
      prometheus_simple:
        config:
          collection_interval: 1s
          labels:
            label_five: actual.label.five.value
            label_four: actual.label.four.value
            label_one: ${LABEL_ONE_VALUE}
            label_three: actual.label.three.value
            label_two: ${LABEL_TWO_VALUE}
        resource_attributes: {}
        rule: type == "container" and labels['test.id'] == '${SPLUNK_TEST_ID}' and
          port == 9090
    watch_observers:
    - docker_observer
service:
  extensions:
  - docker_observer
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
`, stdout)
	require.Contains(
		t, stderr,
		fmt.Sprintf(`Discovering for next 20s...
Successfully discovered "prometheus_simple" using "docker_observer" endpoint "%s:9090".
Discovery complete.
`, prometheus.GetContainerID()),
	)
	require.Zero(t, sc)
}
