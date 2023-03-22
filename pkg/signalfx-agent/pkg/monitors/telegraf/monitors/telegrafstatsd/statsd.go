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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false"`
	// Protocol to use with the listener: `tcp`, `udp4`, `udp6`, or `udp`.
	Protocol string `yaml:"protocol" default:"udp"`
	// The address and port to serve from
	ServiceAddress string `yaml:"serviceAddress" default:":8125"`
	// Maximum number of tcp connections allowed.
	MaxTCPConnections int `yaml:"maxTCPConnections" default:"250"`
	// Indicates whether to keep the tcp connection alive.
	TCPKeepAlive bool `yaml:"TCPKeepAlive" default:"false"`
	// Whether to clear the gauge cache every interval.  Setting this to false means the cache
	// will only be cleared when the monitor is restarted.
	DeleteGauges bool `yaml:"deleteGauges" default:"true"`
	// Whether to clear the counter cache every interval.  Setting this to false means the cache
	// will only be cleared when the monitor is restarted.
	DeleteCounters bool `yaml:"deleteCounters" default:"true"`
	// Whether to clear the sets cache every interval.  Setting this to false means the cache
	// will only be cleared when the monitor is restarted.
	DeleteSets bool `yaml:"deleteSets" default:"true"`
	// Whether to clear the timings cache every interval.  Setting this to false means the cache
	// will only be cleared when the monitor is restarted.
	DeleteTimings bool `yaml:"deleteTimings" default:"true"`
	// The percentiles that are collected for timing and histogram stats.
	Percentiles []int `yaml:"percentiles"`
	// Number of messages allowed to queue up between each collection interval.
	// Packets will be dropped until the next collection interval if this buffer
	// fills up.
	AllowedPendingMessages int `yaml:"allowedPendingMessages" default:"10000"`
	// The maximum number of histogram values to track each measurement when calculating percentiles.
	// Increasing the limit will increase memory consumption but will also improve accuracy.
	PercentileLimit int `yaml:"percentileLimit" default:"1000"`
	// The separator used to separate parts of a metric name
	MetricSeparator string `yaml:"metricSeparator" default:"_"`
	// Templates that transform telegrafstatsd metrics into influx tags and measurements.
	// Please refer to the Telegraf (documentation)[https://github.com/influxdata/telegraf/tree/master/plugins/inputs/statsd#statsd-bucket---influxdb-line-protocol-templates]
	// for more information on templates.
	Templates []string `yaml:"templates"`
	// Indicates whether to parse dogstatsd tags
	ParseDataDogTags bool `yaml:"parseDataDogTags" default:"false"`
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
