package dns

import (
	"context"
	"strings"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/dns_query"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"true"`
	// Domains or subdomains to query. If this is not provided it will be
	// `["."]` and `RecordType` will be forced to `NS`.
	Domains []string `yaml:"domains"`
	// Network is the network protocol name.
	Network string `yaml:"network" default:"udp"`
	// Dns server port.
	Port int `yaml:"port" default:"53"`
	// Servers to query.
	Servers []string `yaml:"servers" validate:"required"`
	// Query record type (A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV).
	RecordType string `yaml:"recordType" default:"NS"`
	// Query timeout. This should be a duration string that is accepted
	// by https://golang.org/pkg/time/#ParseDuration.
	Timeout timeutil.Duration `yaml:"timeout" default:"2s"`
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

func (e *Emitter) AddError(err error) {
	// Suppress invalid answer errors since they will spam the logs like crazy,
	// and since the fact that it is an error is emitted as a metric anyway.
	if strings.Contains(err.Error(), "Invalid answer") {
		return
	}
	e.BaseEmitter.AddError(err)
}

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	plugin := telegrafInputs.Inputs["dns_query"]().(*telegrafPlugin.DnsQuery)
	plugin.Domains = conf.Domains
	plugin.Network = conf.Network
	plugin.Port = conf.Port
	plugin.Servers = conf.Servers
	plugin.RecordType = conf.RecordType
	plugin.Timeout = int(conf.Timeout.AsDuration().Seconds())

	emitter := &Emitter{baseemitter.NewEmitter(m.Output, m.logger)}

	// don't include the telegraf_type dimension
	emitter.SetOmitOriginalMetricType(true)

	// transform "dns_query" to "telegraf/dns"
	emitter.AddTag("plugin", strings.Replace(monitorType, "dns_query", "dns", -1))

	for _, tag := range []string{"rcode", "result"} {
		emitter.OmitTag(tag)
	}

	// transform "dns_query.my_metric" to "dns.my_metric"
	emitter.AddMetricNameTransformation(func(metric string) string {
		name := strings.Replace(metric, "dns_query", "dns", -1)
		m.logger.WithFields(log.Fields{
			"original_name": metric,
			"new_name":      name,
		}).Debug("Renaming telegraf metric")
		return name
	})

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
