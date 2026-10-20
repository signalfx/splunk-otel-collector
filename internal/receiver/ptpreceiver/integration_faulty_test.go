// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPTPReceiverIntegration_FaultyPort exercises clockState()'s FAULTY
// branch, which until now had zero coverage anywhere (mocked or real). It
// waits for the usual real sync, then signals run.sh to disconnect this
// container's network interface entirely, rather than stopping the master
// as TestPTPReceiverIntegration_GrandmasterLossRecovery does. ptp4l itself
// survives losing its interface post-startup -- confirmed empirically, the
// port transitions to FAULTY but the daemon keeps running and keeps
// answering pmc over its (network-independent) Unix management socket -- so
// this is a real, deterministic way to reach FAULTY. An all-invalid-interface
// config was also tried and ruled out: ptp4l can't derive a clockIdentity
// without a valid interface at startup and refuses to start at all, rather
// than coming up FAULTY.
func TestPTPReceiverIntegration_FaultyPort(t *testing.T) {
	sink := startIntegrationReceiver(t, slaveSocketPath)

	waitForGrandmasterSync(t, sink, 120*time.Second)

	baseline := len(sink.AllMetrics())
	signalReady(t, "ready-for-network-disconnect")
	waitForHostSignal(t, "network-disconnected", 30*time.Second)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.Greater(tt, len(batches), baseline, "no scrapes observed since the network was disconnected") {
			return
		}
		got := metricsByName(batches[len(batches)-1])
		if !assert.True(tt, metricHasPortState(got["ptp.port.state"], "FAULTY"), "port never reported FAULTY after losing its network interface") {
			return
		}
		if !assert.Contains(tt, got, "ptp.clock.state") {
			return
		}
		clockState, ok := got["ptp.clock.state"].Gauge().DataPoints().At(0).Attributes().Get("ptp.clock.state")
		if !assert.True(tt, ok, "ptp.clock.state metric missing its ptp.clock.state attribute") {
			return
		}
		assert.Equal(tt, "FAULTY", clockState.Str(), "clock state should report FAULTY once its only port is FAULTY")
	}, 60*time.Second, time.Second)
}
