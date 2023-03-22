package node

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	prometheusexporter.Config `yaml:",inline"`
}

// Monitor for Prometheus Node Exporter
type Monitor struct {
	prometheusexporter.Monitor
}

// Configure the underlying Prometheus exporter monitor
func (m *Monitor) Configure(conf *Config) error {
	return m.Monitor.Configure(&conf.Config)
}
