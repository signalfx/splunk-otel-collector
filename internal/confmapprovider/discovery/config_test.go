// Copyright  Splunk, Inc.
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

package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap/zaptest"
)

func TestServiceEntryPath(t *testing.T) {
	assert.True(t, isServiceEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("%cservice.yaml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf(".%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf(".%cservice.yaml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("dir%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("dir%cservice.yaml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cs.yml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cs.yaml", os.PathSeparator)))
}

func TestExporterEntryPaths(t *testing.T) {
	assert.True(t, isExporterEntryPath(fmt.Sprintf("%cexporters%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf("%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf(".%cexporters%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf(".%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
}

func TestExtensionEntryPaths(t *testing.T) {
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf(".%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf(".%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
}

func TestProcessorEntryPaths(t *testing.T) {
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf(".%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf(".%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
}

func TestReceiverEntryPaths(t *testing.T) {
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator)))
}

func TestReceiverToDiscoverEntryPaths(t *testing.T) {
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator)))
}

var expectedConfig = Config{
	Service: map[string]ServiceEntry{
		"telemetry": {
			Entry{
				"logs": map[any]any{"level": "debug"},
			},
		},
		"pipelines": {
			Entry{
				"metrics": map[any]any{
					"exporters": []any{"logging"},
				},
			},
		},
	},
	Exporters: map[config.ComponentID]ExporterEntry{
		config.NewComponentID("signalfx"): {Entry{
			"api_url":    "http://0.0.0.0/api",
			"ingest_url": "http://0.0.0.0/ingest",
		}},
	},
	Extensions: map[config.ComponentID]ExtensionEntry{
		config.NewComponentID("zpages"): {
			Entry{
				"endpoint": "0.0.0.0:1234",
			},
		},
	},
	DiscoveryObservers: map[config.ComponentID]ExtensionEntry{
		config.NewComponentID("docker_observer"): {
			Entry{
				"endpoint": "tcp://debian:54321",
				"timeout":  "2s",
			},
		},
	},
	Processors: map[config.ComponentID]ProcessorEntry{
		config.NewComponentID("batch"): {},
	},
	Receivers: map[config.ComponentID]ReceiverEntry{
		config.NewComponentID("otlp"): {
			Entry{
				"protocols": map[any]any{
					"grpc": map[any]any{
						"endpoint": "0.0.0.0:4317",
					},
					"http": map[any]any{
						"endpoint": "0.0.0.0:4318",
					},
				},
			},
		},
	},
	ReceiversToDiscover: map[config.ComponentID]ReceiverToDiscoverEntry{
		config.NewComponentIDWithName(config.Type("smartagent"), "postgresql"): {
			Rule: map[config.ComponentID]string{
				config.NewComponentID("docker_observer"): `type == "container" and port == 5432`,
				config.NewComponentID("host_observer"):   `type == "hostport" and command contains "pg" and port == 5432`,
			},

			Config: map[config.ComponentID]map[string]any{
				defaultType: {
					"type":             "postgresql",
					"connectionString": "sslmode=disable user={{.username}} password={{.password}}",
					"params": map[any]any{
						"username": "test_user",
						"password": "test_password",
					},
					"masterDBName": "test_db",
				},
				config.NewComponentID("docker_observer"): {
					"params": map[any]any{
						"password": "`labels[\"auth\"]`",
					},
				},
			},
			Entry: map[string]any{
				"status": map[any]any{
					"metrics": map[any]any{
						"successful": []any{
							map[any]any{
								"strict":     "postgres_block_hit_ratio",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body":          "postgresql SA receiver working!",
								},
							},
						},
					},
					"statements": map[any]any{
						"failed": []any{
							map[any]any{
								"regexp":     ".* connect: connection refused",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body":          "container appears to not be accepting postgres connections",
								},
							},
						},
						"partial": []any{
							map[any]any{
								"regexp":     ".*pq: password authentication failed for user.*",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body": "Please ensure that your password is correctly specified " +
										"in `splunk.discovery.receivers.smartagent/postgresql.config.params.username` and " +
										"`splunk.discovery.receivers.smartagent/postgresql.config.params.password`",
								},
							},
						},
					},
				},
			},
		},
		config.NewComponentIDWithName(config.Type("smartagent"), "collectd/redis"): {
			Rule: map[config.ComponentID]string{
				config.NewComponentID("docker_observer"): `type == "container" and port == 6379`,
			},

			Config: map[config.ComponentID]map[string]any{
				defaultType: {
					"type": "collectd/redis",
					"auth": "password",
				},
				config.NewComponentID("docker_observer"): {
					"auth": "`labels[\"auth\"]`",
				},
			},
			Entry: map[string]any{
				"status": map[any]any{
					"metrics": map[any]any{
						"successful": []any{
							map[any]any{
								"regexp":     ".*",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body":          "smartagent/collectd-redis receiver successful metric status",
								},
							},
						},
					},
					"statements": map[any]any{
						"failed": []any{
							map[any]any{
								"regexp":     `raise ValueError\(\"Unknown Redis response`,
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body":          "container appears to not actually be redis",
								},
							},
							map[any]any{
								"regexp":     "^redis_info plugin: Error connecting to .* - ConnectionRefusedError.*$",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body":          "container appears to not be accepting redis connections",
								},
							},
						},
						"partial": []any{
							map[any]any{
								"regexp":     "^redis_info plugin: Error .* - RedisError\\('-(WRONGPASS|NOAUTH|ERR AUTH).*$",
								"first_only": true,
								"log_record": map[any]any{
									"severity_text": "info",
									"body": "Please ensure that your redis password is correctly specified in " +
										"`splunk.discovery.receivers.smartagent/collectd/redis.config.auth` or via the " +
										"`SPLUNK_DISCOVERY_RECEIVERS_SMARTAGENT_COLLECTD_REDIS_CONFIG_AUTH` environment variable.",
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestDiscoveryConfig(t *testing.T) {
	discoveryDir := filepath.Join(".", "testdata", "config.d")
	cfg := NewConfig(zaptest.NewLogger(t))
	require.NotNil(t, cfg)
	require.NoError(t, cfg.Load(discoveryDir))
	cfg.logger = nil // unset for equality check
	require.Equal(t, expectedConfig, *cfg)
}

func TestDiscoveryConfigWithTwoReceiversInOneFile(t *testing.T) {
	discoveryDir := filepath.Join(".", "testdata", "double-receiver-item-config.d")
	logger := zaptest.NewLogger(t)
	cfg := NewConfig(logger)
	require.NotNil(t, cfg)
	err := cfg.Load(discoveryDir)
	require.Contains(t, err.Error(), "must contain a single mapping of ComponentID to component but contained [otlp otlp/disallowed]")
	cfg.logger = nil // unset for equality check
	require.Equal(t, Config{}, *cfg)
}
