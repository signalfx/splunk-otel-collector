//go:build windows
// +build windows

package dotnet

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/winperfcounters"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = logrus.WithField("monitorType", monitorType)

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	perfcounterConf := &winperfcounters.Config{
		CountersRefreshInterval: conf.CountersRefreshInterval,
		PrintValid:              conf.PrintValid,
		Object: []winperfcounters.PerfCounterObj{
			{
				ObjectName: ".NET CLR Exceptions",
				Counters: []string{
					"# of exceps thrown / sec",
				},
				Instances:     []string{"*"},
				Measurement:   "net_clr_exceptions",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: ".NET CLR LocksAndThreads",
				Counters: []string{
					"# of current logical threads",
					"# of current physical threads",
					"contention rate / sec",
					"current queue length",
				},
				Instances:     []string{"*"},
				Measurement:   "net_clr_locksandthreads",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: ".NET CLR Memory",
				Counters: []string{
					"# bytes in all heaps",
					"% time in gc",
					"# gc handles",
					"# total committed bytes",
					"# total reserved bytes",
					"# of pinned objects",
				},
				Instances:     []string{"*"},
				Measurement:   "net_clr_memory",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
		},
	}

	plugin, err := winperfcounters.GetPlugin(perfcounterConf)
	if err != nil {
		return err
	}

	// create base emitter
	emitter := baseemitter.NewEmitter(m.Output, m.logger)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	emitter.AddTag("plugin", strings.Replace(monitorType, "/", "-", -1))

	// omit objectname tag from dimensions
	emitter.OmitTag("objectname")

	// don't include the telegraf_type dimension
	emitter.SetOmitOriginalMetricType(true)

	// set metric name replacements to match SignalFx PerfCounterReporter
	emitter.AddMetricNameTransformation(winperfcounters.NewPCRMetricNamesTransformer())

	// sanitize the instance tag associated with windows perf counter metrics
	emitter.AddMeasurementTransformation(winperfcounters.NewPCRInstanceTagTransformer())

	// create the accumulator
	ac := accumulator.NewAccumulator(emitter)

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error was encountered while gathering metrics from the plugin")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}
