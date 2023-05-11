//go:build windows
// +build windows

package windowsiis

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
				ObjectName: "web service",
				Counters: []string{
					// Connections
					"current connections",     // Number of current connections to the web service
					"connection attempts/sec", // Rate that connections to web service are attempted
					// Requests
					"post requests/sec",         // Rate of HTTP POST requests
					"get requests/sec",          // Rate of HTTP GET requests
					"total method requests/sec", // Rate at which all HTTP requests are received
					// Bytes Transferred
					"bytes received/sec", // Rate that data is received by web service
					"bytes sent/sec",     // Rate that data is sent by web service
					// Files Transferred
					"files received/sec", // Rate at which files are received by web service
					"files sent/sec",     // Rate at which files are sent by web service
					// Not Found Errors
					"not found errors/sec", // Rate of 'Not Found' Errors
					// Users
					"anonymous users/sec",    // Rate at which users are making anonymous requests to the web service
					"nonanonymous users/sec", // Rate at which users are making nonanonymous requests to the web service
					// Uptime
					"service uptime", // Service uptime
					// ISAPI requests
					"isapi extension requests/sec", // Rate of ISAPI extension request processed simultaneously by the web service
				},
				Instances:     []string{"*"},
				Measurement:   "web_service",
				IncludeTotal:  true,
				WarnOnMissing: true,
			},
			{
				ObjectName: "Process",
				Counters: []string{
					// The total number of handles currently open by this
					// process. This number is equal to the sum of the handles currently open by each thread in this process.
					"Handle Count",
					// The percentage of elapsed time that all process threads used the processor to execution instructions.
					// Code executed to handle some hardware interrupts and trap conditions are included in this count.
					"% Processor Time",
					// The unique identifier of this process. ID Process numbers are reused, so they only identify a process for the lifetime of that process.
					"ID Process",
					// The current size, in bytes, of memory that this process has allocated that cannot be shared with other processes.
					"Private Bytes",
					// The number of threads currently active in this process. Every running process has at least one thread.
					"Thread Count",
					// The current size, in bytes, of the virtual address space the process is using.
					// Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages.
					// Virtual space is finite, and the process can limit its ability to load libraries.
					"Virtual Bytes",
					// The current size, in bytes, of the Working Set of this process.
					// The Working Set is the set of memory pages touched recently by the threads in the process.
					// If free memory in the computer is above a threshold, pages are left in the Working Set of a process even if they are not in use.
					// When free memory falls below a threshold, pages are trimmed from Working Sets.
					// If they are needed, they will then be soft-faulted back into the Working Set before leaving main memory.
					"Working Set",
				},
				Instances:     []string{"w3wp"},
				Measurement:   "process",
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
