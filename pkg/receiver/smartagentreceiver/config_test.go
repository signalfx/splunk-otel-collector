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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/cadvisor"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	"github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/stats"
	"github.com/signalfx/signalfx-agent/pkg/monitors/filesystems"
	"github.com/signalfx/signalfx-agent/pkg/monitors/haproxy"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/volumes"
	"github.com/signalfx/signalfx-agent/pkg/monitors/nagios"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/parser"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/exec"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/ntpq"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

type fakeConfig struct {
	EnhancedMetrics        *bool `yaml:"enhancedMetrics"`
	saconfig.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig    `yaml:",inline"`
	Host                   string `yaml:"host" validate:"required"`
	Port                   uint16 `yaml:"port" validate:"required"`
}

func init() {
	monitors.Register(&monitors.Metadata{
		MonitorType: "collectd/fake",
	},
		nil,
		&fakeConfig{},
	)
}

func TestLoadConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 4, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&haproxyCfg))

	expectedDimensionClients := []string{"nop/one", "nop/two"}
	require.Equal(t, &Config{
		MonitorType:      "haproxy",
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
	require.NoError(t, haproxyCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "redis").String())
	require.NoError(t, err)
	redisCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&redisCfg))

	require.Equal(t, &Config{
		MonitorType:      "collectd/redis",
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
	require.NoError(t, redisCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "etcd").String())
	require.NoError(t, err)
	etcdCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&etcdCfg))

	require.Equal(t, &Config{
		MonitorType: "etcd",
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:                  "localhost",
			Port:                  5309,
			MetricPath:            "/metrics",
			ScrapeFailureLogLevel: "error",
		},
		acceptsEndpoints: true,
	}, etcdCfg)
	require.NoError(t, etcdCfg.Validate())

	tr := true
	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "ntpq").String())
	require.NoError(t, err)
	ntpqCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&ntpqCfg))
	require.Equal(t, &Config{
		MonitorType: "telegraf/ntpq",
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
	require.NoError(t, ntpqCfg.Validate())
}

func TestLoadInvalidConfigWithoutType(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "without_type.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/withouttype")
	require.NoError(t, err)
	withoutType := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&withoutType)
	require.NoError(t, err)
	require.Nil(t, withoutType)
}

func TestLoadInvalidConfigWithUnknownType(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "unknown_type.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/unknowntype")
	require.NoError(t, err)
	unknown := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&unknown)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`no known monitor type "notamonitor"`)
}

func TestLoadInvalidConfigWithUnexpectedTag(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "unexpected_tag.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/unexpectedtag")
	require.NoError(t, err)
	unexpected := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&unexpected)
	require.Error(t, err)
	require.ErrorContains(t, err,
		"failed creating Smart Agent Monitor custom config: yaml: unmarshal errors:\n  line 2: field notASupportedTag not found in type redis.Config")
}

func TestLoadInvalidConfigs(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 2, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "negativeintervalseconds").String())
	require.NoError(t, err)
	negativeIntervalCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&negativeIntervalCfg))
	require.Equal(t, &Config{
		MonitorType: "collectd/redis",
		monitorConfig: &redis.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     -234,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
		},
		acceptsEndpoints: true,
	}, negativeIntervalCfg)
	err = negativeIntervalCfg.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "intervalSeconds must be greater than 0s (-234 provided)")

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "missingrequired").String())
	require.NoError(t, err)
	missingRequiredCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&missingRequiredCfg))
	require.Equal(t, &Config{
		MonitorType: "collectd/fake",
		monitorConfig: &fakeConfig{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "collectd/fake",
				IntervalSeconds:     0,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			Port: 5309,
		},
		acceptsEndpoints: true,
	}, missingRequiredCfg)
	err = missingRequiredCfg.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "Validation error in field 'fakeConfig.host': host is a required field (got '')")
}

