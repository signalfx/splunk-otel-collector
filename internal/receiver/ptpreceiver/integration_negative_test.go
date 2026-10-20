// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

// TestPTPReceiverIntegration_NoRemoteGrandmaster runs the real receiver
// against a real ptp4l instance that has no competing master on its
// network. LinuxPTP's BMCA then selects the local clock itself: the port
// goes straight to MASTER without ever reaching SLAVE, and -- critically --
// ptp4l reports gmPresent=false even though it has a (self) gmIdentity. Run
// by testdata/integration/run.sh with no master sidecar and no shared
// network; see TestPTPReceiverIntegration for the counterpart with a real
// remote grandmaster.
func TestPTPReceiverIntegration_NoRemoteGrandmaster(t *testing.T) {
	sink := startIntegrationReceiver(t, slaveSocketPath)

	var last map[string]pmetric.Metric
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.NotEmpty(tt, batches, "no metrics collected yet") {
			return
		}
		got := metricsByName(batches[len(batches)-1])
		if !assert.True(tt, metricHasPortState(got["ptp.port.state"], "MASTER"), "port never self-elected MASTER") {
			return
		}
		last = got
	}, 60*time.Second, 2*time.Second)

	require.Contains(t, last, "ptp.grandmaster.present")
	assert.Equal(t, int64(0), last["ptp.grandmaster.present"].Gauge().DataPoints().At(0).IntValue(),
		"gmPresent should be false with no remote grandmaster on the network")
	assert.NotContains(t, last, "ptp.grandmaster.info", "ptp.grandmaster.info should be omitted with no remote grandmaster")
	assert.NotContains(t, last, "ptp.offset", "ptp.offset should be omitted with no remote grandmaster")
	assert.NotContains(t, last, "ptp.path.delay", "ptp.path.delay should be omitted with no remote grandmaster")
}

// TestPTPReceiverIntegration_SocketNotPresent points the real receiver at a
// pmc socket path that never exists and asserts it fails each scrape
// cleanly -- logging an error rather than panicking or blocking -- and never
// produces metrics. This has no ptp4l/network dependency: it exercises the
// receiver's resilience to a real missing-socket environment, the live
// counterpart to scraper_test.go's TestScrapeInvalidPMCClientSocketDirectory
// mock.
func TestPTPReceiverIntegration_SocketNotPresent(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.SocketPath = "/var/run/ptp4l-does-not-exist.ro"
	cfg.CollectionInterval = time.Second
	cfg.Timeout = 5 * time.Second

	observedLogs, logs := observer.New(zapcore.ErrorLevel)
	settings := receivertest.NewNopSettings(metadata.Type)
	settings.Logger = zap.New(observedLogs)

	sink := new(consumertest.MetricsSink)
	recv, err := factory.CreateMetrics(context.Background(), settings, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, recv.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, recv.Shutdown(context.Background()))
	})

	require.Eventually(t, func() bool {
		return logs.FilterMessage("Error scraping metrics").Len() > 0
	}, 30*time.Second, time.Second, "receiver never logged a scrape error for a missing socket")

	// scraperhelper calls the next consumer once per collection interval even
	// when every scraper fails, just with an empty pmetric.Metrics; assert no
	// data points were ever recorded rather than that the sink saw no calls.
	for _, batch := range sink.AllMetrics() {
		assert.Equal(t, 0, batch.DataPointCount(), "no metrics should be produced when every scrape fails")
	}
}

// TestPTPReceiverIntegration_PMCTimeout points PMC.Path at a real executable
// that hangs forever instead of the real pmc binary, and asserts the
// receiver's exec.CommandContext-based timeout actually kills it rather than
// wedging the collector. This isn't a real pmc hang: an earlier empirical
// probe (a Unix domain socket that accepts pmc's request and never responds)
// showed real pmc doesn't block on an unresponsive server at all -- it sends
// the request and returns almost immediately regardless, since the
// management protocol runs over SOCK_DGRAM with no connection to wait on. A
// genuinely hung subprocess only happens if the configured binary itself
// hangs, so this test substitutes one via the same PMC.Path knob a real
// deployment could point at a wrapper script. The script execs (rather than
// forks) sleep so there's exactly one process to kill, same as a real
// misbehaving pmc. Self-contained: no ptp4l/network dependency.
func TestPTPReceiverIntegration_PMCTimeout(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "pmc-hangs-forever.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexec sleep 300\n"), 0o755))

	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.SocketPath = "/tmp/ptp-pmc-timeout-test.sock"
	cfg.PMC.Path = scriptPath
	cfg.CollectionInterval = 3 * time.Second
	cfg.Timeout = 2 * time.Second

	observedLogs, logs := observer.New(zapcore.ErrorLevel)
	settings := receivertest.NewNopSettings(metadata.Type)
	settings.Logger = zap.New(observedLogs)

	sink := new(consumertest.MetricsSink)
	recv, err := factory.CreateMetrics(context.Background(), settings, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, recv.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, recv.Shutdown(context.Background()))
	})

	// The first scrape is killed by the 2s Timeout; proving the collector
	// isn't wedged on that hung process requires seeing a *second* one killed
	// too, on the next 3s collection tick.
	require.Eventually(t, func() bool {
		return logs.FilterMessage("Error scraping metrics").Len() >= 2
	}, 20*time.Second, time.Second, "receiver should recover and keep scraping after a hung pmc subprocess is killed, not wedge on the first one")

	for _, batch := range sink.AllMetrics() {
		assert.Equal(t, 0, batch.DataPointCount(), "no metrics should be produced while pmc only ever hangs")
	}

	// The script execs sleep rather than forking it, so ps/pgrep sees "sleep
	// 300" directly, not the script path. Each hung sleep gets killed by its
	// own scrape's Timeout deadline (independent of receiver shutdown), so
	// this should never find one still alive for longer than one Timeout
	// past the last log entry above.
	assert.Eventually(t, func() bool {
		return exec.Command("pgrep", "-f", "sleep 300").Run() != nil
	}, 10*time.Second, 200*time.Millisecond, "sleep process spawned for the hung pmc should get killed between scrape attempts, not leaked as an orphan")
}
