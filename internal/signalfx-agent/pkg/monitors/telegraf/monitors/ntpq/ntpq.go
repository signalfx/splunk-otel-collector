package ntpq

import (
	"context"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/ntpq"
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
	// If false, set the -n ntpq flag. Can reduce metric gather time.
	DNSLookup *bool `yaml:"dnsLookup" default:"true"`
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

	plugin := telegrafInputs.Inputs["ntpq"]().(*telegrafPlugin.NTPQ)
	plugin.DNSLookup = *conf.DNSLookup

	emitter := &Emitter{baseemitter.NewEmitter(m.Output, m.logger)}

	// don't include the telegraf_type dimension
	emitter.SetOmitOriginalMetricType(true)

	emitter.AddTag("plugin", monitorType)
	// This is more of a property or separate metric in and of itself, rather
	// than a dimension.
	emitter.OmitTag("state_prefix")

	accumulator := accumulator.NewAccumulator(emitter)

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

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
