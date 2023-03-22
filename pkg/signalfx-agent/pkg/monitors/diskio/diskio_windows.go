//go:build windows
// +build windows

package diskio

import (
	"context"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/winperfcounters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	conf   *Config
	filter *filter.OverridableStringFilter
	logger logrus.FieldLogger
}

// maps telegraf metricnames to sfx metricnames
var metricNameMapping = map[string]string{
	"logical_disk.Avg._Disk_Write_Queue_Length": "disk_ops.avg_write",
	"logical_disk.Avg._Disk_Bytes/Read":         "disk_octets.avg_read",
	"logical_disk.Avg._Disk_Bytes/Write":        "disk_octets.avg_write",
	"logical_disk.Avg._Disk_sec/Read":           "disk_time.avg_read",
	"logical_disk.Avg._Disk_sec/Write":          "disk_time.avg_write",
	"logical_disk.Avg._Disk_Read_Queue_Length":  "disk_ops.avg_read",
	"logical_disk.Current_Disk_Queue_Length":    diskOpsPending,
}

// applies exhuastive filter to measurements
func (m *Monitor) filterMeasurements(ms telegraf.Metric) error {
	instance, ok := ms.GetTag("instance")

	// skip it if the disk doesn't match
	if !ok || !m.filter.Matches(instance) {
		m.logger.Debugf("skipping disk '%s'", instance)
		// explicitly remove all fields to an empty map so no metrics are emitted
		for field := range ms.Fields() {
			ms.RemoveField(field)
		}
		return nil
	}

	ms.RemoveTag("instance")
	pluginInstance := strings.Replace(instance, " ", "_", -1)
	ms.AddTag("plugin_instance", pluginInstance)
	ms.AddTag("disk", pluginInstance)
	return nil
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// save conf to monitor for convenience
	m.conf = conf
	m.logger = logger.WithField("monitorID", conf.MonitorID)

	// configure filters
	var err error
	if len(conf.Disks) == 0 {
		m.filter, err = filter.NewOverridableStringFilter([]string{"*"})
		m.logger.Debugf("empty disk list defaulting to '*'")
	} else {
		m.filter, err = filter.NewOverridableStringFilter(conf.Disks)
	}

	// return an error if we can't set the filter
	if err != nil {
		return err
	}

	// get the perfcounter plugin
	plugin, err := winperfcounters.GetPlugin(&winperfcounters.Config{
		CountersRefreshInterval: conf.CountersRefreshInterval,
		PrintValid:              conf.PrintValid,
		Object: []winperfcounters.PerfCounterObj{
			{
				// The name of a windows performance counter object
				ObjectName: "LogicalDisk",
				// The name of the counters to collect from the performance counter object
				Counters: []string{
					"Avg. Disk Read Queue Length",
					"Avg. Disk Write Queue Length",
					"Avg. Disk Bytes/Read",
					"Avg. Disk Bytes/Write",
					"Avg. Disk sec/Read",
					"Avg. Disk sec/Write",
					"Current Disk Queue Length",
				},
				// The windows performance counter instances to fetch for the performance counter object
				Instances: []string{"*"},
				// The name of the telegraf measurement that will be used as a metric name
				Measurement: "logical_disk",
				// Log a warning if the perf counter object is missing
				WarnOnMissing: true,
				// Include the total instance when collecting performance counter metrics
				IncludeTotal: false,
			},
		},
	})
	if err != nil {
		return err
	}

	// create batch emitter
	emitter := baseemitter.NewEmitter(m.Output, m.logger)

	// add function to apply exhuastive filters to measurments
	emitter.AddMeasurementTransformation(m.filterMeasurements)

	// add metric map to rename metrics
	emitter.RenameMetrics(metricNameMapping)

	// don't include the telegraf_type dimension
	emitter.SetOmitOriginalMetricType(true)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	emitter.AddTag("plugin", monitorType)

	// omit objectname tag from dimensions
	emitter.OmitTag("objectname")

	// create the accumulator
	accumulator := accumulator.NewAccumulator(emitter)

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		// gather the perfcounters
		if err := plugin.Gather(accumulator); err != nil {
			m.logger.WithError(err).Errorf("unable to gather metrics from plugin")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}
