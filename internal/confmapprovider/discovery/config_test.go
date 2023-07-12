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
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestServiceEntryPath(t *testing.T) {
	assert.True(t, isServiceEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("%cservice.yaml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf(".%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf(".%cservice.yaml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("config.d%cservice.yml", os.PathSeparator)))
	assert.True(t, isServiceEntryPath(fmt.Sprintf("config.d%cservice.yaml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cnot-service.yml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cnot-service.yaml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cs.yml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("%cs.yaml", os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("config.d%cdir%cservice.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isServiceEntryPath(fmt.Sprintf("config.d%cdic%cservice.yaml", os.PathSeparator, os.PathSeparator)))
}

func TestExporterEntryPaths(t *testing.T) {
	assert.True(t, isExporterEntryPath(fmt.Sprintf("%cexporters%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf("%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf(".%cexporters%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf(".%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf("config.d%cexporters%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExporterEntryPath(fmt.Sprintf("config.d%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("config.d%cdir%cexporters%cany.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExporterEntryPath(fmt.Sprintf("config.d%cdir%cexporters%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
}

func TestExtensionEntryPaths(t *testing.T) {
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf(".%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf(".%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("config.d%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isExtensionEntryPath(fmt.Sprintf("config.d%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("config.d%cdir%cextensions%cany.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isExtensionEntryPath(fmt.Sprintf("config.d%cdir%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
}

func TestProcessorEntryPaths(t *testing.T) {
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf(".%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf(".%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("config.d%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isProcessorEntryPath(fmt.Sprintf("config.d%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("%creceivers%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("config.d%cdir%cprocessors%cany.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isProcessorEntryPath(fmt.Sprintf("config.d%cdir%cprocessors%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
}

func TestReceiverEntryPaths(t *testing.T) {
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("config.d%creceivers%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf("config.d%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("config.d%cdir%creceivers%cany.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverEntryPath(fmt.Sprintf("config.d%cdir%creceivers%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverEntryPath(fmt.Sprintf(".%creceivers%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator)))
}
func TestDiscoveryObserverEntryPaths(t *testing.T) {
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf(".%cextensions%cany.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf(".%cextensions%cany.thing.at.all.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cservice.yml", os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cexporters%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cprocessors%cs.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cdir%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cdir%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.True(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isDiscoveryObserverEntryPath(fmt.Sprintf("%cextensions%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator)))
	assert.True(t, isDiscoveryObserverEntryPath(fmt.Sprintf("bundle.d%cextensions%ck8s-observer.discovery.yaml", os.PathSeparator, os.PathSeparator)))
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
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%cdir%creceivers%cany.thing.at.all.discovery.yml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.False(t, isReceiverToDiscoverEntryPath(fmt.Sprintf(".%cdir%creceivers%cany.thing.at.all.discovery.yaml", os.PathSeparator, os.PathSeparator, os.PathSeparator)))
	assert.True(t, isReceiverToDiscoverEntryPath(fmt.Sprintf("bundle.d%creceivers%csmartagent-postgresql.discovery.yaml", os.PathSeparator, os.PathSeparator)))
}

func TestDiscoveryPropertiesEntryPath(t *testing.T) {
	assert.True(t, isDiscoveryPropertiesEntryPath("properties.discovery.yml"))
	assert.True(t, isDiscoveryPropertiesEntryPath("properties.discovery.yaml"))
	assert.True(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cproperties.discovery.yml", os.PathSeparator)))
	assert.True(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cproperties.discovery.yaml", os.PathSeparator)))
	assert.True(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("config.d%cproperties.discovery.yml", os.PathSeparator)))
	assert.True(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("config.d%cproperties.discovery.yaml", os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("config.d%cdir%cproperties.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("config.d%cdir%cproperties.discovery.yaml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cnot-properties.discovery.yml", os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cnot-properties.discovery.yaml", os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%creceivers%cproperties.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cextensions%cproperties.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cexporters%cproperties.discovery.yml", os.PathSeparator, os.PathSeparator)))
	assert.False(t, isDiscoveryPropertiesEntryPath(fmt.Sprintf("%cprocessors%cproperties.discovery.yml", os.PathSeparator, os.PathSeparator)))
}

var (
	tru  = true
	flse = false
)

var expectedConfig = Config{
	Service: ServiceEntry{
		Entry{
			"extensions": []any{"zpages"},
			"telemetry": map[string]any{
				"logs": map[string]any{"level": "debug"},
			},
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"exporters": []any{"logging"},
				},
			},
		},
	},
	Exporters: map[component.ID]ExporterEntry{
		component.NewID("signalfx"): {Entry{
			"api_url":    "http://0.0.0.0/api",
			"ingest_url": "http://0.0.0.0/ingest",
		}},
	},
	Extensions: map[component.ID]ExtensionEntry{
		component.NewID("zpages"): {
			Entry{
				"endpoint": "0.0.0.0:1234",
			},
		},
	},
	DiscoveryObservers: map[component.ID]ObserverEntry{
		component.NewID("docker_observer"): {
			Enabled: &tru,
			Config: Entry{
				"endpoint": "tcp://debian:54321",
				"timeout":  "2s",
			},
		},
	},
	Processors: map[component.ID]ProcessorEntry{
		component.NewID("batch"): {},
	},
	Receivers: map[component.ID]ReceiverEntry{
		component.NewID("otlp"): {
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
	ReceiversToDiscover: map[component.ID]ReceiverToDiscoverEntry{
		component.NewIDWithName(component.Type("smartagent"), "postgresql"): {
			Enabled: &flse,
			Rule: map[component.ID]string{
				component.NewID("docker_observer"): `type == "container" and port == 5432`,
				component.NewID("host_observer"):   `type == "hostport" and command contains "pg" and port == 5432`,
			},

			Config: map[component.ID]map[string]any{
				defaultType: {
					"type":             "postgresql",
					"connectionString": "sslmode=disable user={{.username}} password={{.password}}",
					"params": map[any]any{
						"username": "test_user",
						"password": "test_password",
					},
					"masterDBName": "test_db",
				},
				component.NewID("docker_observer"): {
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
		component.NewIDWithName(component.Type("smartagent"), "collectd/redis"): {
			Rule: map[component.ID]string{
				component.NewID("docker_observer"): `type == "container" and port == 6379`,
			},

			Config: map[component.ID]map[string]any{
				defaultType: {
					"type": "collectd/redis",
					"auth": "password",
				},
				component.NewID("docker_observer"): {
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
	DiscoveryProperties: PropertiesEntry{
		Entry: map[string]any{
			"splunk.discovery.receivers.smartagent/postgresql.config.params.unused_param": "param_value",
			"splunk.discovery.extensions.docker_observer.config.timeout":                  "1s",
		},
	},
}

func TestConfig(t *testing.T) {
	configDir := filepath.Join(".", "testdata", "config.d")
	cfg := NewConfig(zap.NewNop())
	require.NotNil(t, cfg)
	require.NoError(t, cfg.Load(configDir))
	cfg.logger = nil // unset for equality check
	require.Equal(t, expectedConfig, *cfg)
}

var expectedServiceConfig = map[string]any{
	"exporters": map[string]any{
		"signalfx": map[string]any{
			"api_url":    "http://0.0.0.0/api",
			"ingest_url": "http://0.0.0.0/ingest",
		},
	}, "extensions": map[string]any{
		"zpages": map[string]any{
			"endpoint": "0.0.0.0:1234",
		},
	}, "processors": map[string]any{
		"batch": map[string]any{},
	}, "receivers": map[string]any{
		"otlp": map[string]any{
			"protocols": map[string]any{
				"grpc": map[string]any{
					"endpoint": "0.0.0.0:4317",
				}, "http": map[string]any{
					"endpoint": "0.0.0.0:4318",
				},
			},
		},
	}, "service": map[string]any{
		"extensions": []any{"zpages"},
		"pipelines": map[string]any{
			"metrics": map[string]any{
				"exporters": []any{"logging"}}},
		"telemetry": map[string]any{
			"logs": map[string]any{
				"level": "debug"},
		},
	},
}

func TestToServiceConfig(t *testing.T) {
	configDir := filepath.Join(".", "testdata", "config.d")
	cfg := NewConfig(zap.NewNop())
	require.NotNil(t, cfg)
	require.NoError(t, cfg.Load(configDir))
	sc := cfg.toServiceConfig()
	require.Equal(t, expectedServiceConfig, sc)
}

func TestInvalidConfigDirContents(t *testing.T) {
	for _, test := range []struct {
		configDir     string
		expectedError string
	}{
		{
			configDir:     "double-receiver-item-config.d",
			expectedError: "must contain a single mapping of ComponentID to component but contained [otlp otlp/disallowed]",
		},
		{
			configDir:     "invalid-properties.d",
			expectedError: "failed loading discovery.properties from properties.discovery.yaml: failed unmarshalling component discovery.properties: failed parsing \"properties.discovery.yaml\" as yaml",
		},
	} {
		t.Run(test.configDir, func(t *testing.T) {
			configDir := filepath.Join(".", "testdata", test.configDir)
			logger := zap.NewNop()
			cfg := NewConfig(logger)
			require.NotNil(t, cfg)
			err := cfg.Load(configDir)
			require.Contains(t, err.Error(), test.expectedError)
			cfg.logger = nil // unset for equality check
			require.Equal(t, Config{}, *cfg)
		})
	}
}
