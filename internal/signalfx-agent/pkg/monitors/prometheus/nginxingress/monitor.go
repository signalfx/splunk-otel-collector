package nginxingress

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{Monitor: prometheusexporter.Monitor{}}
	}, &prometheusexporter.Config{})
}

// Monitor for Ingress Nginx
type Monitor struct {
	prometheusexporter.Monitor
}

// Configure the underlying Prometheus exporter monitor
func (m *Monitor) Configure(conf *prometheusexporter.Config) error {
	return m.Monitor.Configure(conf)
}
