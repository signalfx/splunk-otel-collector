// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// +build linux

package smartagentreceiver

import (
	"path"
	"testing"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/apache"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/genericjmx"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/kafka"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memcached"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfigWithLinuxOnlyMonitors(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[config.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfig(
		path.Join(".", "testdata", "linux_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 4)

	apacheCfg := cfg.Receivers[config.NewIDWithName(typeStr, "apache")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewIDWithName(typeStr, "apache")),
		monitorConfig: &apache.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/apache",
				IntervalSeconds:     234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "localhost",
			Port: 6379,
			URL:  "http://{{.Host}}:{{.Port}}/mod_status?auto",
		},
	}, apacheCfg)
	require.NoError(t, apacheCfg.validate())

	kafkaCfg := cfg.Receivers[config.NewIDWithName(typeStr, "kafka")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewIDWithName(typeStr, "kafka")),
		monitorConfig: &kafka.Config{
			Config: genericjmx.Config{
				MonitorConfig: saconfig.MonitorConfig{
					Type:                "collectd/kafka",
					IntervalSeconds:     345,
					DatapointsToExclude: []saconfig.MetricFilter{},
				},
				Host:       "localhost",
				Port:       7199,
				ServiceURL: "service:jmx:rmi:///jndi/rmi://{{.Host}}:{{.Port}}/jmxrmi",
			},
			ClusterName: "somecluster",
		},
	}, kafkaCfg)
	require.NoError(t, kafkaCfg.validate())

	memcachedCfg := cfg.Receivers[config.NewIDWithName(typeStr, "memcached")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewIDWithName(typeStr, "memcached")),
		monitorConfig: &memcached.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/memcached",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "localhost",
			Port: 5309,
		},
	}, memcachedCfg)
	require.NoError(t, memcachedCfg.validate())

	phpCfg := cfg.Receivers[config.NewIDWithName(typeStr, "php")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewIDWithName(typeStr, "php")),
		monitorConfig: &php.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/php-fpm",
				IntervalSeconds:     0,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Path: "/status",
		},
	}, phpCfg)
	require.NoError(t, phpCfg.validate())
}
