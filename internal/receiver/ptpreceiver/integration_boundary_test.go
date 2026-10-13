// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// TestPTPReceiverIntegration_BoundaryClock runs the real receiver against a
// single ptp4l instance managing two ports (a boundary clock): one on a
// network with the usual low-priority1 master sidecar (where it loses BMCA
// and becomes the SLAVE/UNCALIBRATED port syncing upstream), and one on an
// isolated network with no peers at all (where it self-elects and becomes
// the MASTER port). testdata/integration/run.sh starts this container with
// PTP4L_IFACE_ARGS="-i eth0 -i eth1" and both networks already attached, so
// LinuxPTP reports CLOCK_DESCRIPTION's boundary-clock bit regardless of
// per-port state. See TestPTPReceiverIntegration for the single-port
// counterpart.
func TestPTPReceiverIntegration_BoundaryClock(t *testing.T) {
	sink := startIntegrationReceiver(t, slaveSocketPath)

	var last map[string]pmetric.Metric
	var lastResourceAttrs pcommon.Map
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.NotEmpty(tt, batches, "no metrics collected yet") {
			return
		}
		latest := batches[len(batches)-1]
		got := metricsByName(latest)
		portMetric, ok := got["ptp.port.state"]
		if !assert.True(tt, ok, "ptp.port.state not emitted yet") {
			return
		}
		if !assert.Equal(tt, 2, portMetric.Gauge().DataPoints().Len(), "expected two PTP ports") {
			return
		}
		if !assert.True(tt, metricHasPortState(portMetric, "SLAVE"), "upstream-facing port never reached SLAVE") {
			return
		}
		if !assert.True(tt, metricHasPortState(portMetric, "MASTER"), "downstream-facing port never self-elected MASTER") {
			return
		}
		last = got
		lastResourceAttrs = latest.ResourceMetrics().At(0).Resource().Attributes()
	}, 120*time.Second, 2*time.Second)

	clockType, ok := lastResourceAttrs.Get("ptp.clock.type")
	if assert.True(t, ok, "resource missing ptp.clock.type attribute") {
		assert.Equal(t, "BC", clockType.Str(), "a ptp4l instance managing two ports should report a boundary clock")
	}

	// metadata.yaml documents that a boundary clock with both a slave and a
	// master port reports clock state SLAVE, to represent its upstream
	// synchronization.
	require.Contains(t, last, "ptp.clock.state")
	clockState, ok := last["ptp.clock.state"].Gauge().DataPoints().At(0).Attributes().Get("ptp.clock.state")
	require.True(t, ok)
	assert.Equal(t, "SLAVE", clockState.Str(), "a boundary clock with a synced upstream port should report clock state SLAVE")

	ports := last["ptp.port.state"].Gauge().DataPoints()
	identities := make(map[string]bool, ports.Len())
	for i := 0; i < ports.Len(); i++ {
		identity, ok := ports.At(i).Attributes().Get("ptp.port.identity")
		if assert.True(t, ok, "port missing ptp.port.identity attribute") {
			assert.True(t, validPortIdentity(identity.Str()), "ptp.port.identity %q is not well-formed", identity.Str())
			identities[identity.Str()] = true
		}
	}
	assert.Len(t, identities, 2, "the two ports should have distinct identities")
}
