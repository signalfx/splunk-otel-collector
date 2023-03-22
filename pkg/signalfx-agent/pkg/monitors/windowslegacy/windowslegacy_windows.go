//go:build windows
// +build windows

package windowslegacy

import (
	"context"
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
				ObjectName: "Processor",
				Counters: []string{
					"% Processor Time",
					"% Privileged Time",
					"% User Time",
					"Interrupts/sec",
				},
				Instances:     []string{"*"},
				Measurement:   "processor",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "System",
				Counters: []string{
					"Processor Queue Length",
					"System Calls/sec",
					"Context Switches/sec",
				},
				Instances:     []string{"*"},
				Measurement:   "system",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "Memory",
				Counters: []string{
					"Available MBytes",
					"Pages Input/sec",
				},
				Instances:     []string{"*"},
				Measurement:   "memory",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "Paging File",
				Counters: []string{
					"% Usage",
					"% Usage Peak",
				},
				Instances:     []string{"*"},
				Measurement:   "paging_file",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "PhysicalDisk",
				Counters: []string{
					"Avg. Disk sec/Write",
					"Avg. Disk sec/Read",
					"Avg. Disk sec/Transfer",
				},
				Instances:     []string{"*"},
				Measurement:   "physicaldisk",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "LogicalDisk",
				Counters: []string{
					"Disk Read Bytes/sec",
					"Disk Write Bytes/sec",
					"Disk Transfers/sec",
					"Disk Reads/sec",
					"Disk Writes/sec",
					"Free Megabytes",
					"% Free Space",
				},
				Instances:     []string{"*"},
				Measurement:   "logicaldisk",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "Network Interface",
				Counters: []string{
					"Bytes Total/sec",
					"Bytes Received/sec",
					"Bytes Sent/sec",
					"Current Bandwidth",
					"Packets Received/sec",
					"Packets Sent/sec",
					"Packets Received Errors",
					"Packets Outbound Errors",
					"Packets Received Discarded",
					"Packets Outbound Discarded",
				},
				Instances:     []string{"*"},
				Measurement:   "network_interface",
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
	emitter.AddTag("plugin", monitorType)

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
