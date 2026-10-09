// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build ptp_integration

package ptpreceiver

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
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
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.SocketPath = slaveSocketPath
	cfg.CollectionInterval = time.Second
	cfg.Timeout = 5 * time.Second

	sink := new(consumertest.MetricsSink)
	recv, err := factory.CreateMetrics(context.Background(), receivertest.NewNopSettings(metadata.Type), cfg, sink)
	require.NoError(t, err)
	require.NoError(t, recv.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, recv.Shutdown(context.Background()))
	})

	// Real BMCA and clock synchronization take tens of seconds: the port
	// reports a grandmaster (and UNCALIBRATED state) soon after it sees the
	// master's announce messages, but only reaches SLAVE once the clock
	// servo has run for a while.
	var last map[string]pmetric.Metric
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		batches := sink.AllMetrics()
		if !assert.NotEmpty(tt, batches, "no metrics collected yet") {
			return
		}
		got := metricsByName(batches[len(batches)-1])
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
	}, 120*time.Second, 2*time.Second)

	// The metric contract (metadata.yaml) defines six metrics; ptp.offset and
	// ptp.path.delay are only emitted once a remote grandmaster is selected,
	// which the EventuallyWithT loop above already waited for.
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
