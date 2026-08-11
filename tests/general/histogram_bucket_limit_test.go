// Copyright Splunk, Inc.
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

//go:build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	// oversizedInputBuckets is the bucket count of the synthetic histogram sent to
	// the collector. It exceeds maxAllowedBuckets so the transform must compact it.
	oversizedInputBuckets = 41

	// maxAllowedBuckets is the SignalFx backend limit that merge_histogram_buckets(32)
	// enforces: at most 32 buckets per histogram datapoint.
	maxAllowedBuckets = 32

	// testMetricName identifies the synthetic histogram in the sink.
	testMetricName = "test.histogram.bucket.limit"
)

// TestHistogramBucketLimitReducesBuckets sends a histogram with oversizedInputBuckets
// (41) buckets through the pipeline defined in histogram_bucket_limit.yaml and asserts
// that the OTLPReceiverSink receives a histogram compacted to at most maxAllowedBuckets
// (32) buckets.
//
// This exercises the transform/limit_histogram_buckets processor that is also
// added to the default agent_config.yaml by OTL-4345.
func TestHistogramBucketLimitReducesBuckets(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	// Reserve a port for the collector's own OTLP/gRPC receiver.
	// The test pushes histograms to this port; the collector processes them and
	// forwards the result to tc.OTLPReceiverSink.
	collectorOTLPPort := testutils.GetAvailablePort(t)

	_, shutdown := tc.SplunkOtelCollectorProcess(
		"histogram_bucket_limit.yaml",
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				"COLLECTOR_OTLP_PORT": fmt.Sprintf("%d", collectorOTLPPort),
			})
		},
	)
	defer shutdown()

	client := newOTLPMetricsClient(t, fmt.Sprintf("127.0.0.1:%d", collectorOTLPPort))

	// Retry the send until the collector's OTLP receiver accepts data. Each attempt
	// is bounded by the same interval used between retries, so a stalled dial can't
	// stretch the retry cadence beyond the ticker.
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return exportHistogram(ctx, client, oversizedInputBuckets) == nil
	}, 30*time.Second, time.Second, "collector OTLP receiver did not become ready")

	// Wait for the sink to receive the processed metric and verify its bucket count.
	assert.Eventually(t, func() bool {
		for _, batch := range tc.OTLPReceiverSink.AllMetrics() {
			rms := batch.ResourceMetrics()
			for i := 0; i < rms.Len(); i++ {
				sms := rms.At(i).ScopeMetrics()
				for j := 0; j < sms.Len(); j++ {
					ms := sms.At(j).Metrics()
					for k := 0; k < ms.Len(); k++ {
						m := ms.At(k)
						if m.Name() != testMetricName || m.Type() != pmetric.MetricTypeHistogram {
							continue
						}
						dps := m.Histogram().DataPoints()
						for l := 0; l < dps.Len(); l++ {
							buckets := dps.At(l).BucketCounts().Len()
							assert.LessOrEqualf(t, buckets, maxAllowedBuckets,
								"histogram %q datapoint %d has %d buckets; want ≤ %d",
								testMetricName, l, buckets, maxAllowedBuckets)
						}
						return true
					}
				}
			}
		}
		return false
	}, 30*time.Second, 100*time.Millisecond,
		"timed out waiting for %q metric in OTLPReceiverSink", testMetricName)
}

// newOTLPMetricsClient dials the collector's OTLP/gRPC receiver at addr and returns a
// metrics export client. The underlying connection is closed when the test finishes.
func newOTLPMetricsClient(t *testing.T, addr string) pmetricotlp.GRPCClient {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	return pmetricotlp.NewGRPCClient(conn)
}

// exportHistogram sends a single synthetic histogram with numBuckets buckets to the
// collector's OTLP/gRPC receiver.
func exportHistogram(ctx context.Context, client pmetricotlp.GRPCClient, numBuckets int) error {
	payload := pmetricotlp.NewExportRequestFromMetrics(buildHistogramMetric(numBuckets))
	_, err := client.Export(ctx, payload)
	return err
}

// buildHistogramMetric constructs a pmetric.Metrics containing one explicit-bounds
// histogram datapoint with numBuckets buckets (and therefore numBuckets-1 explicit
// bounds).
func buildHistogramMetric(numBuckets int) pmetric.Metrics {
	now := time.Now()

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("splunk-otel-collector/test")

	m := sm.Metrics().AppendEmpty()
	m.SetName(testMetricName)
	m.SetDescription("Synthetic histogram used by the histogram-bucket-limit integration test")

	hist := m.SetEmptyHistogram()
	hist.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	dp := hist.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now.Add(-time.Minute)))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetCount(uint64(numBuckets))

	// numBuckets buckets require numBuckets-1 explicit bounds (1.0, 2.0, …).
	numBounds := numBuckets - 1
	bounds := dp.ExplicitBounds()
	bounds.EnsureCapacity(numBounds)
	var sum float64
	for i := 1; i <= numBounds; i++ {
		v := float64(i)
		bounds.Append(v)
		sum += v
	}
	dp.SetSum(sum)

	// One observation per bucket.
	counts := dp.BucketCounts()
	counts.EnsureCapacity(numBuckets)
	for i := 0; i < numBuckets; i++ {
		counts.Append(1)
	}

	return md
}
