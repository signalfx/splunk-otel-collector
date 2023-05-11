// Package types exists to avoid circular references between things that need
// to reference common types
package types

// MonitorID is a unique identifier for a specific instance of a monitor
type MonitorID string

// UtilizationMetricPluginName is the name used for the plugin dimension on utilization metrics
const UtilizationMetricPluginName = "signalfx-metadata"
