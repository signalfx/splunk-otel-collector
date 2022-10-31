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

package discoveryreceiver

import (
	"encoding/base64"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/service/servicetest"
	"go.uber.org/zap/zaptest"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestValidConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	collectorConfig, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, collectorConfig)

	assert.Equal(t, len(collectorConfig.Receivers), 1)

	cfg := collectorConfig.Receivers[config.NewComponentIDWithName(typeStr, "discovery-name")].(*Config)
	require.Equal(t, &Config{
		Receivers: map[config.ComponentID]ReceiverEntry{
			config.NewComponentIDWithName("smartagent", "redis"): {
				Config: map[string]any{
					"auth": "password",
					"host": "`host`",
					"port": "`port`",
					"type": "collectd/redis",
				},
				Status: &Status{
					Metrics: map[discovery.StatusType][]Match{
						discovery.Successful: {
							Match{
								Record: &LogRecord{
									Attributes: map[string]string{
										"attr_one": "attr_one_val",
										"attr_two": "attr_two_val",
									},
									SeverityText: "info",
									Body:         "smartagent/redis receiver successful status",
								},
								Strict:    "",
								Regexp:    ".*",
								Expr:      "",
								FirstOnly: true,
							},
						},
					},
					Statements: map[discovery.StatusType][]Match{
						discovery.Failed: {
							{
								Strict:    "",
								Regexp:    "ConnectionRefusedError",
								Expr:      "",
								FirstOnly: true,
								Record: &LogRecord{
									Attributes:   map[string]string{},
									SeverityText: "info",
									Body:         "container appears to not be accepting redis connections",
								},
							},
						},
						discovery.Partial: {
							{
								Strict:    "",
								Regexp:    "(WRONGPASS|NOAUTH|ERR AUTH)",
								Expr:      "",
								FirstOnly: false,
								Record: &LogRecord{
									Attributes:   nil,
									SeverityText: "warn",
									Body:         "desired log invalid auth log body",
								},
							},
						},
					},
				},
				ResourceAttributes: map[string]string{
					"receiver_attribute": "receiver_attribute_value",
				},
				Rule: "type == \"container\"",
			},
		},
		ReceiverSettings:    config.NewReceiverSettings(config.NewComponentIDWithName("discovery", "discovery-name")),
		LogEndpoints:        true,
		EmbedReceiverConfig: true,
		CorrelationTTL:      25 * time.Second,
		WatchObservers: []config.ComponentID{
			config.NewComponentID("an_observer"),
			config.NewComponentIDWithName("another_observer", "with_name"),
		},
	},
		cfg)
	require.NoError(t, cfg.Validate())
}

func TestInvalidConfigs(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)
	factory := NewFactory()
	factories.Receivers[typeStr] = factory

	tests := []struct{ name, expectedError string }{
		{name: "no_watch_observers", expectedError: "receiver \"discovery\" has invalid configuration: `watch_observers` must be defined and include at least one configured observer extension"},
		{name: "missing_status", expectedError: "receiver \"discovery\" has invalid configuration: receiver \"a_receiver\" validation failure: `status` must be defined and contain at least one `metrics` or `statements` mapping"},
		{name: "missing_status_metrics_and_statements", expectedError: "receiver \"discovery\" has invalid configuration: receiver \"a_receiver\" validation failure: `status` must be defined and contain at least one `metrics` or `statements` mapping"},
		{name: "invalid_status_types", expectedError: `receiver "discovery" has invalid configuration: receiver "a_receiver" validation failure: invalid status "unsupported". must be one of [successful partial failed]; invalid status "another_unsupported". must be one of [successful partial failed]`},
		{name: "multiple_status_match_types", expectedError: "receiver \"discovery\" has invalid configuration: receiver \"a_receiver\" validation failure: `metrics` status source type `successful` match type validation failed. Must provide one of [regexp strict expr] but received [strict regexp]; `statements` status source type `failed` match type validation failed. Must provide one of [regexp strict expr] but received [strict expr]"},
		{name: "reserved_receiver_creator", expectedError: `receiver "discovery" has invalid configuration: receiver "receiver_creator/with-name" validation failure: receiver cannot be a receiver_creator`},
		{name: "reserved_receiver_name", expectedError: `receiver "discovery" has invalid configuration: receiver "a_receiver/with-receiver_creator/in-name" validation failure: receiver name cannot contain "receiver_creator/"`},
		{name: "reserved_receiver_name_with_endpoint", expectedError: `receiver "discovery" has invalid configuration: receiver "receiver/with{endpoint=}/" validation failure: receiver name cannot contain "{endpoint=[^}]*}/"`},
	}

	for _, test := range tests {
		func(name, expectedError string) {
			t.Run(name, func(t *testing.T) {
				_, err = servicetest.LoadConfigAndValidate(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", name)), factories)
				require.Error(t, err)
				require.EqualError(t, err, expectedError)
			})
		}(test.name, test.expectedError)
	}
}

