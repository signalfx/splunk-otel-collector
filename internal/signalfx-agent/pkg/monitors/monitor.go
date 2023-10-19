package monitors

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// MonitorFactory is a niladic function that creates an unconfigured instance
// of a monitor.
type MonitorFactory func() interface{}

// MonitorFactories holds all of the registered monitor factories
var MonitorFactories = map[string]MonitorFactory{}

// ConfigTemplates are blank (zero-value) instances of the configuration struct
// for a particular monitor type.
var ConfigTemplates = map[string]config.MonitorCustomConfig{}

// MonitorMetadatas contains a mapping of monitor type to its metadata.
var MonitorMetadatas = map[string]*Metadata{}

// MetricInfo contains metadata about a metric.
type MetricInfo struct {
	Type       datapoint.MetricType
	Group      string
	Dimensions map[string]string
}

// Metadata describes information about a monitor.
type Metadata struct {
	MonitorType     string
	SendAll         bool
	SendUnknown     bool
	DefaultMetrics  map[string]bool
	Metrics         map[string]MetricInfo
	Groups          map[string]bool
	GroupMetricsMap map[string][]string
}

// HasMetric returns whether the metric exists at all (custom or included).
func (metadata *Metadata) HasMetric(metric string) bool {
	_, ok := metadata.Metrics[metric]
	return ok
}

// HasDefaultMetric returns whether the metric is an included metric.
func (metadata *Metadata) HasDefaultMetric(metric string) bool {
	return metadata.DefaultMetrics[metric]
}

// HasGroup returns whether the group exists or not.
func (metadata *Metadata) HasGroup(group string) bool {
	return metadata.Groups[group]
}

// NonDefaultMetrics returns list of metrics that are non-included.
// Note that it is not that efficient so cache calls if necessary or change
// implementation.
func (metadata *Metadata) NonDefaultMetrics() []string {
	var metrics []string
	for metric := range metadata.Metrics {
		if !metadata.HasDefaultMetric(metric) {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

// Register a new monitor type with the agent.  This is intended to be called
// from the init function of the module of a specific monitor
// implementation. configTemplate should be a zero-valued struct that is of the
// same type as the parameter to the Configure method for this monitor type.
func Register(metadata *Metadata, factory MonitorFactory, configTemplate config.MonitorCustomConfig) {
	if _, ok := MonitorFactories[metadata.MonitorType]; ok {
		panic("Monitor type '" + metadata.MonitorType + "' already registered")
	}
	MonitorFactories[metadata.MonitorType] = factory
	ConfigTemplates[metadata.MonitorType] = configTemplate
	MonitorMetadatas[metadata.MonitorType] = metadata
}

// Initializable represents a monitor that has a distinct InitMonitor method.
// This should be called once after the monitor is created and before any of
// its other methods are called.  It is useful for things that are not
// appropriate to do in the monitor factory function.
type Initializable interface {
	Init() error
}

// Shutdownable should be implemented by all monitors that need to clean up
// resources before being destroyed.
type Shutdownable interface {
	Shutdown()
}
