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
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestConfigDInitialAndEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	collector, shutdown := tc.SplunkOtelCollector(
		"config-to-merge-with.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "merged-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))

			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				"SPLUNK_CONFIG_DIR":             "/opt/config.d",
				"CONFIG_FILE_PORT_FROM_ENV_VAR": "12345",
				"CONFIGD_PORT_FROM_ENV_VAR":     "34567",
			}).WithArgs("--configd")
		},
	)

	defer shutdown()

	cc := collector.(*testutils.CollectorContainer)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	n, r, err := cc.Container.Exec(ctx, []string{"curl", "-s", "http://localhost:55554/debug/configz/initial"})
	require.NoError(t, err)
	require.Zero(t, n)
	out, err := io.ReadAll(r)
	// strip control character from curl output
	require.True(t, len(out) >= 8, "invalid config server output")
	initial := strings.TrimSpace(string(out[8 : len(out)-1]))

	actual := map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(initial), &actual))
	expected := map[string]any{
		"file": map[string]any{
			"exporters": map[string]any{
				"otlp/from-config-file": map[string]any{
					"endpoint": "0.0.0.0:${CONFIG_FILE_PORT_FROM_ENV_VAR}",
				},
			},
			"extensions": map[string]any{
				"health_check/from-config-file": map[string]any{
					"endpoint": "0.0.0.0:23456",
				},
			},
			"processors": map[string]any{
				"batch/from-config-file": nil,
			},
			"receivers": map[string]any{
				"otlp/from-config-file": map[string]any{
					"protocols": map[string]any{
						"http": nil,
					},
				},
			},
			"service": map[string]any{
				"extensions": []any{"health_check/from-config-file"},
				"pipelines": map[string]any{
					"metrics/from-config-file": map[string]any{
						"exporters":  []any{"otlp/from-config-file"},
						"processors": []any{"batch/from-config-file"},
						"receivers":  []any{"otlp/from-config-file"},
					},
				},
			},
		},
		"splunk.config.d": map[string]any{
			"exporters": map[string]any{
				"otlp/from-configd": map[string]any{
					"endpoint": "0.0.0.0:${CONFIGD_PORT_FROM_ENV_VAR}",
				},
			},
			"extensions": map[string]any{
				"health_check/from-configd": map[string]any{
					"endpoint": "0.0.0.0:45678",
				},
			},
			"processors": map[string]any{
				"batch/from-configd": map[string]any{},
			},
			"receivers": map[string]any{
				"otlp/from-configd": map[string]any{
					"protocols": map[string]any{
						"grpc": nil,
					},
				},
			},
			"service": map[string]any{
				"extensions": []any{"health_check/from-configd"},
				"pipelines": map[string]any{
					"metrics/from-configd": map[string]any{
						"exporters":  []any{"otlp/from-configd"},
						"processors": []any{"batch/from-configd"},
						"receivers":  []any{"otlp/from-configd"},
					},
				},
				"telemetry": map[string]any{
					"logs": map[string]any{
						"level": "debug",
					},
				},
			},
		},
	}

	require.Equal(t, expected, confmap.NewFromStringMap(actual).ToStringMap())

	n, r, err = cc.Container.Exec(ctx, []string{"curl", "-s", "http://localhost:55554/debug/configz/effective"})
	require.NoError(t, err)
	require.Zero(t, n)
	out, err = io.ReadAll(r)
	require.True(t, len(out) >= 8, "invalid config server output")
	effective := strings.TrimSpace(string(out[8 : len(out)-1])) // strip control character

	actual = map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(effective), &actual))

	expected = map[string]any{
		"exporters": map[string]any{
			"otlp/from-config-file": map[string]any{
				"endpoint": "0.0.0.0:12345",
			},
			"otlp/from-configd": map[string]any{
				"endpoint": "0.0.0.0:34567",
			},
		},
		"extensions": map[string]any{
			"health_check/from-config-file": map[string]any{
				"endpoint": "0.0.0.0:23456",
			},
			"health_check/from-configd": map[string]any{
				"endpoint": "0.0.0.0:45678",
			},
		},
		"processors": map[string]any{
			"batch/from-config-file": nil,
			"batch/from-configd":     map[string]any{},
		},
		"receivers": map[string]any{
			"otlp/from-config-file": map[string]any{
				"protocols": map[string]any{
					"http": nil,
				},
			},
			"otlp/from-configd": map[string]any{
				"protocols": map[string]any{
					"grpc": nil,
				},
			},
		},
		"service": map[string]any{
			"extensions": []any{"health_check/from-configd"},
			"pipelines": map[string]any{
				"metrics/from-config-file": map[string]any{
					"exporters":  []any{"otlp/from-config-file"},
					"processors": []any{"batch/from-config-file"},
					"receivers":  []any{"otlp/from-config-file"},
				},
				"metrics/from-configd": map[string]any{
					"exporters":  []any{"otlp/from-configd"},
					"processors": []any{"batch/from-configd"},
					"receivers":  []any{"otlp/from-configd"},
				},
			},
			"telemetry": map[string]any{
				"logs": map[string]any{
					"level": "debug",
				},
			},
		},
	}

	require.Equal(t, expected, confmap.NewFromStringMap(actual).ToStringMap())
}

func TestStandaloneConfigD(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollector(
		"empty-config.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "standalone-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))

			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				"SPLUNK_CONFIG_DIR": "/opt/config.d",
			}).WithArgs("--configd")
		},
	)
	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("memory.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
