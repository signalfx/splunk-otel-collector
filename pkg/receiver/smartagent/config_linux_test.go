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

//go:build linux

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
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfigWithLinuxOnlyMonitors(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "linux_config.yaml"))
	require.NoError(t, err)

	assert.Equal(t, 4, len(configs.ToStringMap()))

	cm, err := configs.Sub(component.NewIDWithName(typeStr, "apache").String())
	require.NoError(t, err)
	apacheCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, apacheCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
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
		acceptsEndpoints: true,
	}, apacheCfg)
	require.NoError(t, apacheCfg.validate())

	cm, err = configs.Sub(component.NewIDWithName(typeStr, "kafka").String())
	require.NoError(t, err)
	kafkaCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, kafkaCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
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
		acceptsEndpoints: true,
	}, kafkaCfg)
	require.NoError(t, kafkaCfg.validate())

	cm, err = configs.Sub(component.NewIDWithName(typeStr, "memcached").String())
	require.NoError(t, err)
	memcachedCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, memcachedCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		monitorConfig: &memcached.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/memcached",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "localhost",
			Port: 5309,
		},
		acceptsEndpoints: true,
	}, memcachedCfg)
	require.NoError(t, memcachedCfg.validate())

	cm, err = configs.Sub(component.NewIDWithName(typeStr, "php").String())
	require.NoError(t, err)
	phpCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, phpCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		monitorConfig: &php.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/php-fpm",
				IntervalSeconds:     0,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Path: "/status",
		},
		acceptsEndpoints: true,
	}, phpCfg)
	require.NoError(t, phpCfg.validate())
}
