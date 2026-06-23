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
	// inputBounds is the number of explicit bounds in the synthetic histogram
	// sent to the collector (40 > 31, intentionally triggering the limit).
	inputBounds = 40

	// maxAllowedBounds mirrors the SignalFx backend limit: 32 buckets = 31
	// explicit bounds. The config uses merge_histogram_buckets(32, method="limit_buckets")
	// where 32 is the maximum number of *buckets*, yielding at most 31 explicit bounds.
	maxAllowedBounds = 31

	// testMetricName is used to identify the synthetic histogram in the sink.
	testMetricName = "test.histogram.bucket.limit"
)

// TestHistogramBucketLimitReducesBuckets sends a histogram with inputBounds (40)
// explicit bounds through the pipeline defined in histogram_bucket_limit.yaml and
// asserts that the OTLPReceiverSink receives a histogram with at most maxAllowedBounds
// (31) explicit bounds.
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

	// Dial the collector's OTLP gRPC receiver.
	collectorAddr := fmt.Sprintf("127.0.0.1:%d", collectorOTLPPort)
	conn, err := grpc.NewClient(collectorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pmetricotlp.NewGRPCClient(conn)
	payload := pmetricotlp.NewExportRequestFromMetrics(buildHistogramMetric(inputBounds))

	// Wait for the collector's OTLP receiver to become ready, then send once.
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, sendErr := client.Export(ctx, payload)
		return sendErr == nil
	}, 30*time.Second, 500*time.Millisecond, "collector OTLP receiver did not become ready")

	// Wait for the sink to receive the processed metric and verify bucket count.
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
							bounds := dps.At(l).ExplicitBounds().Len()
							assert.LessOrEqualf(t, bounds, maxAllowedBounds,
								"histogram %q datapoint %d has %d explicit bounds; want ≤ %d",
								testMetricName, l, bounds, maxAllowedBounds)
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

// buildHistogramMetric constructs a pmetric.Metrics containing one explicit-bounds
// histogram datapoint with numBounds boundary values.  The resulting metric has
// numBounds+1 bucket counts (one per interval).
func buildHistogramMetric(numBounds int) pmetric.Metrics {
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
	dp.SetCount(uint64(numBounds + 1))

	// Explicit bounds: 1.0, 2.0, …, float64(numBounds).
	bounds := dp.ExplicitBounds()
	bounds.EnsureCapacity(numBounds)
	var sum float64
	for i := 1; i <= numBounds; i++ {
		v := float64(i)
		bounds.Append(v)
		sum += v
	}
	dp.SetSum(sum)

	// Bucket counts: one observation per bucket (numBounds boundaries → numBounds+1 buckets).
	counts := dp.BucketCounts()
	counts.EnsureCapacity(numBounds + 1)
	for i := 0; i <= numBounds; i++ {
		counts.Append(1)
	}

	return md
}
