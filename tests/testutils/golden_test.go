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

//go:build testutils

package testutils

import (
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestCompareMetricsAgainstAnyBatch(t *testing.T) {
	expected := testMetrics("expected", time.Unix(0, 0))
	batches := []pmetric.Metrics{
		testMetrics("other", time.Unix(0, 0)),
		expected,
	}

	require.NoError(t, CompareMetricsAgainstAnyBatch(expected, batches))
}

func TestCompareMetricsAgainstAnyBatchOptions(t *testing.T) {
	expected := testMetrics("expected", time.Unix(0, 0))
	actual := testMetrics("expected", time.Unix(1, 0))

	require.Error(t, CompareMetricsAgainstAnyBatch(expected, []pmetric.Metrics{actual}))
	require.NoError(t, CompareMetricsAgainstAnyBatch(
		expected,
		[]pmetric.Metrics{actual},
		pmetrictest.IgnoreTimestamp(),
	))
}

func TestCompareMetricsAgainstAnyBatchEmpty(t *testing.T) {
	require.EqualError(t,
		CompareMetricsAgainstAnyBatch(pmetric.NewMetrics(), nil),
		"no metrics batches received",
	)
}

func testMetrics(name string, timestamp time.Time) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	metric := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	metric.SetName(name)
	dataPoint := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(timestamp))
	dataPoint.SetDoubleValue(1)
	return metrics
}
