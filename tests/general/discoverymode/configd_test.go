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
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestConfigDInitialAndEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cc, shutdown := tc.SplunkOtelCollectorContainer(
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
			}).WithArgs("--configd", "--set", "processors.batch/from-config-file.send_batch_size=123456789")
		},
	)

	defer shutdown()

	expectedInitial := map[string]any{
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
				"zpages": map[string]any{
					"endpoint": "0.0.0.0:55679",
					"expvar": map[string]any{
						"enabled": true,
					},
				}
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
				"extensions": []any{"health_check/from-config-file", "zpages"},
				"pipelines": map[string]any{
					"metrics/from-config-file": map[string]any{
						"exporters":  []any{"otlp/from-config-file"},
						"processors": []any{"batch/from-config-file"},
						"receivers":  []any{"otlp/from-config-file"},
					},
				},
			},
		},
		"splunk.configd": map[string]any{
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
	assert.Equal(t, expectedInitial, cc.InitialConfig(t))

	expectedEffective := map[string]any{
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
			"zpages": map[string]any{
				"endpoint": "0.0.0.0:55679",
				"expvar": map[string]any{
					"enabled": true,
				},
			},
		},
		"processors": map[string]any{
			"batch/from-config-file": map[string]any{
				"send_batch_size": 123456789,
			},
			"batch/from-configd": map[string]any{},
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
			"extensions": []any{"health_check/from-configd", "zpages"},
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

	assert.Equal(t, expectedEffective, cc.EffectiveConfig(t))

	sc, stdout, stderr := cc.Container.AssertExec(
		tc, 15*time.Second, "sh", "-c",
		"SPLUNK_DEBUG_CONFIG_SERVER=false /otelcol --config-dir /opt/config.d --configd --set processors.batch/from-config-file.send_batch_size=123456789 --dry-run 2>/dev/null",
	)
	assert.Equal(t, `exporters:
  otlp/from-config-file:
    endpoint: 0.0.0.0:${CONFIG_FILE_PORT_FROM_ENV_VAR}
  otlp/from-configd:
    endpoint: 0.0.0.0:${CONFIGD_PORT_FROM_ENV_VAR}
extensions:
  health_check/from-config-file:
    endpoint: 0.0.0.0:23456
  health_check/from-configd:
    endpoint: 0.0.0.0:45678
  zpages:
    endpoint: 0.0.0.0:55679
    expvar:
      enabled: true
processors:
  batch/from-config-file:
    send_batch_size: 123456789
  batch/from-configd: {}
receivers:
  otlp/from-config-file:
    protocols:
      http: null
  otlp/from-configd:
    protocols:
      grpc: null
service:
  extensions:
  - health_check/from-configd
  - zpages
  pipelines:
    metrics/from-config-file:
      exporters:
      - otlp/from-config-file
      processors:
      - batch/from-config-file
      receivers:
      - otlp/from-config-file
    metrics/from-configd:
      exporters:
      - otlp/from-configd
      processors:
      - batch/from-configd
      receivers:
      - otlp/from-configd
  telemetry:
    logs:
      level: debug
`, stdout)
	assert.Equal(t, "", stderr)
	assert.Zero(t, sc)
}

func TestStandaloneConfigD(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("empty-config.yaml",
		func(c testutils.Collector) testutils.Collector {
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "standalone-config.d"))
			require.NoError(t, err)
			if cc, ok := c.(*testutils.CollectorContainer); ok {
				cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
				configd = "/opt/config.d"
			}
			return c.WithEnv(map[string]string{"SPLUNK_CONFIG_DIR": configd}).WithArgs("--configd")
		},
	)
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "memory-expected.yaml"))
	require.NoError(t, err)
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1],
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
			pmetrictest.IgnoreResourceAttributeValue("net.host.name"),
			pmetrictest.IgnoreResourceAttributeValue("server.address"),
			pmetrictest.IgnoreResourceAttributeValue("container.name"),
			pmetrictest.IgnoreResourceAttributeValue("server.port"),
			pmetrictest.IgnoreResourceAttributeValue("service.name"),
			pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
			pmetrictest.IgnoreResourceAttributeValue("service_version"),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
		)
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)
}
