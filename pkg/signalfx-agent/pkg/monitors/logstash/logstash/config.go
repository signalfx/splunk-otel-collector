package logstash

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true" singleInstance:"false"`
	// The hostname of Logstash monitoring API
	Host string `yaml:"host" default:"127.0.0.1"`
	// The port number of Logstash monitoring API
	Port uint16 `yaml:"port" default:"9600"`
	// If true, the agent will connect to the host using HTTPS instead of plain HTTP.
	UseHTTPS bool `yaml:"useHTTPS"`
	// The maximum amount of time to wait for API requests
	TimeoutSeconds int `yaml:"timeoutSeconds" default:"5"`
}

func (c *Config) getMetricTypeMap() map[string]datapoint.MetricType {
	metricTypeMap := make(map[string]datapoint.MetricType)

	for metricName := range defaultMetrics {
		metricTypeMap[metricName] = metricSet[metricName].Type
	}

	for _, groupName := range c.ExtraGroups {
		if m, exists := groupMetricsMap[groupName]; exists {
			for _, metricName := range m {
				metricTypeMap[metricName] = metricSet[metricName].Type
			}
		}
	}

	return metricTypeMap
}
