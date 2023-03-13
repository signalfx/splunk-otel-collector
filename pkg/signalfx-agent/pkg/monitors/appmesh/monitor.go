package appmesh

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/statsd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" singleInstance:"false"`
	// The host/address on which to bind the UDP listener that accepts statsd
	// datagrams
	ListenAddress string `yaml:"listenAddress" default:"localhost"`
	// The port on which to listen for statsd messages (**default:** `8125`)
	ListenPort *uint16 `yaml:"listenPort"`
	// A prefix in metric names that needs to be removed before metric name conversion
	MetricPrefix string `yaml:"metricPrefix"`
}

// Monitor that listens to incoming statsd metrics and converts the metrics in AWS AppMesh metric format
type Monitor struct {
	Output  types.Output
	monitor *statsd.Monitor
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	// Give default value to ListenPort if not given by user.
	// Cannot use yaml default to take also 0 as a valid value.
	if conf.ListenPort == nil {
		conf.ListenPort = new(uint16)
		*conf.ListenPort = 8125
	}

	var err error
	m.monitor, err = m.statsDMonitor(conf)

	if err != nil {
		return fmt.Errorf("could not start StatsD monitor: %v", err)
	}

	return nil
}

func (m *Monitor) statsDMonitor(conf *Config) (*statsd.Monitor, error) {
	monitor := &statsd.Monitor{Output: m.Output.Copy()}

	return monitor, monitor.Configure(&statsd.Config{
		MonitorConfig: conf.MonitorConfig,
		ListenAddress: conf.ListenAddress,
		ListenPort:    conf.ListenPort,
		MetricPrefix:  conf.MetricPrefix,
		Converters: []statsd.ConverterInput{
			{
				Pattern:    "cluster.cds_{traffic}_{mesh}_{service}-vn_{}.{action}",
				MetricName: "{action}",
			},
		},
	})
}

// Shutdown shuts down the internal StatsD monitor
func (m *Monitor) Shutdown() {
	if m.monitor != nil {
		m.monitor.Shutdown()
	}
}
