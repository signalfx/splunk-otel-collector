// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signalDir is where this test and testdata/integration/run.sh hand off
// control of the external master sidecar container via marker files: run.sh
// polls for a file with `docker exec`, and this test polls for a file with
// os.Stat, since the two don't otherwise share a process or a mounted
// volume.
const signalDir = "/tmp/ptp-integration-signals"

// TestPTPReceiverIntegration_GrandmasterLossRecovery exercises a scenario
// the mocked scraper_test.go can't: a real grandmaster disappearing and
// coming back mid-collection. It waits for the usual real sync, signals
// run.sh to stop the master sidecar, verifies the receiver reports losing
// the grandmaster, signals run.sh to restart the same sidecar container
// (same identity, since docker start/stop preserves it), and verifies the
// receiver re-selects the same remote grandmaster and returns to SLAVE.
func TestPTPReceiverIntegration_GrandmasterLossRecovery(t *testing.T) {
	sink := startIntegrationReceiver(t, slaveSocketPath)

	first, _ := waitForGrandmasterSync(t, sink, 120*time.Second)
	firstGMIdentity, ok := first["ptp.grandmaster.info"].Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
	require.True(t, ok, "initial sync did not report a grandmaster identity")

	baseline := len(sink.AllMetrics())
	signalReady(t, "ready-for-master-stop")
	waitForHostSignal(t, "master-stopped", 30*time.Second)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.Greater(tt, len(batches), baseline, "no scrapes observed since the master stopped") {
			return
		}
		got := metricsByName(batches[len(batches)-1])
		if !assert.Contains(tt, got, "ptp.grandmaster.present") {
			return
		}
		assert.Equal(tt, int64(0), got["ptp.grandmaster.present"].Gauge().DataPoints().At(0).IntValue(),
			"receiver should report no grandmaster once the real one disappears")
	}, 30*time.Second, time.Second)

	signalReady(t, "ready-for-master-restart")
	waitForHostSignal(t, "master-restarted", 30*time.Second)

	recovered, _ := waitForGrandmasterSync(t, sink, 120*time.Second)
	recoveredGMIdentity, ok := recovered["ptp.grandmaster.info"].Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
	require.True(t, ok, "recovery did not report a grandmaster identity")
	assert.Equal(t, firstGMIdentity.Str(), recoveredGMIdentity.Str(),
		"should re-select the same grandmaster identity after the sidecar restarts, since it's the same container")
}

// signalReady tells testdata/integration/run.sh that this test is ready for
// the next host-level action.
func signalReady(t *testing.T, name string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(signalDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(signalDir, name), nil, 0o644))
}

// waitForHostSignal waits for testdata/integration/run.sh to confirm it
// finished a host-level action (stopping or starting the master sidecar).
func waitForHostSignal(t *testing.T, name string, timeout time.Duration) {
	t.Helper()
	path := filepath.Join(signalDir, name)
	require.Eventually(t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, timeout, 500*time.Millisecond, "timed out waiting for host signal %q", name)
}
