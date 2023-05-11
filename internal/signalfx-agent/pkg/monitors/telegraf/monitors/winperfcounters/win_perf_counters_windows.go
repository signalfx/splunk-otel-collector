//go:build windows
// +build windows

package winperfcounters

import (
	"context"
	"strings"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"

	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType})

// GetPlugin takes a perf counter monitor config and returns a configured perf counter plugin.
// This is used for other monitors based on perf counter that manage their own life cycle
// (i.e. system utilization, windows iis)
func GetPlugin(conf *Config) (*telegrafPlugin.Win_PerfCounters, error) {
	plugin := telegrafInputs.Inputs["win_perf_counters"]().(*telegrafPlugin.Win_PerfCounters)

	// copy top level struct fields
	if err := deepcopier.Copy(conf).To(plugin); err != nil {
		return nil, err
	}

	// Telegraf has a struct wrapper around time.Duration, but it's defined
	// in an internal package which the gocomplier won't compile from
	plugin.CountersRefreshInterval.Duration = conf.CountersRefreshInterval.AsDuration()

	// copy nested perf objects
	for _, perfobj := range conf.Object {
		// The perfcounter object is an unexported struct from the original plugin.
		// We can fill this array using anonymous structs.
		plugin.Object = append(plugin.Object, struct {
			ObjectName    string
			Counters      []string
			Instances     []string
			Measurement   string
			WarnOnMissing bool
			FailOnMissing bool
			IncludeTotal  bool
		}{
			perfobj.ObjectName,
			perfobj.Counters,
			perfobj.Instances,
			perfobj.Measurement,
			perfobj.WarnOnMissing,
			perfobj.FailOnMissing,
			perfobj.IncludeTotal,
		})
	}
	return plugin, nil
}

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	plugin, err := GetPlugin(conf)
	if err != nil {
		return err
	}

	// create the emitter
	emitter := baseemitter.NewEmitter(m.Output, m.logger)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	emitter.AddTag("plugin", strings.Replace(monitorType, "/", "-", -1))

	if conf.PCRMetricNames {
		// set metric name replacements to match SignalFx PerfCounterReporter
		emitter.AddMetricNameTransformation(NewPCRMetricNamesTransformer())

		// sanitize the instance tag associated with windows perf counter metrics
		emitter.AddMeasurementTransformation(NewPCRInstanceTagTransformer())
	}

	// create the accumulator
	ac := accumulator.NewAccumulator(emitter)

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}
