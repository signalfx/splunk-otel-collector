// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

// slaveSocketPath must match uds_ro_address in testdata/integration/ptp4l-slave.conf.
const slaveSocketPath = "/var/run/ptp4l-slave.ro"

// TestPTPReceiverIntegration exercises the real receiver, the real pmc
// binary, and a real local ptp4l daemon synchronized -- via LinuxPTP's
// software timestamping, so no PTP-capable NIC is required -- to a second
// ptp4l instance acting as grandmaster. Both daemons are started by
// testdata/integration/entrypoint.sh before this test runs; see
// testdata/integration/README.md for how to run it locally in Docker.
func TestPTPReceiverIntegration(t *testing.T) {
	sink := startIntegrationReceiver(t, slaveSocketPath)

	last, lastResourceAttrs := waitForGrandmasterSync(t, sink, 120*time.Second)

	// The metric contract (metadata.yaml) defines six metrics; ptp.offset and
	// ptp.path.delay are only emitted once a remote grandmaster is selected,
	// which waitForGrandmasterSync above already waited for.
	for _, name := range []string{
		"ptp.clock.state",
		"ptp.grandmaster.info",
		"ptp.grandmaster.present",
		"ptp.offset",
		"ptp.path.delay",
		"ptp.port.state",
	} {
		assert.Contains(t, last, name, "metric contract violation: %s was not emitted", name)
	}

	for name, metric := range last {
		t.Logf("metric %s: %s", name, formatDataPoints(metric.Gauge().DataPoints()))
	}

	assertFinalGrandmasterContract(t, last, lastResourceAttrs)
	assertStartupStateHistory(t, sink.AllMetrics())
}

// startIntegrationReceiver creates and starts a real ptpreceiver pointed at
// socketPath, with a short collection interval so tests observe many samples
// over the course of real BMCA negotiation and clock-servo convergence.
func startIntegrationReceiver(t *testing.T, socketPath string) *consumertest.MetricsSink {
	t.Helper()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.SocketPath = socketPath
	cfg.CollectionInterval = time.Second
	cfg.Timeout = 5 * time.Second

	sink := new(consumertest.MetricsSink)
	recv, err := factory.CreateMetrics(context.Background(), receivertest.NewNopSettings(metadata.Type), cfg, sink)
	require.NoError(t, err)
	require.NoError(t, recv.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, recv.Shutdown(context.Background()))
	})
	return sink
}

// waitForGrandmasterSync waits until the slave reports a remote grandmaster
// and SLAVE port state, returning the metrics and resource attributes of the
// batch that first satisfied both. Real BMCA and clock synchronization take
// tens of seconds: the port reports a grandmaster (and UNCALIBRATED state)
// soon after it sees the master's announce messages, but only reaches SLAVE
// once the clock servo has run for a while.
func waitForGrandmasterSync(t *testing.T, sink *consumertest.MetricsSink, timeout time.Duration) (map[string]pmetric.Metric, pcommon.Map) {
	t.Helper()
	var last map[string]pmetric.Metric
	var lastResourceAttrs pcommon.Map
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.NotEmpty(tt, batches, "no metrics collected yet") {
			return
		}
		latest := batches[len(batches)-1]
		got := metricsByName(latest)
		if !assert.Contains(tt, got, "ptp.grandmaster.present") || !assert.Contains(tt, got, "ptp.port.state") {
			return
		}
		if !assert.Equal(tt, int64(1), got["ptp.grandmaster.present"].Gauge().DataPoints().At(0).IntValue(), "slave never saw a grandmaster") {
			return
		}
		if !assert.True(tt, metricHasPortState(got["ptp.port.state"], "SLAVE"), "no PTP port reported SLAVE state yet") {
			return
		}
		last = got
		lastResourceAttrs = latest.ResourceMetrics().At(0).Resource().Attributes()
	}, timeout, 2*time.Second)
	return last, lastResourceAttrs
}

