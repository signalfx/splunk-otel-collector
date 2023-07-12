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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDefaultGatewayConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs().WithEnv(map[string]string{
				"SPLUNK_ACCESS_TOKEN": "not.real",
				"SPLUNK_REALM":        "not.real",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t, 55554)
	require.Equal(t, map[string]any{
		"exporters": map[string]any{
			"sapm": map[string]any{
				"access_token": "<redacted>",
				"endpoint":     "https://ingest.not.real.signalfx.com/v2/trace",
				"sending_queue": map[string]any{
					"num_consumers": 32,
				},
			},
			"signalfx": map[string]any{
				"access_token": "<redacted>",
				"realm":        "not.real",
				"sending_queue": map[string]any{
					"num_consumers": 32,
				},
			},
			"signalfx/internal": map[string]any{
				"access_token":       "<redacted>",
				"realm":              "not.real",
				"sync_host_metadata": true,
			},
			"splunk_hec": map[string]any{
				"endpoint":               "https://ingest.not.real.signalfx.com/v1/log",
				"source":                 "otel",
				"sourcetype":             "otel",
				"token":                  "<redacted>",
				"profiling_data_enabled": false,
			},
			"splunk_hec/profiling": map[string]any{
				"endpoint":         "https://ingest.not.real.signalfx.com/v1/log",
				"token":            "<redacted>",
				"log_data_enabled": false,
			},
		},
		"extensions": map[string]any{
			"health_check": map[string]any{
				"endpoint": "0.0.0.0:13133",
			},
			"http_forwarder": map[string]any{
				"egress": map[string]any{
					"endpoint": "https://api.not.real.signalfx.com",
				},
				"ingress": map[string]any{
					"endpoint": "0.0.0.0:6060",
				},
			},
			"memory_ballast": map[string]any{
				"size_mib": "168",
			},
			"zpages": map[string]any{
				"endpoint": "0.0.0.0:55679",
			},
		},
		"processors": map[string]any{
			"batch": nil,
			"memory_limiter": map[string]any{
				"check_interval": "2s",
				"limit_mib":      "460",
			},
			"resourcedetection/internal": map[string]any{
				"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
				"override":  true,
			},
		},
		"receivers": map[string]any{
			"jaeger": map[string]any{
				"protocols": map[string]any{
					"grpc":           map[string]any{"endpoint": "0.0.0.0:14250"},
					"thrift_binary":  map[string]any{"endpoint": "0.0.0.0:6832"},
					"thrift_compact": map[string]any{"endpoint": "0.0.0.0:6831"},
					"thrift_http":    map[string]any{"endpoint": "0.0.0.0:14268"},
				},
			},
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": map[string]any{"endpoint": "0.0.0.0:4317"},
					"http": map[string]any{"endpoint": "0.0.0.0:4318"},
				},
			},
			"prometheus/internal": map[string]any{
				"config": map[string]any{
					"scrape_configs": []any{map[string]any{
						"job_name": "otel-collector",
						"metric_relabel_configs": []any{
							map[string]any{
								"action":        "drop",
								"regex":         ".*grpc_io.*",
								"source_labels": []any{"__name__"},
							},
						},
						"scrape_interval": "10s",
						"static_configs": []any{
							map[string]any{
								"targets": []any{"0.0.0.0:8888"},
							},
						},
					},
					},
				},
			},
			"sapm": map[string]any{
				"endpoint": "0.0.0.0:7276",
			},
			"signalfx": map[string]any{
				"endpoint": "0.0.0.0:9943",
			},
			"zipkin": map[string]any{
				"endpoint": "0.0.0.0:9411",
			},
		},
		"service": map[string]any{
			"extensions": []any{"health_check", "http_forwarder", "zpages", "memory_ballast"},
			"pipelines": map[string]any{
				"logs": map[string]any{
					"exporters":  []any{"splunk_hec", "splunk_hec/profiling"},
					"processors": []any{"memory_limiter", "batch"},
					"receivers":  []any{"otlp"},
				},
				"logs/signalfx": map[string]any{
					"exporters":  []any{"signalfx"},
					"processors": []any{"memory_limiter", "batch"},
					"receivers":  []any{"signalfx"},
				},
				"metrics": map[string]any{
					"exporters":  []any{"signalfx"},
					"processors": []any{"memory_limiter", "batch"},
					"receivers":  []any{"otlp", "signalfx"},
				},
				"metrics/internal": map[string]any{
					"exporters":  []any{"signalfx/internal"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection/internal"},
					"receivers":  []any{"prometheus/internal"},
				},
				"traces": map[string]any{
					"exporters":  []any{"sapm"},
					"processors": []any{"memory_limiter", "batch"},
					"receivers":  []any{"jaeger", "otlp", "sapm", "zipkin"},
				},
			},
		},
	}, config)
}

func TestDefaultAgentConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs(
				"--config", "/etc/otel/collector/agent_config.yaml",
			).WithEnv(map[string]string{
				"SPLUNK_ACCESS_TOKEN": "not.real",
				"SPLUNK_REALM":        "not.real",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t, 55554)
	require.Equal(t, map[string]any{
		"exporters": map[string]any{
			"logging": map[string]any{
				"verbosity": "detailed",
			},
			"otlp": map[string]any{
				"endpoint": ":4317",
				"tls": map[string]any{
					"insecure": true,
				},
			},
			"sapm": map[string]any{
				"access_token": "<redacted>",
				"endpoint":     "https://ingest.not.real.signalfx.com/v2/trace",
			},
			"signalfx": map[string]any{
				"access_token":       "<redacted>",
				"api_url":            "https://api.not.real.signalfx.com",
				"correlation":        nil,
				"ingest_url":         "https://ingest.not.real.signalfx.com",
				"sync_host_metadata": true,
			},
			"splunk_hec": map[string]any{
				"endpoint":               "https://ingest.not.real.signalfx.com/v1/log",
				"source":                 "otel",
				"sourcetype":             "otel",
				"token":                  "<redacted>",
				"profiling_data_enabled": false,
			},
			"splunk_hec/profiling": map[string]any{
				"endpoint":         "https://ingest.not.real.signalfx.com/v1/log",
				"token":            "<redacted>",
				"log_data_enabled": false,
			},
		},
		"extensions": map[string]any{
			"health_check": map[string]any{"endpoint": "0.0.0.0:13133"},
			"http_forwarder": map[string]any{
				"egress": map[string]any{
					"endpoint": "https://api.not.real.signalfx.com",
				},
				"ingress": map[string]any{
					"endpoint": "0.0.0.0:6060",
				},
			},
			"memory_ballast": map[string]any{"size_mib": "168"},
			"smartagent": map[string]any{
				"bundleDir": "/usr/lib/splunk-otel-collector/agent-bundle",
				"collectd": map[string]any{
					"configDir": "/usr/lib/splunk-otel-collector/agent-bundle/run/collectd",
				},
			},
			"zpages": nil,
		},
		"processors": map[string]any{
			"batch": nil,
			"memory_limiter": map[string]any{
				"check_interval": "2s",
				"limit_mib":      "460",
			},
			"resourcedetection": map[string]any{"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
				"override": true,
			}},
		"receivers": map[string]any{"fluentforward": map[string]any{"endpoint": "127.0.0.1:8006"},
			"hostmetrics": map[string]any{
				"collection_interval": "10s",
				"scrapers": map[string]any{
					"cpu":        nil,
					"disk":       nil,
					"filesystem": nil,
					"load":       nil,
					"memory":     nil,
					"network":    nil,
					"paging":     nil,
					"processes":  nil,
				},
			},
			"jaeger": map[string]any{
				"protocols": map[string]any{
					"grpc":           map[string]any{"endpoint": "0.0.0.0:14250"},
					"thrift_binary":  map[string]any{"endpoint": "0.0.0.0:6832"},
					"thrift_compact": map[string]any{"endpoint": "0.0.0.0:6831"},
					"thrift_http":    map[string]any{"endpoint": "0.0.0.0:14268"}}},
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": map[string]any{"endpoint": "0.0.0.0:4317"},
					"http": map[string]any{"endpoint": "0.0.0.0:4318"},
				},
			},
			"prometheus/internal": map[string]any{
				"config": map[string]any{
					"scrape_configs": []any{
						map[string]any{
							"job_name": "otel-collector",
							"metric_relabel_configs": []any{
								map[string]any{
									"action":        "drop",
									"regex":         ".*grpc_io.*",
									"source_labels": []any{"__name__"},
								},
							},
							"scrape_interval": "10s",
							"static_configs": []any{
								map[string]any{
									"targets": []any{"0.0.0.0:8888"},
								},
							},
						},
					},
				},
			},
			"signalfx":               map[string]any{"endpoint": "0.0.0.0:9943"},
			"smartagent/processlist": map[string]any{"type": "processlist"},
			"smartagent/signalfx-forwarder": map[string]any{
				"listenAddress": "0.0.0.0:9080",
				"type":          "signalfx-forwarder",
			},
			"zipkin": map[string]any{
				"endpoint": "0.0.0.0:9411"}},
		"service": map[string]any{
			"extensions": []any{"health_check", "http_forwarder", "zpages", "memory_ballast", "smartagent"},
			"pipelines": map[string]any{
				"logs": map[string]any{
					"exporters":  []any{"splunk_hec", "splunk_hec/profiling"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection"},
					"receivers":  []any{"fluentforward", "otlp"}},
				"logs/signalfx": map[string]any{
					"exporters":  []any{"signalfx"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection"},
					"receivers":  []any{"signalfx", "smartagent/processlist"}},
				"metrics": map[string]any{
					"exporters":  []any{"signalfx"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection"},
					"receivers":  []any{"hostmetrics", "otlp", "signalfx", "smartagent/signalfx-forwarder"}},
				"metrics/internal": map[string]any{
					"exporters":  []any{"signalfx"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection"},
					"receivers":  []any{"prometheus/internal"}},
				"traces": map[string]any{
					"exporters":  []any{"sapm", "signalfx"},
					"processors": []any{"memory_limiter", "batch", "resourcedetection"},
					"receivers":  []any{"jaeger", "otlp", "smartagent/signalfx-forwarder", "zipkin"},
				},
			},
		},
	}, config)

	require.Eventually(t, func() bool {
		for _, log := range tc.ObservedLogs.All() {
			// confirm the smartagent extension's config has been sourced by receiver instance.
			if strings.Contains(log.Message, "Smart Agent Config provider configured") {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)
}
