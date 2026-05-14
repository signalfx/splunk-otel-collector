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
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memcached"
)

func TestLoadConfigWithLinuxOnlyMonitors(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "linux_config.yaml"))
	require.NoError(t, err)

	assert.Equal(t, 1, len(configs.ToStringMap()))

	cm, err := configs.Sub(component.MustNewIDWithName(typeStr, "memcached").String())
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
}
