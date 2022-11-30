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
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestDockerObserver verifies basic discovery mode functionality within the collector container by
// starting a collector with the daemon domain socket mounted and the container running with its group id
// before starting a prometheus container with a label the receiver creator rule matches against.
func TestDockerObserver(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	finfo, err := os.Stat("/var/run/docker.sock")
	require.NoError(t, err)
	fsys := finfo.Sys()
	stat, ok := fsys.(*syscall.Stat_t)
	require.True(t, ok)
	dockerGID := stat.Gid

	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"otlp-exporter-no-internal-prometheus.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "docker-observer-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
			cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/dock.e.r.sock:ro")
			cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
			// uid check is for basic collector functionality not using the splunk-otel-collector user
			// but the docker gid is required to reach the daemon
			cc.Container = cc.Container.WithUser(fmt.Sprintf("%d:%d", os.Getuid(), dockerGID))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				// runner seems to be slow
				"SPLUNK_DISCOVERY_DURATION": "20s",
			}).WithArgs("--discovery", "--config-dir", "/opt/config.d")
		},
	)
	defer shutdown()

	_, shutdownPrometheus := tc.Containers(
		testutils.NewContainer().WithImage("bitnami/prometheus").WithLabel("test.id", tc.ID).WillWaitForLogs("Server is ready to receive web requests."),
	)
	defer shutdownPrometheus()

	expectedResourceMetrics := tc.ResourceMetrics("docker-observer-internal-prometheus.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))

	expected := map[string]any{
		"file": map[string]any{
			"exporters": map[string]any{
				"otlp": map[string]any{
					"endpoint": "${OTLP_ENDPOINT}",
					"insecure": true,
				},
			},
			"service": map[string]any{
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"exporters": []any{"otlp"},
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
					"endpoint": "unix:///var/run/dock.e.r.sock",
				},
			},
			"receivers": map[string]any{
				"receiver_creator/discovery": map[string]any{
					"receivers": map[string]any{
						"prometheus_simple": map[string]any{
							"config": map[string]any{
								"collection_interval": "1s",
							},
							"resource_attributes": map[string]any{},
							// TODO: support unexpanded env vars in config
							"rule": "placeholder",
						},
					},
					"watch_observers": []any{"docker_observer"},
				},
			},
			"service": map[string]any{
				"extensions": []any{"docker_observer"},
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"receivers": []any{"receiver_creator/discovery"},
					},
				},
			},
		},
	}

	actual := cc.InitialConfig(t, 55554)
	act := confmap.NewFromStringMap(actual)
	act.Merge(confmap.NewFromStringMap(
		map[string]any{
			"splunk.discovery": map[string]any{
				"receivers": map[string]any{
					"receiver_creator/discovery": map[string]any{
						"receivers": map[string]any{
							"prometheus_simple": map[string]any{
								"rule": "placeholder",
							},
						},
					},
				},
			},
		},
	))

	require.Equal(t, expected, act.ToStringMap())

	expected = map[string]any{
		"exporters": map[string]any{
			"otlp": map[string]any{
				"endpoint": tc.OTLPEndpoint,
				"tls": map[string]any{
					"insecure": true,
				},
			},
		},
		"service": map[string]any{
			"extensions": []any{"docker_observer"},
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"receivers": []any{"receiver_creator/discovery"},
					"exporters": []any{"otlp"},
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
						},
						"resource_attributes": map[string]any{},
						"rule":                "placeholder",
					},
				},
				"watch_observers": []any{"docker_observer"},
			},
		},
	}

	actual = cc.EffectiveConfig(t, 55554)
	act = confmap.NewFromStringMap(actual)
	act.Merge(confmap.NewFromStringMap(
		map[string]any{
			"receivers": map[string]any{
				"receiver_creator/discovery": map[string]any{
					"receivers": map[string]any{
						"prometheus_simple": map[string]any{
							"rule": "placeholder",
						},
					},
				},
			},
		},
	))

	require.Equal(t, expected, act.ToStringMap())
}
