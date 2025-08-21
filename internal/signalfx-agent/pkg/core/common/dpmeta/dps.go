package dpmeta

// constants for standard datapoint Meta fields that the agent uses
const (
	// Should be set to true if the datapoint is not specific to the particular
	// host that collectd is running on (e.g. cluster wide metrics in a k8s
	// cluster).
	NotHostSpecificMeta = "sfx-not-host-specific"
)