func TestReceiverCreatorFactoryAndConfig(t *testing.T) {
	conf, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	conf, err = conf.Sub("receivers")
	require.NoError(t, err)
	require.NotEmpty(t, conf.ToStringMap())
	conf, err = conf.Sub("discovery/discovery-name")
	require.NoError(t, err)
	require.NotEmpty(t, conf.ToStringMap())
	dCfg := Config{ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName("discovery", "discovery-name"))}
	require.NoError(t, conf.Unmarshal(&dCfg, confmap.WithErrorUnused()))

	correlations := newCorrelationStore(zaptest.NewLogger(t), time.Second)
	factory, rCfg, err := dCfg.receiverCreatorFactoryAndConfig(correlations)
	require.NoError(t, err)
	require.Equal(t, config.Type("receiver_creator"), factory.Type())

	require.NoError(t, rCfg.Validate())
	require.Equal(t, config.NewComponentIDWithName("receiver_creator", "discovery/discovery-name"), rCfg.ID())

	creatorCfg, ok := rCfg.(*receivercreator.Config)
	require.True(t, ok)
	require.NotNil(t, creatorCfg)

	// receiver templates aren't exported so limited to WatchObservers
	require.Equal(t, []config.ComponentID{
		config.NewComponentID("an_observer"),
		config.NewComponentIDWithName("another_observer", "with_name"),
	}, creatorCfg.WatchObservers)

	receiverTemplate, err := dCfg.receiverCreatorReceiversConfig(correlations)
	require.NoError(t, err)
	expectedConfigHash := "cmVjZWl2ZXJzOgogIHNtYXJ0YWdlbnQvcmVkaXM6CiAgICBjb25maWc6CiAgICAgIGF1dGg6IHBhc3N3b3JkCiAgICAgIGhvc3Q6ICdgaG9zdGAnCiAgICAgIHBvcnQ6ICdgcG9ydGAnCiAgICAgIHR5cGU6IGNvbGxlY3RkL3JlZGlzCiAgICByZXNvdXJjZV9hdHRyaWJ1dGVzOgogICAgICByZWNlaXZlcl9hdHRyaWJ1dGU6IHJlY2VpdmVyX2F0dHJpYnV0ZV92YWx1ZQogICAgcnVsZTogdHlwZSA9PSAiY29udGFpbmVyIgo="
	expectedTemplate := map[string]any{
		"smartagent/redis": map[string]any{
			"config": map[string]any{
				"auth": "password",
				"host": "`host`",
				"port": "`port`",
				"type": "collectd/redis",
			},
			"resource_attributes": map[string]string{
				"discovery.endpoint.id":     "`id`",
				"discovery.receiver.config": expectedConfigHash,
				"discovery.receiver.name":   "redis",
				"discovery.receiver.rule":   `type == "container"`,
				"discovery.receiver.type":   "smartagent",
				"receiver_attribute":        "receiver_attribute_value",
			},
			"rule": `type == "container"`,
		},
	}
	require.Equal(t, expectedTemplate, receiverTemplate)

	decoded, err := base64.StdEncoding.DecodeString(expectedConfigHash)
	require.NoError(t, err)
	embedded := map[string]any{}
	yaml.Unmarshal(decoded, &embedded)
	require.Equal(t, map[string]any{
		"receivers": map[any]any{
			"smartagent/redis": map[any]any{
				"config": map[any]any{
					"auth": "password",
					"host": "`host`",
					"port": "`port`",
					"type": "collectd/redis",
				},
				"resource_attributes": map[any]any{
					"receiver_attribute": "receiver_attribute_value",
				},
				"rule": `type == "container"`,
			},
		},
	}, embedded)
}
