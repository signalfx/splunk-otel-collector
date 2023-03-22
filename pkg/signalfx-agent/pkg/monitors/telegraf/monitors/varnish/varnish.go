//go:build linux
// +build linux

package varnish

import (
	"context"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/varnish"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"true"`
	// If running as a restricted user enable this flag to prepend sudo.
	UseSudo bool `yaml:"useSudo" default:"false"`
	// The location of the varnishstat binary.
	Binary string `yaml:"binary" default:"/usr/bin/varnishstat"`
	// Which stats to gather. Glob matching can be used (i.e. `stats = ["MAIN.*"]`).
	//Stats []string `yaml:"stats" default:"[\"MAIN.cache_hit\", \"MAIN.cache_miss\", \"MAIN.uptime\"]"`
	Stats []string `yaml:"stats" default:"[\"MAIN.*\"]"`
	// Optional name for the varnish instance to query. It corresponds to `-n` parameter value.
	InstanceName string `yaml:"instanceName"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	logger log.FieldLogger
}

type Emitter struct {
	*baseemitter.BaseEmitter
}

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	plugin := telegrafInputs.Inputs["varnish"]().(*telegrafPlugin.Varnish)
	plugin.UseSudo = conf.UseSudo
	plugin.Binary = conf.Binary
	plugin.Stats = conf.Stats
	plugin.InstanceName = conf.InstanceName

	emitter := &Emitter{baseemitter.NewEmitter(m.Output, m.logger)}

	// don't include the telegraf_type dimension
	emitter.SetOmitOriginalMetricType(true)

	emitter.AddTag("plugin", monitorType)

	accumulator := accumulator.NewAccumulator(emitter)

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(accumulator); err != nil {
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
