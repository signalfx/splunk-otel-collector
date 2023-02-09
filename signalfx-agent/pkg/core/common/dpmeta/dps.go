// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
