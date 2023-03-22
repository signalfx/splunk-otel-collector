package elasticsearch

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			python.PyMonitor{
				MonitorCore: subproc.New(),
			},
		}
	}, &Config{})
}

var _ config.ExtraMetrics = &Config{}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	// AdditionalMetrics to report on
	AdditionalMetrics []string `yaml:"additionalMetrics"`
	// Cluster name to which the node belongs. This is an optional config that
	// will override the cluster name fetched from a node and will be used to
	// populate the plugin_instance dimension
	Cluster string `yaml:"cluster"`
	// DetailedMetrics turns on additional metric time series
	DetailedMetrics *bool `yaml:"detailedMetrics" default:"true"`
	// EnableClusterHealth enables reporting on the cluster health
	EnableClusterHealth *bool `yaml:"enableClusterHealth" default:"true"`
	// EnableIndexStats reports metrics about indexes
	EnableIndexStats *bool `yaml:"enableIndexStats" default:"true"`
	// Indexes to report on
	Indexes []string `yaml:"indexes" default:"[\"_all\"]"`
	// IndexInterval is an interval in seconds at which the plugin will report index stats.
	// It must be greater than or equal, and divisible by the Interval configuration
	IndexInterval *uint `yaml:"indexInterval" default:"300"`
	// IndexStatsMasterOnly sends index stats from the master only
	IndexStatsMasterOnly *bool `yaml:"indexStatsMasterOnly" default:"false"`
	IndexSummaryOnly     *bool `yaml:"indexSummaryOnly" default:"false"`
	// Password used to access elasticsearch stats api
	Password string `yaml:"password" neverLog:"true"`
	// Protocol used to connect: http or https
	Protocol string `yaml:"protocol"`
	// ThreadPools to report on
	ThreadPools []string `yaml:"threadPools" default:"[\"search\", \"index\"]"`
	// Username used to access elasticsearch stats api
	Username string `yaml:"username"`
	Version  string `yaml:"version"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	m.Logger().Warn("The collectd/elasticsearch monitor is deprecated in favor of the elasticsearch monitor.")
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "elasticsearch_collectd",
		ModulePaths:   []string{collectd.MakePythonPluginPath("elasticsearch")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"Host":                 conf.Host,
			"Port":                 conf.Port,
			"Cluster":              conf.Cluster,
			"DetailedMetrics":      conf.DetailedMetrics,
			"EnableClusterHealth":  conf.EnableClusterHealth,
			"EnableIndexStats":     conf.EnableIndexStats,
			"IndexInterval":        conf.IndexInterval,
			"IndexStatsMasterOnly": conf.IndexStatsMasterOnly,
			"IndexSummaryOnly":     conf.IndexSummaryOnly,
			"Interval":             conf.IntervalSeconds,
			"Verbose":              false,
			"AdditionalMetrics":    conf.AdditionalMetrics,
			"Indexes":              conf.Indexes,
			"Password":             conf.Password,
			"Protocol":             conf.Protocol,
			"Username":             conf.Username,
			"ThreadPools":          conf.ThreadPools,
			"Version":              conf.Version,
		},
	}

	return m.PyMonitor.Configure(conf)
}

// GetExtraMetrics returns additional metrics that should be allowed through.
func (c *Config) GetExtraMetrics() []string {
	var extraMetrics []string

	for _, metric := range c.AdditionalMetrics {
		counterType := fmt.Sprintf("counter.%s", metric)
		gaugeType := fmt.Sprintf("gauge.%s", metric)

		// AdditionalMetrics doesn't specify the full metric name but it's either
		// a counter or a gauge so just check both.
		if monitorMetadata.HasMetric(counterType) {
			extraMetrics = append(extraMetrics, counterType)
			continue
		}

		if monitorMetadata.HasMetric(gaugeType) {
			extraMetrics = append(extraMetrics, gaugeType)
			continue
		}

		// We don't know about the metric so just enable both.
		extraMetrics = append(extraMetrics, counterType)
		extraMetrics = append(extraMetrics, gaugeType)
	}

	return extraMetrics
}
