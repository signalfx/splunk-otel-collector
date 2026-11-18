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
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// enableExpvar turns on zpages expvar, which the shipped configs leave disabled,
// so that the tests can read the initial and effective config from /debug/expvarz.
const enableExpvar = "--set=extensions.zpages.expvar.enabled=true"

func TestDefaultGatewayConfig(t *testing.T) {
	for _, ip := range []string{"default", "0.0.0.0"} {
		ip := ip
		t.Run(ip, func(t *testing.T) {
			tc := testutils.NewTestcase(t)
			defer tc.PrintLogsOnFailure()
			defer tc.ShutdownOTLPReceiverSink()

			collector, shutdown := tc.SplunkOtelCollectorContainer(
				"",
				func(collector testutils.Collector) testutils.Collector {
					env := map[string]string{
						"SPLUNK_ACCESS_TOKEN": "not.real",
						"SPLUNK_REALM":        "not.real",
					}
					if ip != "default" {
						env["SPLUNK_LISTEN_INTERFACE"] = ip
					}
					return collector.WithArgs(enableExpvar).WithEnv(env)
				},
			)
			defer shutdown()

			// expected default
			if ip == "default" {
				ip = "0.0.0.0"
			}

			config := collector.EffectiveConfig(t)
			require.Equal(t, map[string]any{
				"exporters": map[string]any{
					"otlp_http": map[string]any{
						"traces_endpoint": "https://ingest.not.real.observability.splunkcloud.com/v2/trace/otlp",
						"sending_queue": map[string]any{
							"num_consumers": 32,
						},
						"headers": map[string]any{
							"X-SF-Token": "<redacted>",
						},
						"auth": map[string]any{
							"authenticator": "<redacted>",
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
						"endpoint":               "https://ingest.not.real.observability.splunkcloud.com/v1/log",
						"source":                 "otel",
						"sourcetype":             "otel",
						"token":                  "<redacted>",
						"profiling_data_enabled": false,
					},
					"splunk_hec/profiling": map[string]any{
						"endpoint":         "https://ingest.not.real.observability.splunkcloud.com/v1/log",
						"token":            "<redacted>",
						"log_data_enabled": false,
					},
					"otlp_http/entities": map[string]any{
						"logs_endpoint": "https://ingest.not.real.observability.splunkcloud.com/v3/event",
						"headers": map[string]any{
							"X-SF-Token": "<redacted>",
						},
						"auth": map[string]any{
							"authenticator": "<redacted>",
						},
					},
				},
				"extensions": map[string]any{
					"config_source_telemetry": map[string]any{},
					"headers_setter": map[string]any{
						"headers": []any{
							map[string]any{
								"action":        "upsert",
								"key":           "X-SF-TOKEN",
								"from_context":  "X-SF-TOKEN",
								"default_value": "not.real",
							},
						},
					},
					"health_check": map[string]any{
						"endpoint": fmt.Sprintf("%s:13133", ip),
					},
					"http_forwarder": map[string]any{
						"egress": map[string]any{
							"endpoint": "https://api.not.real.observability.splunkcloud.com",
						},
						"ingress": map[string]any{
							"endpoint": fmt.Sprintf("%s:6060", ip),
						},
					},
					"http_forwarder/opamp_splunk_o11y": map[string]any{
						"egress": map[string]any{
							"endpoint": "https://ingest.not.real.observability.splunkcloud.com",
							"headers": map[string]any{
								"X-SF-Token": "<redacted>",
							},
						},
						"ingress": map[string]any{
							"endpoint": fmt.Sprintf("%s:4320", ip),
						},
					},
					"http_forwarder/signalfx": map[string]any{
						"egress": map[string]any{
							"endpoint": "https://ingest.not.real.observability.splunkcloud.com",
						},
						"ingress": map[string]any{
							"endpoint": fmt.Sprintf("%s:9943", ip),
						},
					},
					"opamp/splunk_o11y": map[string]any{
						"agent_description": map[string]any{
							"include_resource_attributes": true,
						},
						"server": map[string]any{
							"http": map[string]any{
								"endpoint": "https://ingest.not.real.observability.splunkcloud.com/v1/opamp",
								"headers": map[string]any{
									"X-SF-Token": "<redacted>",
								},
								"polling_interval": "30s",
							},
						},
					},
					"zpages": map[string]any{
						"expvar": map[string]any{
							"enabled": true,
						},
					},
				},
				"processors": map[string]any{
					"batch": map[string]any{
						"metadata_keys": []any{"X-SF-Token"},
					},
					"memory_limiter": map[string]any{
						"check_interval": "2s",
						"limit_mib":      460,
					},
					"resource_detection/internal": map[string]any{
						"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
						"override":  true,
					},
					"transform/limit_histogram_buckets": map[string]any{
						"metric_statements": []any{
							map[string]any{
								"context":    "datapoint",
								"statements": []any{`merge_histogram_buckets(32, method="limit_buckets")`},
							},
						},
					},
				},
				"receivers": map[string]any{
					"jaeger": map[string]any{
						"protocols": map[string]any{
							"grpc":           map[string]any{"endpoint": fmt.Sprintf("%s:14250", ip)},
							"thrift_binary":  map[string]any{"endpoint": fmt.Sprintf("%s:6832", ip)},
							"thrift_compact": map[string]any{"endpoint": fmt.Sprintf("%s:6831", ip)},
							"thrift_http":    map[string]any{"endpoint": fmt.Sprintf("%s:14268", ip)},
						},
					},
					"otlp": map[string]any{
						"protocols": map[string]any{
							"grpc": map[string]any{"endpoint": fmt.Sprintf("%s:4317", ip)},
							"http": map[string]any{"endpoint": fmt.Sprintf("%s:4318", ip)},
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
											"regex":         "promhttp_metric_handler_errors.*",
											"source_labels": []any{"__name__"},
										},
										map[string]any{
											"action":        "drop",
											"regex":         "otelcol_processor_batch_.*",
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
					"zipkin": map[string]any{
						"endpoint": fmt.Sprintf("%s:9411", ip),
					},
				},
				"connectors": map[string]any{
					"routing/logs": map[string]any{
						"default_pipelines": []any{"logs"},
						"table": []any{
							map[string]any{
								"context":   "log",
								"condition": "instrumentation_scope.attributes[\"otel.entity.event_as_log\"] == true",
								"pipelines": []any{"logs/entities"},
							},
						},
					},
				},
				"service": map[string]any{
					"telemetry": map[string]any{
						"resource": map[string]any{
							"attributes": []any{
								map[string]any{
									"name":  "otelcol.service.mode",
									"value": "gateway",
								},
							},
						},
						"metrics": map[string]any{
							"readers": []any{
								map[string]any{
									"pull": map[string]any{
										"exporter": map[string]any{
											"prometheus": map[string]any{
												"host": "127.0.0.1",
												"port": 8888,
											},
										},
									},
								},
							},
						},
						"logs": map[string]any{
							"level": "info",
						},
					},
					"extensions": []any{"headers_setter", "health_check", "http_forwarder", "http_forwarder/opamp_splunk_o11y", "http_forwarder/signalfx", "opamp/splunk_o11y", "zpages", "config_source_telemetry"},
					"pipelines": map[string]any{
						"logs": map[string]any{
							"exporters":  []any{"splunk_hec", "splunk_hec/profiling"},
							"processors": []any{"memory_limiter", "batch"},
							"receivers":  []any{"routing/logs"},
						},
						"logs/entities": map[string]any{
							"exporters":  []any{"otlp_http/entities"},
							"processors": []any{"memory_limiter", "batch"},
							"receivers":  []any{"routing/logs"},
						},
						"logs/split": map[string]any{
							"receivers": []any{"otlp"},
							"exporters": []any{"routing/logs"},
						},
						"metrics": map[string]any{
							"exporters":  []any{"signalfx"},
							"processors": []any{"memory_limiter", "transform/limit_histogram_buckets", "batch"},
							"receivers":  []any{"otlp"},
						},
						"metrics/internal": map[string]any{
							"exporters":  []any{"signalfx/internal"},
							"processors": []any{"memory_limiter", "batch", "resource_detection/internal"},
							"receivers":  []any{"prometheus/internal"},
						},
						"traces": map[string]any{
							"exporters":  []any{"otlp_http"},
							"processors": []any{"memory_limiter", "batch"},
							"receivers":  []any{"jaeger", "otlp", "zipkin"},
						},
					},
				},
			}, config)
		})
	}
}

func TestDefaultAgentConfig(t *testing.T) {
	for _, ip := range []string{"default", "0.0.0.0"} {
		ip := ip
		t.Run(ip, func(t *testing.T) {
			tc := testutils.NewTestcase(t)
			defer tc.PrintLogsOnFailure()
			defer tc.ShutdownOTLPReceiverSink()

			collector, shutdown := tc.SplunkOtelCollectorContainer(
				"",
				func(collector testutils.Collector) testutils.Collector {
					env := map[string]string{
						"SPLUNK_ACCESS_TOKEN": "not.real",
						"SPLUNK_REALM":        "not.real",
					}
					if ip != "default" {
						env["SPLUNK_LISTEN_INTERFACE"] = ip
					}
					return collector.WithArgs(
						"--config", "/etc/otel/collector/agent_config.yaml",
						enableExpvar,
					).WithEnv(env)
				},
			)
			defer shutdown()

			// expected default
			if ip == "default" {
				ip = "127.0.0.1"
			}

			config := collector.EffectiveConfig(t)
			require.Equal(t, map[string]any{
				"exporters": map[string]any{
					"debug": map[string]any{
						"verbosity": "detailed",
					},
					"otlp_grpc/gateway": map[string]any{
						"endpoint": ":4317",
						"tls": map[string]any{
							"insecure": true,
						},
						"auth": map[string]any{
							"authenticator": "<redacted>",
						},
					},
					"otlp_http": map[string]any{
						"headers": map[string]any{
							"X-SF-Token": "<redacted>",
						},
						"traces_endpoint": "https://ingest.not.real.observability.splunkcloud.com/v2/trace/otlp",
						"auth": map[string]any{
							"authenticator": "<redacted>",
						},
					},
					"signalfx": map[string]any{
						"access_token":       "<redacted>",
						"api_url":            "https://api.not.real.observability.splunkcloud.com",
						"correlation":        nil,
						"ingest_url":         "https://ingest.not.real.observability.splunkcloud.com",
						"sync_host_metadata": true,
					},
					"splunk_hec": map[string]any{
						"endpoint":               "https://ingest.not.real.observability.splunkcloud.com/v1/log",
						"source":                 "otel",
						"sourcetype":             "otel",
						"token":                  "<redacted>",
						"profiling_data_enabled": false,
					},
					"splunk_hec/profiling": map[string]any{
						"endpoint":         "https://ingest.not.real.observability.splunkcloud.com/v1/log",
						"token":            "<redacted>",
						"log_data_enabled": false,
					},
					"otlp_http/entities": map[string]any{
						"logs_endpoint": "https://ingest.not.real.observability.splunkcloud.com/v3/event",
						"headers": map[string]any{
							"X-SF-Token": "<redacted>",
						},
						"auth": map[string]any{
							"authenticator": "<redacted>",
						},
					},
				},
				"extensions": map[string]any{
					"config_source_telemetry": map[string]any{},
					"headers_setter": map[string]any{
						"headers": []any{
							map[string]any{
								"action":        "upsert",
								"key":           "X-SF-TOKEN",
								"from_context":  "X-SF-TOKEN",
								"default_value": "not.real",
							},
						},
					},
					"health_check": map[string]any{"endpoint": fmt.Sprintf("%s:13133", ip)},
					"http_forwarder": map[string]any{
						"egress": map[string]any{
							"endpoint": "https://api.not.real.observability.splunkcloud.com",
						},
						"ingress": map[string]any{
							"endpoint": fmt.Sprintf("%s:6060", ip),
						},
					},
					"http_forwarder/opamp_splunk_o11y": map[string]any{
						"egress": map[string]any{
							"endpoint": "https://ingest.not.real.observability.splunkcloud.com",
							"headers": map[string]any{
								"X-SF-Token": "<redacted>",
							},
						},
						"ingress": map[string]any{
							"endpoint": fmt.Sprintf("%s:4320", ip),
						},
					},
					"opamp/splunk_o11y": map[string]any{
						"agent_description": map[string]any{
							"include_resource_attributes": true,
						},
						"server": map[string]any{
							"http": map[string]any{
								"endpoint": "https://ingest.not.real.observability.splunkcloud.com/v1/opamp",
								"headers": map[string]any{
									"X-SF-Token": "<redacted>",
								},
								"polling_interval": "30s",
							},
						},
					},
					"zpages": map[string]any{
						"expvar": map[string]any{
							"enabled": true,
						},
					},
				},
				"processors": map[string]any{
					"batch": map[string]any{
						"metadata_keys": []any{"X-SF-Token"},
					},
					"memory_limiter": map[string]any{
						"check_interval": "2s",
						"limit_mib":      460,
					},
					"resource_detection": map[string]any{
						"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
						"override":  true,
					},
					"transform/limit_histogram_buckets": map[string]any{
						"metric_statements": []any{
							map[string]any{
								"context":    "datapoint",
								"statements": []any{`merge_histogram_buckets(32, method="limit_buckets")`},
							},
						},
					},
				},
				"receivers": map[string]any{
					"host_metrics": map[string]any{
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
							"grpc":           map[string]any{"endpoint": fmt.Sprintf("%s:14250", ip)},
							"thrift_binary":  map[string]any{"endpoint": fmt.Sprintf("%s:6832", ip)},
							"thrift_compact": map[string]any{"endpoint": fmt.Sprintf("%s:6831", ip)},
							"thrift_http":    map[string]any{"endpoint": fmt.Sprintf("%s:14268", ip)},
						},
					},
					"otlp": map[string]any{
						"protocols": map[string]any{
							"grpc": map[string]any{"endpoint": fmt.Sprintf("%s:4317", ip)},
							"http": map[string]any{"endpoint": fmt.Sprintf("%s:4318", ip)},
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
											"regex":         "promhttp_metric_handler_errors.*",
											"source_labels": []any{"__name__"},
										},
										map[string]any{
											"action":        "drop",
											"regex":         "otelcol_processor_batch_.*",
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
					"smartagent/processlist": map[string]any{"type": "processlist"},
					"zipkin":                 map[string]any{"endpoint": fmt.Sprintf("%s:9411", ip)},
					"nop":                    nil,
				},
				"service": map[string]any{
					"extensions": []any{"headers_setter", "health_check", "http_forwarder", "http_forwarder/opamp_splunk_o11y", "opamp/splunk_o11y", "zpages", "config_source_telemetry"},
					"pipelines": map[string]any{
						"logs": map[string]any{
							"exporters":  []any{"splunk_hec", "splunk_hec/profiling"},
							"processors": []any{"memory_limiter", "batch", "resource_detection"},
							"receivers":  []any{"otlp"},
						},
						"logs/signalfx": map[string]any{
							"exporters":  []any{"signalfx"},
							"processors": []any{"memory_limiter", "batch", "resource_detection"},
							"receivers":  []any{"smartagent/processlist"},
						},
						"metrics": map[string]any{
							"exporters":  []any{"signalfx"},
							"processors": []any{"memory_limiter", "transform/limit_histogram_buckets", "batch", "resource_detection"},
							"receivers":  []any{"host_metrics", "otlp"},
						},
						"metrics/internal": map[string]any{
							"exporters":  []any{"signalfx"},
							"processors": []any{"memory_limiter", "batch", "resource_detection"},
							"receivers":  []any{"prometheus/internal"},
						},
						"traces": map[string]any{
							"exporters":  []any{"otlp_http"},
							"processors": []any{"memory_limiter", "batch", "resource_detection"},
							"receivers":  []any{"jaeger", "otlp", "zipkin"},
						},
						"logs/entities": map[string]any{
							"receivers":  []any{"nop"},
							"processors": []any{"memory_limiter", "batch", "resource_detection"},
							"exporters":  []any{"otlp_http/entities"},
						},
					},
					"telemetry": map[string]any{
						"resource": map[string]any{
							"attributes": []any{
								map[string]any{
									"name":  "otelcol.service.mode",
									"value": "agent",
								},
							},
						},
						"metrics": map[string]any{
							"readers": []any{
								map[string]any{
									"pull": map[string]any{
										"exporter": map[string]any{
											"prometheus": map[string]any{
												"host": "127.0.0.1",
												"port": 8888,
											},
										},
									},
								},
							},
						},
						"logs": map[string]any{
							"level": "info",
						},
					},
				},
			}, config)

			require.Eventually(t, func() bool {
				for _, log := range tc.ObservedLogs.All() {
					// confirm all components were loaded successfully
					if strings.Contains(log.Message, "Everything is ready. Begin running and processing data.") {
						return true
					}
				}
				return false
			}, 20*time.Second, time.Second)
		})
	}
}

// TestDefaultConfigsDoNotEnableExpvar ensures that none of the default configs expose
// the zpages expvar page, since it can leak the effective configuration.
func TestDefaultConfigsDoNotEnableExpvar(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "gateway"},
		{name: "agent", args: []string{"--config", "/etc/otel/collector/agent_config.yaml"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := testutils.NewTestcase(t)
			defer tc.PrintLogsOnFailure()
			defer tc.ShutdownOTLPReceiverSink()

			_, shutdown := tc.SplunkOtelCollectorContainer(
				"",
				func(collector testutils.Collector) testutils.Collector {
					return collector.WithArgs(tt.args...).WithEnv(map[string]string{
						"SPLUNK_ACCESS_TOKEN": "not.real",
						"SPLUNK_REALM":        "not.real",
					})
				},
			)
			defer shutdown()

			// servicez is always served by the zpages extension: use it to know that the
			// extension is up before asserting that expvarz is not being served.
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				assert.Equal(c, http.StatusOK, getStatus(t, "http://localhost:55679/debug/servicez"))
			}, 30*time.Second, 100*time.Millisecond)

			require.Equal(t, http.StatusNotFound, getStatus(t, "http://localhost:55679/debug/expvarz"))
		})
	}
}

func getStatus(t *testing.T, url string) int {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
