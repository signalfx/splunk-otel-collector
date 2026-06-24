// Package closedexporters wires closed-source exporters from the private
// github.com/signalfx/splunk-otel-collector-components module.
// This is a PoC demonstrating private Go module dependency integration.
package closedexporters

import _ "github.com/signalfx/splunk-otel-collector-components"
