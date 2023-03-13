package procstat

import (
	"context"
	"os"
	"path"
	"strings"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/procstat"
	log "github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/gopsutilhelper"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false"`
	// The name of an executable to monitor.  (ie: `exe: "signalfx-agent*"`)
	Exe string `yaml:"exe"`
	// Regular expression pattern to match against.
	Pattern string `yaml:"pattern"`
	// Username to match against
	User string `yaml:"user"`
	// Path to Pid file to monitor.  (ie: `pidFile: "/var/run/signalfx-agent.pid"`)
	PidFile string `yaml:"pidFile"`
	// Used to override the process name dimension
	ProcessName string `yaml:"processName"`
	// Prefix to be added to each dimension
	Prefix string `yaml:"prefix"`
	// Whether to add PID as a dimension instead of part of the metric name
	PidTag bool `yaml:"pidTag"`
	// When true add the full cmdline as a dimension.
	CmdLineTag bool `yaml:"cmdLineTag"`
	// The name of the cgroup to monitor.  This cgroup name will be appended to
	// the configured `sysPath`.  See the agent config schema for more information
	// about the `sysPath` agent configuration.
	CGroup string `yaml:"cGroup"`
	// The name of a windows service to report procstat information on.
	WinService string `yaml:"WinService"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	logger log.FieldLogger
}

// fetch the factory used to generate the perf counter plugin
var factory = telegrafInputs.Inputs["procstat"]

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	plugin := factory().(*telegrafPlugin.Procstat)

	// create the emitter
	em := baseemitter.NewEmitter(m.Output, m.logger)

	// use the agent's configured host sys to get cgroup information
	if conf.CGroup != "" {
		conf.CGroup = path.Join(os.Getenv(gopsutilhelper.HostSys), "fs", "cgroup", conf.CGroup)
	}

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	em.AddTag("plugin", strings.Replace(monitorType, "/", "-", -1))

	// create the accumulator
	ac := accumulator.NewAccumulator(em)

	// copy configurations to the plugin
	if err = deepcopier.Copy(conf).To(plugin); err != nil {
		m.logger.Error("unable to copy configurations to plugin")
		return err
	}

	// set the pid finder to native because we don't bundle pgrep at the moment
	// and containerizing pgrep is likely difficult
	plugin.PidFinder = "native"

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return err
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
