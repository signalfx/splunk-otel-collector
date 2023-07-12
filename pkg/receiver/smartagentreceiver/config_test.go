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
	"github.com/signalfx/signalfx-agent/pkg/monitors/cadvisor"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/consul"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/hadoop"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {

	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 5, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, haproxyCfg))

	expectedDimensionClients := []string{"nop/one", "nop/two"}
	require.Equal(t, &Config{
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "redis").String())
	require.NoError(t, err)
	redisCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, redisCfg))

	require.Equal(t, &Config{
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "hadoop").String())
	require.NoError(t, err)
	hadoopCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, hadoopCfg))

	require.Equal(t, &Config{
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "etcd").String())
	require.NoError(t, err)
	etcdCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, etcdCfg))

	require.Equal(t, &Config{
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
	require.NoError(t, etcdCfg.validate())

	tr := true
	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "ntpq").String())
	require.NoError(t, err)
	ntpqCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, ntpqCfg))
	require.Equal(t, &Config{
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "without_type.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/withouttype")
	require.NoError(t, err)
	withoutType := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, withoutType)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`you must specify a "type" for a smartagent receiver`)
}

func TestLoadInvalidConfigWithUnknownType(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "unknown_type.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub("smartagent/unknowntype")
	require.NoError(t, err)
	unknown := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, unknown)
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
	err = component.UnmarshalConfig(cm, unexpected)
	require.Error(t, err)
	require.ErrorContains(t, err,
		"failed creating Smart Agent Monitor custom config: yaml: unmarshal errors:\n  line 2: field notASupportedTag not found in type redis.Config")
}

func TestLoadInvalidConfigs(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 2, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "negativeintervalseconds").String())
	require.NoError(t, err)
	negativeIntervalCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, negativeIntervalCfg))
	require.Equal(t, &Config{
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "missingrequired").String())
	require.NoError(t, err)
	missingRequiredCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, missingRequiredCfg))
	require.Equal(t, &Config{
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "endpoints_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 6, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, haproxyCfg))
	require.Equal(t, &Config{
		Endpoint: "[fe80::20c:29ff:fe59:9446]:2345",
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "redis").String())
	require.NoError(t, err)
	redisCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, redisCfg))
	require.Equal(t, &Config{
		Endpoint: "redishost",
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "hadoop").String())
	require.NoError(t, err)
	hadoopCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, hadoopCfg))
	require.Equal(t, &Config{
		Endpoint: "[::]:12345",
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "etcd").String())
	require.NoError(t, err)
	etcdCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, etcdCfg))
	require.Equal(t, &Config{
		Endpoint: "etcdhost:5555",
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
	require.NoError(t, etcdCfg.validate())

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "elasticsearch").String())
	require.NoError(t, err)
	elasticCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, elasticCfg))
	tru := true
	require.Equal(t, &Config{
		Endpoint: "elastic:567",
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
	require.NoError(t, elasticCfg.validate())

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "kubelet-stats").String())
	require.NoError(t, err)
	kubeletCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, kubeletCfg))
	require.Equal(t, &Config{
		Endpoint: "disregarded:678",
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
	require.NoError(t, kubeletCfg.validate())
}

func TestLoadInvalidConfigWithInvalidEndpoint(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_endpoint.yaml"))
	require.NoError(t, err)

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, haproxyCfg)
	require.ErrorContains(t, err,
		`cannot determine port via Endpoint: strconv.ParseUint: parsing "notaport": invalid syntax`)
}

// Even though this smart-agent monitor does not accept endpoints,
// we can create it without setting Host/Port fields.
func TestLoadConfigWithUnsupportedEndpoint(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "unsupported_endpoint.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "nagios").String())
	require.NoError(t, err)
	nagiosCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, nagiosCfg))

	require.Equal(t, &Config{
		Endpoint: "localhost:12345",
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_nonarray_dimension_clients.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, haproxyCfg)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`dimensionClients must be an array of compatible exporter names`)
}

func TestLoadInvalidConfigWithNonStringArrayDimensionClients(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_float_dimension_clients.yaml"))
	require.NoError(t, err)
	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "haproxy").String())
	require.NoError(t, err)
	haproxyCfg := CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, haproxyCfg)
	require.Error(t, err)
	require.ErrorContains(t, err,
		`dimensionClients must be an array of compatible exporter names`)
}

func TestFilteringConfig(t *testing.T) {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "filtering_config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "filesystems").String())
	require.NoError(t, err)
	fsCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, fsCfg))

	require.Equal(t, &Config{
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "invalid_filtering_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "filesystems").String())
	require.NoError(t, err)
	fsCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, fsCfg))
	require.Equal(t, &Config{
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
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "nested_monitor_config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 2, len(cfg.ToStringMap()))

	cm, err := cfg.Sub(component.NewIDWithName(typeStr, "exec").String())
	require.NoError(t, err)
	telegrafExecCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, telegrafExecCfg))
	require.Equal(t, &Config{
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

	cm, err = cfg.Sub(component.NewIDWithName(typeStr, "kubernetes_volumes").String())
	require.NoError(t, err)
	k8sVolumesCfg := CreateDefaultConfig().(*Config)
	require.NoError(t, component.UnmarshalConfig(cm, k8sVolumesCfg))
	tru := true
	require.Equal(t, &Config{
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
