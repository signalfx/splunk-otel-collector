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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestSetProperties(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	configServerPort := testutils.GetAvailablePort(t)
	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"set_properties_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			return collector.
				WithArgs(
					"--config", "/etc/config.yaml",
					"--set", "receivers.otlp.protocols.grpc.max_recv_msg_size_mib=1",
					"--set=receivers.otlp.protocols.http.endpoint=localhost:0",
					"--set", "processors.filter/one.metrics.include.match_type=regexp",
					"--set=processors.filter/one.metrics.include.metric_names=[a.name]",
					"--set", "processors.filter/two.metrics.include.match_type=strict",
					"--set=processors.filter/two.metrics.include.metric_names=[another.name]",
				).
				WithEnv(
					map[string]string{
						"SPLUNK_DEBUG_CONFIG_SERVER_PORT": fmt.Sprintf("%d", configServerPort),
					},
				)
		},
	)
	defer shutdown()

	expectedInitial := map[string]any{
		"file": map[string]any{
			"receivers": map[string]any{
				"otlp": map[string]any{"protocols": "overwritten"},
			},
			"processors": map[string]any{
				"filter/one": map[string]any{
					"metrics": map[string]any{
						"include": map[string]any{
							"match_type":   "overwritten",
							"metric_names": "overwritten",
						},
					},
				},
				"filter/two": map[string]any{
					"metrics": map[string]any{
						"include": map[string]any{
							"match_type":   "overwritten",
							"metric_names": "overwritten",
						},
					},
				},
			},
			"exporters": map[string]any{
				"otlp": map[string]any{
					"endpoint": "${OTLP_ENDPOINT}",
					"tls": map[string]any{
						"insecure": true,
					},
				},
			},
			"service": map[string]any{
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"receivers":  []any{"otlp"},
						"processors": []any{"filter/one", "filter/two"},
						"exporters":  []any{"otlp"},
					},
				},
			},
		},
	}

	require.Equal(t, expectedInitial, cc.InitialConfig(t, configServerPort))

	expectedEffective := map[string]any{
		"receivers": map[string]any{
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": map[string]any{
						"max_recv_msg_size_mib": 1,
					},
					"http": map[string]any{
						"endpoint": "localhost:0",
					},
				},
			},
		},
		"processors": map[string]any{
			"filter/one": map[string]any{
				"metrics": map[string]any{
					"include": map[string]any{
						"match_type":   "regexp",
						"metric_names": []any{"a.name"},
					},
				},
			},
			"filter/two": map[string]any{
				"metrics": map[string]any{
					"include": map[string]any{
						"match_type":   "strict",
						"metric_names": []any{"another.name"},
					},
				},
			},
		},
		"exporters": map[string]any{
			"otlp": map[string]any{
				"endpoint": tc.OTLPEndpointForCollector,
				"tls": map[string]any{
					"insecure": true,
				},
			},
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"receivers":  []any{"otlp"},
					"processors": []any{"filter/one", "filter/two"},
					"exporters":  []any{"otlp"},
				},
			},
		},
	}

	require.Equal(t, expectedEffective, cc.EffectiveConfig(t, configServerPort))
}
