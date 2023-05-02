package winperfcounters

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// MetricReplacements is the default replacement set of perfcounter metric names.
var MetricReplacements = []string{
	" ", "_", // PCR bad char
	";", "_", // PCR bad char
	":", "_", // PCR bad char
	"/", "_", // PCR bad char
	"(", "_", // PCR bad char
	")", "_", // PCR bad char
	"*", "_", // PCR bad char
	"\\", "_", // PCR bad char
	"#", "num", // telegraf -> PCR
	"percent", "pct", // telegraf -> PCR
	"_persec", "_sec", // telegraf -> PCR
	"._", "_", // telegraf -> PCR (this is more of a side affect of telegraf's conversion)
	"____", "_", // telegraf -> PCR (this is also a side affect)
	"___", "_", // telegraf -> PCR (this is also a side affect)
	"__", "_", // telegraf/PCR (this is a side affect of both telegraf and PCR conversion)
}

// PerfCounterObj represents a windows performance counter object to monitor
type PerfCounterObj struct {
	// The name of a windows performance counter object
	ObjectName string `yaml:"objectName"`
	// The name of the counters to collect from the performance counter object
	Counters []string `yaml:"counters" default:"[]"`
	// The windows performance counter instances to fetch for the performance counter object
	Instances []string `yaml:"instances" default:"[]"`
	// The name of the telegraf measurement that will be used as a metric name
	Measurement string `yaml:"measurement"`
	// Log a warning if the perf counter object is missing
	WarnOnMissing bool `yaml:"warnOnMissing" default:"false"`
	// Panic if the performance counter object is missing (this will stop the agent)
	FailOnMissing bool `yaml:"failOnMissing" default:"false"`
	// Include the total instance when collecting performance counter metrics
	IncludeTotal bool `yaml:"includeTotal" default:"false"`
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" deepcopier:"skip"`
	Object               []PerfCounterObj `yaml:"objects" default:"[]"`
	// The frequency that counter paths should be expanded
	// and how often to refresh counters from configuration.
	// This is expressed as a duration.
	CountersRefreshInterval timeutil.Duration `yaml:"counterRefreshInterval" default:"5s"`
	// If `true`, instance indexes will be included in instance names, and wildcards will
	// be expanded and localized (if applicable).  If `false`, non partial wildcards will
	// be expanded and instance names will not include instance indexes.
	UseWildcardsExpansion bool `yaml:"useWildCardExpansion"`
	// Print out the configurations that match available performance counters
	PrintValid bool `yaml:"printValid"`
	// If `true`, metric names will be emitted in the format emitted by the
	// SignalFx PerfCounterReporter
	PCRMetricNames bool `yaml:"pcrMetricNames" default:"false"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	logger logrus.FieldLogger // nolint: structcheck,unused
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

// NewPCRReplacer returns a new replacer for sanitizing metricnames and instances like
// SignalFx PCR
func NewPCRReplacer() *strings.Replacer {
	return strings.NewReplacer(MetricReplacements...)
}

// NewPCRMetricNamesTransformer returns a function for tranforming perf counter
// metric names as parsed from telegraf into something matching the
// SignalFx PerfCounterReporter
func NewPCRMetricNamesTransformer() func(string) string {
	replacer := NewPCRReplacer()
	return func(in string) string {
		return replacer.Replace(strings.ToLower(in))
	}
}

// NewPCRInstanceTagTransformer returns a function for transforming perf counter measurements
func NewPCRInstanceTagTransformer() func(telegraf.Metric) error {
	replacer := NewPCRReplacer()
	return func(ms telegraf.Metric) error {
		val, ok := ms.GetTag("instance")
		if ok {
			ms.AddTag("instance", replacer.Replace(strings.ToLower(val)))
		}
		return nil
	}
}
