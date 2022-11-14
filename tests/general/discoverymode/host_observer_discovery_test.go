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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/confmap"

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
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	promPort := testutils.GetAvailablePort(t)

	c, shutdown := tc.SplunkOtelCollector(
		"otlp-exporter-no-internal-prometheus.yaml",
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
			}).WithArgs("--discovery", "--config-dir", "/opt/config.d")
		},
	)
	defer shutdown()
	cc := c.(*testutils.CollectorContainer)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sc, r, err := cc.Container.Exec(ctx, []string{
		// no config server to prevent port collisions
		"bash", "-c", "SPLUNK_DEBUG_CONFIG_SERVER=false /otelcol --configd --config-dir /opt/internal-prometheus-config.d &",
	})
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

	expectedResourceMetrics := tc.ResourceMetrics("host-observer-internal-prometheus.yaml")
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
				"host_observer": map[string]any{
					"refresh_interval": "1s",
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
							// TODO: support unexpanded env vars in config
							// "rule": `type == "hostport" and command contains "otelcol" and port == ${INTERNAL_PROMETHEUS_PORT}`,
						},
					},
					"watch_observers": []any{"host_observer"},
				},
			},
			"service": map[string]any{
				"extensions": []any{"host_observer"},
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
			"extensions": []any{"host_observer"},
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
			"host_observer": map[string]any{
				"refresh_interval": "1s",
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
				"watch_observers": []any{"host_observer"},
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
