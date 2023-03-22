//go:build windows
// +build windows

package aspdotnet

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
				ObjectName: "ASP.NET",
				Counters: []string{
					"applications running",
					"application restarts",
					"requests current",
					"requests queued",
					"requests rejected",
					"worker processes running",
					"worker process restarts",
				},
				Instances:     []string{"*"},
				Measurement:   "asp_net",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "ASP.NET Applications",
				Counters: []string{
					"requests failed",
					"requests/sec",
					"errors during execution",
					"errors unhandled during execution/sec",
					"errors total/sec",
					"pipeline instance count",
					"sessions active",
					"session sql server connections total",
				},
				Instances:     []string{"*"},
				Measurement:   "asp_net_applications",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
		},
	}

	plugin, err := winperfcounters.GetPlugin(perfcounterConf)
	if err != nil {
		return err
	}

	// create batch emitter
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
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics from the plugin")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}
