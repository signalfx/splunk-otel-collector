//go:build windows && !arm64

package winperfcounters

import (
	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/ulule/deepcopier"
)

// GetPlugin takes a perf counter config and returns a configured perf counter plugin.
// Other monitors that use performance counters manage their own life cycle.
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