func TestLoadConfigWithEndpoints(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "endpoints_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 5, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&haproxyCfg))
	require.Equal(t, &Config{
		MonitorType: "haproxy",
		Endpoint:    "[fe80::20c:29ff:fe59:9446]:2345",
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
	require.NoError(t, haproxyCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "redis").String())
	require.NoError(t, err)
	redisCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&redisCfg))
	require.Equal(t, &Config{
		MonitorType: "collectd/redis",
		Endpoint:    "redishost",
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
	require.NoError(t, redisCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "etcd").String())
	require.NoError(t, err)
	etcdCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&etcdCfg))
	require.Equal(t, &Config{
		MonitorType: "etcd",
		Endpoint:    "etcdhost:5555",
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:                  "localhost",
			Port:                  5555,
			MetricPath:            "/metrics",
			ScrapeFailureLogLevel: "error",
		},
		acceptsEndpoints: true,
	}, etcdCfg)
	require.NoError(t, etcdCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "elasticsearch").String())
	require.NoError(t, err)
	elasticCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&elasticCfg))
	tru := true
	require.Equal(t, &Config{
		MonitorType: "elasticsearch",
		Endpoint:    "elastic:567",
		monitorConfig: &stats.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "elasticsearch",
				IntervalSeconds:     567,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:                           "elastic",
			Port:                           "567",
			EnableIndexStats:               &tru,
			IndexStatsIntervalSeconds:      60,
			IndexStatsMasterOnly:           &tru,
			EnableClusterHealth:            &tru,
			ClusterHealthStatsMasterOnly:   &tru,
			ThreadPools:                    []string{"search", "index"},
			MetadataRefreshIntervalSeconds: 30,
		},
		acceptsEndpoints: true,
	}, elasticCfg)
	require.NoError(t, elasticCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "kubelet-stats").String())
	require.NoError(t, err)
	kubeletCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&kubeletCfg))
	require.Equal(t, &Config{
		MonitorType: "kubelet-stats",
		Endpoint:    "disregarded:678",
		monitorConfig: &cadvisor.KubeletStatsConfig{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "kubelet-stats",
				IntervalSeconds:     789,
				DatapointsToExclude: []saconfig.MetricFilter{},
			},
			KubeletAPI: kubelet.APIConfig{
				AuthType:   "serviceAccount",
				SkipVerify: &tru,
			},
		},
		acceptsEndpoints: true,
	}, kubeletCfg)
	require.NoError(t, kubeletCfg.Validate())
}

func TestLoadInvalidConfigWithInvalidEndpoint(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_endpoint.yaml"))
	require.NoError(t, err)

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&haproxyCfg)
	require.ErrorContains(t, err,
		`cannot determine port via Endpoint: strconv.ParseUint: parsing "notaport": invalid syntax`)
}

// Even though this smart-agent monitor does not accept endpoints,
// we can create it without setting Host/Port fields.
func TestLoadConfigWithUnsupportedEndpoint(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "unsupported_endpoint.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "nagios").String())
	require.NoError(t, err)
	nagiosCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&nagiosCfg))

	require.Equal(t, &Config{
		MonitorType: "nagios",
		Endpoint:    "localhost:12345",
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
	require.NoError(t, nagiosCfg.Validate())
}

func TestLoadInvalidConfigWithNonArrayDimensionClients(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_nonarray_dimension_clients.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&haproxyCfg)
	require.NoError(t, err)
	require.Equal(t, &Config{
		MonitorType:      "haproxy",
		DimensionClients: []string{"notanarray"},
		monitorConfig: &haproxy.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:                "haproxy",
				DatapointsToExclude: []saconfig.MetricFilter{},
				IntervalSeconds:     123,
			},
			Username:  "SomeUser",
			Password:  "secret",
			Path:      "stats?stats;csv",
			SSLVerify: true,
			Timeout:   5000000000,
		},
		acceptsEndpoints: true,
	}, haproxyCfg)
}

func TestLoadInvalidConfigWithNonStringArrayDimensionClients(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_float_dimension_clients.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = cm.Unmarshal(&haproxyCfg)
	require.Error(t, err)
	require.ErrorContains(t, err, `expected type 'string'`)
}

func TestFilteringConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "filtering_config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "filesystems").String())
	require.NoError(t, err)
	fsCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&fsCfg))

	require.Equal(t, &Config{
		MonitorType: "filesystems",
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
	require.NoError(t, fsCfg.Validate())
}

func TestInvalidFilteringConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_filtering_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "filesystems").String())
	require.NoError(t, err)
	fsCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&fsCfg))
	require.Equal(t, &Config{
		MonitorType: "filesystems",
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

	err = fsCfg.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "unexpected end of input")
}

func TestLoadConfigWithNestedMonitorConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "nested_monitor_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 2, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.MustNewIDWithName(typeStr, "exec").String())
	require.NoError(t, err)
	telegrafExecCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&telegrafExecCfg))
	require.Equal(t, &Config{
		MonitorType: "telegraf/exec",
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
	require.NoError(t, telegrafExecCfg.Validate())

	cm, err = cfg.Sub(component.MustNewIDWithName(typeStr, "kubernetes_volumes").String())
	require.NoError(t, err)
	k8sVolumesCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&k8sVolumesCfg))
	tru := true
	require.Equal(t, &Config{
		MonitorType: "kubernetes-volumes",
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
	require.NoError(t, k8sVolumesCfg.Validate())
}

func TestInvalidMonitorConfig(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", -123)
	assert.EqualError(t, cfg.Validate(), "intervalSeconds must be greater than 0s (-123 provided)")
}
