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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/apache"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memcached"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/php"
)

func TestLoadConfigWithLinuxOnlyMonitors(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "linux_config.yaml"))
	require.NoError(t, err)

	assert.Equal(t, 3, len(configs.ToStringMap()))

	cm, err := configs.Sub(component.MustNewIDWithName(typeStr, "apache").String())
	require.NoError(t, err)
	apacheCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&apacheCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		MonitorType: "collectd/apache",
		monitorConfig: &apache.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/apache",
				IntervalSeconds:     234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "localhost",
			Port: 6379,
			URL:  "http://{{.Host}}:{{.Port}}/server-status?auto",
		},
		acceptsEndpoints: true,
	}, apacheCfg)
	require.NoError(t, apacheCfg.Validate())

	cm, err = configs.Sub(component.MustNewIDWithName(typeStr, "memcached").String())
	require.NoError(t, err)
	memcachedCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&memcachedCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		MonitorType: "collectd/memcached",
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
	require.NoError(t, memcachedCfg.Validate())

	cm, err = configs.Sub(component.MustNewIDWithName(typeStr, "php").String())
	require.NoError(t, err)
	phpCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&phpCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		MonitorType: "collectd/php-fpm",
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
	require.NoError(t, phpCfg.Validate())
}
