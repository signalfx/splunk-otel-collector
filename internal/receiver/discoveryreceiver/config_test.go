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
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/servicetest"
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

	cfg := collectorConfig.Receivers[config.NewComponentID(typeStr)].(*Config)
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
					Metrics: map[string][]Match{
						"successful": {
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
					Statements: map[string][]Match{
						"failed": {
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
						"partial": {
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
		ReceiverSettings:    config.NewReceiverSettings(config.NewComponentID("discovery")),
		LogEndpoints:        true,
		EmbedReceiverConfig: true,
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
		{name: "invalid_status_types", expectedError: `receiver "discovery" has invalid configuration: receiver "a_receiver" validation failure: unsupported status "unsupported". must be one of [successful partial failed]; unsupported status "another_unsupported". must be one of [successful partial failed]`},
		{name: "multiple_status_match_types", expectedError: "receiver \"discovery\" has invalid configuration: receiver \"a_receiver\" validation failure: `metrics` status source type `successful` match type validation failed. Must provide one of [regexp strict expr] but received [strict regexp]; `statements` status source type `failed` match type validation failed. Must provide one of [regexp strict expr] but received [strict expr]"},
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
