// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestMetricsBuilderStartTime(t *testing.T) {
	settings := receivertest.NewNopSettings(receivertest.NopType)
	mb := NewMetricsBuilder(NewDefaultMetricsBuilderConfig(), settings, WithStartTime(1))

	mb.RecordPtpOffsetDataPoint(10, 42)
	metrics := mb.Emit(WithStartTimeOverride(2))
	require.Equal(t, pcommon.Timestamp(2), metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).StartTimestamp())

	mb.Reset(WithStartTime(3))
	mb.RecordPtpOffsetDataPoint(20, 43)
	metrics = mb.Emit()
	require.Equal(t, pcommon.Timestamp(3), metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).StartTimestamp())
}
