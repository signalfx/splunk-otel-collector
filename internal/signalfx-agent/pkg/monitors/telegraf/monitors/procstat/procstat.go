package procstat

import (
	"context"
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
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false"`
	Exe                  string `yaml:"exe"`
	Pattern              string `yaml:"pattern"`
	User                 string `yaml:"user"`
	PidFile              string `yaml:"pidFile"`
	ProcessName          string `yaml:"processName"`
	Prefix               string `yaml:"prefix"`
	CGroup               string `yaml:"cGroup"`
	WinService           string `yaml:"WinService"`
	PidTag               bool   `yaml:"pidTag"`
	CmdLineTag           bool   `yaml:"cmdLineTag"`
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
		conf.CGroup = path.Join(hostfs.HostSys(), "fs", "cgroup", conf.CGroup)
	}

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	em.AddTag("plugin", strings.ReplaceAll(monitorType, "/", "-"))

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