// assertFinalGrandmasterContract checks the values and attributes of the
// final successful-path snapshot against real LinuxPTP output, not just
// metric presence: a finite, small offset and path delay (the clock servo
// has been locked onto a remote grandmaster for a while by this point), a
// well-formed grandmaster identity, the ordinary-clock type (this topology
// is a single-interface, non-boundary clock), the configured socket
// resource attribute, and a well-formed port identity.
func assertFinalGrandmasterContract(t *testing.T, last map[string]pmetric.Metric, resourceAttrs pcommon.Map) {
	const maxPlausibleOffsetOrDelaySeconds = 0.05 // generous bound for a locked servo in CI

	offset := last["ptp.offset"].Gauge().DataPoints().At(0).DoubleValue()
	assert.False(t, math.IsNaN(offset) || math.IsInf(offset, 0), "ptp.offset is not finite: %v", offset)
	assert.Less(t, math.Abs(offset), maxPlausibleOffsetOrDelaySeconds, "ptp.offset implausibly large for a locked slave: %v", offset)

	pathDelay := last["ptp.path.delay"].Gauge().DataPoints().At(0).DoubleValue()
	assert.False(t, math.IsNaN(pathDelay) || math.IsInf(pathDelay, 0), "ptp.path.delay is not finite: %v", pathDelay)
	assert.Greater(t, pathDelay, 0.0, "ptp.path.delay should be positive for software-timestamped UDP")
	assert.Less(t, pathDelay, maxPlausibleOffsetOrDelaySeconds, "ptp.path.delay implausibly large: %v", pathDelay)

	gmIdentity, ok := last["ptp.grandmaster.info"].Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
	if assert.True(t, ok, "ptp.grandmaster.info missing ptp.grandmaster.identity attribute") {
		assert.True(t, validClockIdentity(gmIdentity.Str()), "ptp.grandmaster.identity %q is not a well-formed clock identity", gmIdentity.Str())
	}

	ports := last["ptp.port.state"].Gauge().DataPoints()
	require.Equal(t, 1, ports.Len(), "this topology has a single, non-boundary PTP port")
	portIdentity, ok := ports.At(0).Attributes().Get("ptp.port.identity")
	if assert.True(t, ok, "ptp.port.state missing ptp.port.identity attribute") {
		assert.True(t, validPortIdentity(portIdentity.Str()), "ptp.port.identity %q is not well-formed", portIdentity.Str())
	}

	clockType, ok := resourceAttrs.Get("ptp.clock.type")
	if assert.True(t, ok, "resource missing ptp.clock.type attribute") {
		assert.Equal(t, "OC", clockType.Str(), "a single-interface instance with no competing ports should report an ordinary clock")
	}
	socketPath, ok := resourceAttrs.Get("ptp.socket_path")
	if assert.True(t, ok, "resource missing ptp.socket_path attribute") {
		assert.Equal(t, slaveSocketPath, socketPath.Str())
	}
}

// assertStartupStateHistory confirms the port actually transitioned through
// a pre-sync state (LISTENING or UNCALIBRATED) before reaching SLAVE, rather
// than the test only ever observing a single, already-synchronized snapshot.
// The sink retains every scrape (one per CollectionInterval) since Start, so
// the full timeline is available even though the EventuallyWithT loop above
// only inspected the latest batch on each of its own polls.
//
// LISTENING specifically is not guaranteed to be observed: the master
// sidecar is already broadcasting Announce messages before this container
// even starts, and LinuxPTP's default announce interval is ~2s, so LISTENING
// only lasts up to one announce interval -- routinely shorter than the time
// this container's `go test` takes to download modules and compile with no
// cache. By the receiver's first (immediate-on-Start) scrape, the port has
// therefore usually already moved on to UNCALIBRATED.
func assertStartupStateHistory(t *testing.T, batches []pmetric.Metrics) {
	var sawPreSync, sawSlave bool
	for _, batch := range batches {
		portMetric, ok := metricsByName(batch)["ptp.port.state"]
		if !ok {
			continue
		}
		if !sawSlave {
			sawPreSync = sawPreSync || metricHasPortState(portMetric, "LISTENING") || metricHasPortState(portMetric, "UNCALIBRATED")
		}
		sawSlave = sawSlave || metricHasPortState(portMetric, "SLAVE")
	}
	assert.True(t, sawPreSync, "port never reported a pre-sync state (LISTENING/UNCALIBRATED) before reaching SLAVE")
}

func formatDataPoints(points pmetric.NumberDataPointSlice) string {
	var parts []string
	for i := 0; i < points.Len(); i++ {
		p := points.At(i)
		var value string
		if p.ValueType() == pmetric.NumberDataPointValueTypeDouble {
			value = fmt.Sprintf("%g", p.DoubleValue())
		} else {
			value = fmt.Sprintf("%d", p.IntValue())
		}
		attrs := p.Attributes().AsRaw()
		parts = append(parts, fmt.Sprintf("value=%s attrs=%v", value, attrs))
	}
	return strings.Join(parts, ", ")
}

func metricHasPortState(metric pmetric.Metric, want string) bool {
	points := metric.Gauge().DataPoints()
	for i := 0; i < points.Len(); i++ {
		if state, ok := points.At(i).Attributes().Get("ptp.port.state"); ok && state.Str() == want {
			return true
		}
	}
	return false
}

// metricsByName reuses the allMetrics helper from scraper_test.go.
func metricsByName(metrics pmetric.Metrics) map[string]pmetric.Metric {
	got := map[string]pmetric.Metric{}
	rms := metrics.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		sms := rms.At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			for _, metric := range allMetrics(sms.At(j).Metrics()) {
				got[metric.Name()] = metric
			}
		}
	}
	return got
}
