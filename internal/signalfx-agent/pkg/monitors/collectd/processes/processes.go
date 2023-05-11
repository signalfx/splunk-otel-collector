//go:build linux
// +build linux

package processes

//go:generate ../../../../scripts/collectd-template-to-go processes.tmpl

import (
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			MonitorCore: *collectd.NewMonitorCore(CollectdTemplate),
			logger:      logger,
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true"`
	// A list of process names to match
	Processes []string `yaml:"processes"`
	// A map with keys specifying the `plugin_instance` value to be sent for
	// the values which are regexes that match process names.  See example in
	// description.
	ProcessMatch map[string]string `yaml:"processMatch"`
	// Collect metrics on the number of context switches made by the process
	CollectContextSwitch bool `yaml:"collectContextSwitch" default:"false"`
	// (Deprecated) Please set the agent configuration `procPath` instead of
	// this monitor configuration option.
	// The path to the proc filesystem -- useful to override if the agent is
	// running in a container.
	ProcFSPath string `yaml:"procFSPath" default:""`
}

// Validate will check the config for correctness.
func (c *Config) Validate() error {
	return nil
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
	logger log.FieldLogger
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	m.logger = m.logger.WithField("monitorID", conf.MonitorID)
	if conf.ProcFSPath != "" {
		m.logger.Warningf("Please set the `procPath` top level agent configuration instead of the monitor level configuration")
	} else {
		// get top level configuration for /proc path
		conf.ProcFSPath = hostfs.HostProc()
	}
	return m.SetConfigurationAndRun(conf)
}
