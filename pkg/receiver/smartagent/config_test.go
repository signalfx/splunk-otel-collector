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

package smartagentreceiver

import (
	"path"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/consul"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/hadoop"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	"github.com/signalfx/signalfx-agent/pkg/monitors/filesystems"
	"github.com/signalfx/signalfx-agent/pkg/monitors/haproxy"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/volumes"
	"github.com/signalfx/signalfx-agent/pkg/monitors/nagios"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/parser"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/exec"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/ntpq"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 5)

	haproxyCfg := cfg.Receivers[component.NewIDWithName(typeStr, "haproxy")].(*Config)
	expectedDimensionClients := []string{"nop/one", "nop/two"}
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "haproxy")),
		DimensionClients: expectedDimensionClients,
		monitorConfig: &haproxy.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "haproxy",
				IntervalSeconds:     123,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Username:  "SomeUser",
			Password:  "secret",
			Path:      "stats?stats;csv",
			SSLVerify: true,
			Timeout:   timeutil.Duration(5 * time.Second),
		},
		acceptsEndpoints: true,
	}, haproxyCfg)
	require.NoError(t, haproxyCfg.validate())

	redisCfg := cfg.Receivers[component.NewIDWithName(typeStr, "redis")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "redis")),
		DimensionClients: []string{},
		monitorConfig: &redis.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "localhost",
			Port: 6379,
		},
		acceptsEndpoints: true,
	}, redisCfg)
	require.NoError(t, redisCfg.validate())

	hadoopCfg := cfg.Receivers[component.NewIDWithName(typeStr, "hadoop")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "hadoop")),
		monitorConfig: &hadoop.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/hadoop",
				IntervalSeconds:     345,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			CommonConfig: python.CommonConfig{},
			Host:         "localhost",
			Port:         8088,
		},
		acceptsEndpoints: true,
	}, hadoopCfg)
	require.NoError(t, hadoopCfg.validate())

	etcdCfg := cfg.Receivers[component.NewIDWithName(typeStr, "etcd")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "etcd")),
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:       "localhost",
			Port:       5309,
			MetricPath: "/metrics",
		},
		acceptsEndpoints: true,
	}, etcdCfg)
	require.NoError(t, etcdCfg.validate())

	tr := true
	ntpqCfg := cfg.Receivers[component.NewIDWithName(typeStr, "ntpq")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "ntpq")),
		monitorConfig: &ntpq.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "telegraf/ntpq",
				IntervalSeconds:     567,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			DNSLookup: &tr,
		},
		acceptsEndpoints: true,
	}, ntpqCfg)
	require.NoError(t, ntpqCfg.validate())
}

func TestLoadInvalidConfigWithoutType(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "without_type.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`error reading receivers configuration for "smartagent/withouttype": you must specify a "type" for a smartagent receiver`)
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithUnknownType(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "unknown_type.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`error reading receivers configuration for "smartagent/unknowntype": no known monitor type "notamonitor"`)
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithUnexpectedTag(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "unexpected_tag.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		"error reading receivers configuration for \"smartagent/unexpectedtag\": failed creating Smart Agent Monitor custom config: yaml: unmarshal errors:\n  line 2: field notASupportedTag not found in type redis.Config")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigs(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "invalid_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Receivers), 2)

	negativeIntervalCfg := cfg.Receivers[component.NewIDWithName(typeStr, "negativeintervalseconds")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "negativeintervalseconds")),
		monitorConfig: &redis.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     -234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
		},
		acceptsEndpoints: true,
	}, negativeIntervalCfg)
	err = negativeIntervalCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "intervalSeconds must be greater than 0s (-234 provided)")

	missingRequiredCfg := cfg.Receivers[component.NewIDWithName(typeStr, "missingrequired")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "missingrequired")),
		monitorConfig: &consul.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/consul",
				IntervalSeconds:     0,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Port:          5309,
			TelemetryHost: "0.0.0.0",
			TelemetryPort: 8125,
		},
		acceptsEndpoints: true,
	}, missingRequiredCfg)
	err = missingRequiredCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "Validation error in field 'Config.host': host is a required field (got '')")
}

