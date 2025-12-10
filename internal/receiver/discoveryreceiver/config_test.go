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
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

func TestValidConfig(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, configs)

	assert.Len(t, configs.ToStringMap(), 1)

	cm, err := configs.Sub("discovery")
	require.NoError(t, err)
	cfg := createDefaultConfig().(*Config)
	err = cm.Unmarshal(&cfg)
	require.NoError(t, err)

	require.Equal(t, &Config{
		Receivers: map[component.ID]ReceiverEntry{
			component.MustNewIDWithName("smartagent", "redis"): {
				Config: map[string]any{
					"auth": "password",
					"host": "`host`",
					"port": "`port`",
					"type": "collectd/redis",
				},
				ResourceAttributes: map[string]string{
					"receiver_attribute": "receiver_attribute_value",
				},
				Rule: mustNewRule(`type == "container" && name matches "(?i)redis"`),
			},
		},
		EmbedReceiverConfig: true,
		CorrelationTTL:      25 * time.Second,
		WatchObservers: []component.ID{
			component.MustNewID("an_observer"),
			component.MustNewIDWithName("another_observer", "with_name"),
		},
	},
		cfg)
	require.NoError(t, cfg.Validate())
}

func TestInvalidConfigs(t *testing.T) {
	tests := []struct{ name, expectedError string }{
		{name: "no_watch_observers", expectedError: "`watch_observers` must be defined and include at least one configured observer extension"},
		{name: "reserved_receiver_creator", expectedError: `receiver "receiver_creator/with-name" validation failure: receiver cannot be a receiver_creator`},
		{name: "reserved_receiver_name", expectedError: "receiver \"a_receiver/with-receiver_creator/in-name\" validation failure: receiver name cannot contain \"receiver_creator/\""},
	}

	for _, test := range tests {
		func(name, expectedError string) {
			t.Run(name, func(t *testing.T) {
				config, err := confmaptest.LoadConf(path.Join(".", "testdata", fmt.Sprintf("%s.yaml", name)))
				require.NoError(t, err)
				cm, err := config.Sub(typeStr)
				require.NoError(t, err)
				cfg := createDefaultConfig().(*Config)
				err = cm.Unmarshal(&cfg)
				require.NoError(t, err)
				err = cfg.Validate()
				require.Error(t, err)
				require.EqualError(t, err, expectedError)
			})
		}(test.name, test.expectedError)
	}
}

func TestReceiverCreatorFactoryAndConfig(t *testing.T) {
	conf, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	conf, err = conf.Sub("discovery")
	require.NoError(t, err)
	require.NotEmpty(t, conf.ToStringMap())
	dCfg := Config{}
	require.NoError(t, conf.Unmarshal(&dCfg))

	factory, rCfg, err := dCfg.receiverCreatorFactoryAndConfig()
	require.NoError(t, err)
	require.Equal(t, component.MustNewType("receiver_creator"), factory.Type())

	require.NoError(t, xconfmap.Validate(rCfg))

	creatorCfg, ok := rCfg.(*receivercreator.Config)
	require.True(t, ok)
	require.NotNil(t, creatorCfg)

	// receiver templates aren't exported so limited to WatchObservers
	require.Equal(t, []component.ID{
		component.MustNewID("an_observer"),
		component.MustNewIDWithName("another_observer", "with_name"),
	}, creatorCfg.WatchObservers)

	receiverTemplate := dCfg.receiverCreatorReceiversConfig()
	expectedTemplate := map[string]any{
		"smartagent/redis": map[string]any{
			"config": map[string]any{
				"auth": "password",
				"host": "`host`",
				"port": "`port`",
				"type": "collectd/redis",
			},
			"resource_attributes": map[string]string{
				"discovery.endpoint.id":   "`id`",
				"discovery.receiver.name": "redis",
				"discovery.receiver.type": "smartagent",
				"receiver_attribute":      "receiver_attribute_value",
			},
			"rule": `type == "container" && name matches "(?i)redis"`,
		},
	}
	require.Equal(t, expectedTemplate, receiverTemplate)
}
