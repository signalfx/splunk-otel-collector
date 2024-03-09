package telegrafstatsd

import (
	"context"
	"time"

	"github.com/ulule/deepcopier"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/statsd"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig   `yaml:",inline" acceptsEndpoints:"false"`
	Protocol               string   `yaml:"protocol" default:"udp"`
	ServiceAddress         string   `yaml:"serviceAddress" default:":8125"`
	MetricSeparator        string   `yaml:"metricSeparator" default:"_"`
	Percentiles            []int    `yaml:"percentiles"`
	Templates              []string `yaml:"templates"`
	MaxTCPConnections      int      `yaml:"maxTCPConnections" default:"250"`
	AllowedPendingMessages int      `yaml:"allowedPendingMessages" default:"10000"`
	PercentileLimit        int      `yaml:"percentileLimit" default:"1000"`
	DeleteSets             bool     `yaml:"deleteSets" default:"true"`
	DeleteTimings          bool     `yaml:"deleteTimings" default:"true"`
	DeleteCounters         bool     `yaml:"deleteCounters" default:"true"`
	DeleteGauges           bool     `yaml:"deleteGauges" default:"true"`
	TCPKeepAlive           bool     `yaml:"TCPKeepAlive" default:"false"`
	ParseDataDogTags       bool     `yaml:"parseDataDogTags" default:"false"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	plugin *telegrafPlugin.Statsd
	logger log.FieldLogger
}

// fetch the factory used to generate the perf counter plugin
var factory = telegrafInputs.Inputs["statsd"]

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	m.plugin = factory().(*telegrafPlugin.Statsd)

	// copy configurations to the plugin
	if err = deepcopier.Copy(conf).To(m.plugin); err != nil {
		m.logger.Error("unable to copy configurations to plugin")
		return err
	}

	// create the accumulator
	ac := accumulator.NewAccumulator(baseemitter.NewEmitter(m.Output, m.logger))

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// start the plugin
	if err = m.plugin.Start(ac); err != nil {
		return err
	}

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := m.plugin.Gather(ac); err != nil {
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
	if m.plugin != nil {
		m.plugin.Stop()
	}
}