func TestLoadConfigWithEndpoints(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "endpoints_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 4)

	haproxyCfg := cfg.Receivers[component.NewIDWithName(typeStr, "haproxy")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "haproxy")),
		Endpoint:         "[fe80::20c:29ff:fe59:9446]:2345",
		monitorConfig: &haproxy.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "haproxy",
				IntervalSeconds:     123,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host:      "fe80::20c:29ff:fe59:9446",
			Port:      2345,
			Username:  "SomeUser",
			Password:  "secret",
			Path:      "stats?stats;csv",
			SSLVerify: true,
			Timeout:   timeutil.Duration(5 * time.Second),
		},
		acceptsEndpoints: true,
	}, haproxyCfg)
	require.NoError(t, haproxyCfg.validate())

	redisCfg := cfg.Receivers[component.NewIDWithName(typeStr, "redis")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "redis")),
		Endpoint:         "redishost",
		monitorConfig: &redis.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Host: "redishost",
			Port: 6379,
		},
		acceptsEndpoints: true,
	}, redisCfg)
	require.NoError(t, redisCfg.validate())

	hadoopCfg := cfg.Receivers[component.NewIDWithName(typeStr, "hadoop")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "hadoop")),
		Endpoint:         "[::]:12345",
		monitorConfig: &hadoop.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/hadoop",
				IntervalSeconds:     345,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			CommonConfig: python.CommonConfig{},
			Host:         "localhost",
			Port:         8088,
		},
		acceptsEndpoints: true,
	}, hadoopCfg)
	require.NoError(t, hadoopCfg.validate())

	etcdCfg := cfg.Receivers[component.NewIDWithName(typeStr, "etcd")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "etcd")),
		Endpoint:         "etcdhost:5555",
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:       "localhost",
			Port:       5555,
			MetricPath: "/metrics",
		},
		acceptsEndpoints: true,
	}, etcdCfg)
	require.NoError(t, etcdCfg.validate())
}

func TestLoadInvalidConfigWithInvalidEndpoint(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "invalid_endpoint.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`error reading receivers configuration for "smartagent/haproxy": cannot determine port via Endpoint: strconv.ParseUint: parsing "notaport": invalid syntax`)
	require.Nil(t, cfg)
}

// Even though this smart-agent monitor does not accept endpoints,
// we can create it without setting Host/Port fields.
func TestLoadConfigWithUnsupportedEndpoint(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "unsupported_endpoint.yaml"), factories,
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	nagiosCfg := cfg.Receivers[component.NewIDWithName(typeStr, "nagios")].(*Config)
	require.Equal(t, &Config{
		Endpoint:         "localhost:12345",
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "nagios")),
		monitorConfig: &nagios.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "nagios",
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Command: "some_command",
			Service: "some_service",
			Timeout: 9,
		},
		acceptsEndpoints: false,
	}, nagiosCfg)
	require.NoError(t, nagiosCfg.validate())
}

func TestLoadInvalidConfigWithNonArrayDimensionClients(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "invalid_nonarray_dimension_clients.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`error reading receivers configuration for "smartagent/haproxy": dimensionClients must be an array of compatible exporter names`)
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithNonStringArrayDimensionClients(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "invalid_float_dimension_clients.yaml"), factories,
	)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`error reading receivers configuration for "smartagent/haproxy": dimensionClients must be an array of compatible exporter names`)
	require.Nil(t, cfg)
}

func TestFilteringConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "filtering_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	fsCfg := cfg.Receivers[component.NewIDWithName(typeStr, "filesystems")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "filesystems")),
		monitorConfig: &filesystems.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type: "filesystems",
				DatapointsToExclude: []saconfig.MetricFilter{
					{
						MetricName: "df_inodes.*",
						Dimensions: map[string]any{
							"mountpoint": []any{"*", "!/hostfs/var/lib/cni"},
						},
					},
				},
				ExtraGroups:  []string{"inodes"},
				ExtraMetrics: []string{"percent_bytes.reserved"},
			},
		},
	}, fsCfg)
	require.NoError(t, fsCfg.validate())
}

func TestInvalidFilteringConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "invalid_filtering_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	fsCfg := cfg.Receivers[component.NewIDWithName(typeStr, "filesystems")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "filesystems")),
		monitorConfig: &filesystems.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type: "filesystems",
				DatapointsToExclude: []saconfig.MetricFilter{
					{
						MetricNames: []string{"./[0-"},
					},
				},
			},
		},
	}, fsCfg)

	err = fsCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "unexpected end of input")
}

func TestLoadConfigWithNestedMonitorConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfig(
		path.Join(".", "testdata", "nested_monitor_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 2)

	telegrafExecCfg := cfg.Receivers[component.NewIDWithName(typeStr, "exec")].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "exec")),
		monitorConfig: &exec.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "telegraf/exec",
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Commands: []string{
				`powershell.exe -Command "\Monitoring\Get_Directory.ps1"`,
			},
			TelegrafParser: &parser.Config{
				DataFormat: "influx",
			},
			Timeout: timeutil.Duration(5 * time.Second),
		},
	}, telegrafExecCfg)
	require.NoError(t, telegrafExecCfg.validate())

	k8sVolumesCfg := cfg.Receivers[component.NewIDWithName(typeStr, "kubernetes_volumes")].(*Config)
	tru := true
	require.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(component.NewIDWithName(typeStr, "kubernetes_volumes")),
		monitorConfig: &volumes.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "kubernetes-volumes",
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			KubeletAPI: kubelet.APIConfig{
				URL:        "https://192.168.99.103:10250",
				AuthType:   "serviceAccount",
				SkipVerify: &tru,
			},
			KubernetesAPI: &kubernetes.APIConfig{
				AuthType:   "serviceAccount",
				SkipVerify: false,
			},
		},
	}, k8sVolumesCfg)
	require.NoError(t, k8sVolumesCfg.validate())
}
