package dpmeta

// constants for standard datapoint Meta fields that the agent uses
const (
	// The monitor instance id
	MonitorIDMeta = "signalfx-monitor-id"
	// The monitor type that generated the datapoint
	MonitorTypeMeta = "signalfx-monitor-type"
	// The endpoint itself
	EndpointMeta = "signalfx-endpoint"
	// A hash of the configuration struct instance for the monitor instance
	// that generated the datapoint.
	ConfigHashMeta = "sfx-config-hash"
	// Should be set to true if the datapoint is not specific to the particular
	// host that collectd is running on (e.g. cluster wide metrics in a k8s
	// cluster).
	NotHostSpecificMeta = "sfx-not-host-specific"
)
