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

package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestMetricsBuilder_GetValidRecord(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: "JSON unmarshal failed for OCI metric record",
		},
		{
			name:    "missing name",
			input:   `{"namespace":"ns","compartmentId":"c1"}`,
			wantErr: "no name set on OCI metric record",
		},
		{
			name:    "missing compartmentId",
			input:   `{"namespace":"ns","name":"m"}`,
			wantErr: "no compartmentId set on OCI metric record",
		},
		{
			name:    "missing namespace",
			input:   `{"compartmentId":"c1","name":"m"}`,
			wantErr: "no namespace set on OCI metric record",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := b.getValidRecord([]byte(tt.input))
			require.Nil(t, rec)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}

	rec, err := b.getValidRecord([]byte(`{"namespace":"ns","compartmentId":"c1","name":"m"}`))
	require.NoError(t, err)
	require.Equal(t, "ns", rec.Namespace)
	require.Equal(t, "c1", rec.CompartmentID)
	require.Equal(t, "m", rec.Name)
}

func TestMetricsBuilder_GetDatapoints_SkipsZeroTimestamp(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	rec := &ociMetricRecord{
		Name:      "m",
		Namespace: "ns",
		Datapoints: []ociMetricDatapoint{
			{Timestamp: 0, Value: 1.0},
			{Timestamp: 1673388760000, Value: 2.0},
		},
	}

	dps := b.getDatapoints(rec)
	require.Equal(t, 1, dps.Len())
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388760000)), dps.At(0).Timestamp())
	require.InDelta(t, 2.0, dps.At(0).DoubleValue(), .001)
}

func TestMetricsBuilder_GetDatapoints_SetsAttributesFromDimensions(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	rec := &ociMetricRecord{
		Name:       "m",
		Namespace:  "ns",
		Dimensions: map[string]any{"appName": "myApp"},
		Datapoints: []ociMetricDatapoint{{Timestamp: 1673388760000, Value: 1.0}},
	}

	dps := b.getDatapoints(rec)
	require.Equal(t, 1, dps.Len())
	appName, ok := dps.At(0).Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myApp", appName.AsString())
}

func TestMetricsBuilder_GetDatapoints_LogsInvalidDimensionValue(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	rec := &ociMetricRecord{
		Name:      "m",
		Namespace: "ns",
		// complex128 can't occur via JSON decoding but exercises the
		// FromRaw error-handling path when converting dimensions.
		Dimensions: map[string]any{"bad": complex(1, 2)},
		Datapoints: []ociMetricDatapoint{{Timestamp: 1673388760000, Value: 1.0}},
	}

	dps := b.getDatapoints(rec)
	require.Equal(t, 1, dps.Len())
	v, ok := dps.At(0).Attributes().Get("bad")
	require.True(t, ok)
	require.Equal(t, pcommon.ValueTypeEmpty, v.Type())
}

func TestMetricsBuilder_UnmarshalRecord_SkipsInvalidRecord(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	b.unmarshalRecord([]byte(`not json`))
	require.Equal(t, 0, b.build().ResourceMetrics().Len())
}

func TestMetricsBuilder_UnmarshalRecord_SkipsRecordWithoutValidDatapoints(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no datapoints",
			input: `{"namespace":"ns","compartmentId":"c1","name":"m","datapoints":[]}`,
		},
		{
			name:  "only zero timestamp datapoints",
			input: `{"namespace":"ns","compartmentId":"c1","name":"m","datapoints":[{"timestamp":0,"value":1.0}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newMetricsBuilder(zap.NewNop())
			b.unmarshalRecord([]byte(tt.input))
			require.Equal(t, 0, b.build().ResourceMetrics().Len())
		})
	}
}

func TestMetricsBuilder_UnmarshalRecord_SkipsResourceWithOnlyEmptyRecord(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"empty","datapoints":[]}`))
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"valid","datapoints":[{"timestamp":1673388760000,"value":42.0}]}`))

	md := b.build()
	require.Equal(t, 1, md.ResourceMetrics().Len())

	scopeMetrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	require.Equal(t, 1, scopeMetrics.Metrics().Len())
	require.Equal(t, "valid", scopeMetrics.Metrics().At(0).Name())
}

func TestMetricsBuilder_UnmarshalRecord_SetsGaugeType(t *testing.T) {
	for _, unit := range []string{"", "percent", "bytes", "milliseconds", "sum", "count", "Sum", "COUNT"} {
		t.Run(unit, func(t *testing.T) {
			b := newMetricsBuilder(zap.NewNop())
			b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"` + unit + `"},"datapoints":[{"timestamp":1673388760000,"value":42.0}]}`))

			md := b.build()
			require.Equal(t, 1, md.ResourceMetrics().Len())

			m := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
			require.Equal(t, pmetric.MetricTypeGauge, m.Type())
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.InDelta(t, 42.0, m.Gauge().DataPoints().At(0).DoubleValue(), .001)
		})
	}
}

func TestMetricsBuilder_UnmarshalRecord_MergesSameMetricIdentity(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"successRate","dimensions":{"appName":"myAppA"},"metadata":{"unit":"percent","displayName":"Success rate"},"datapoints":[{"timestamp":1673388760000,"value":83.0}]}`))
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"successRate","dimensions":{"appName":"myAppB"},"metadata":{"unit":"percent","displayName":"Application success rate"},"datapoints":[{"timestamp":1673388761000,"value":90.0}]}`))

	md := b.build()
	require.Equal(t, 1, md.ResourceMetrics().Len())

	scopeMetrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	require.Equal(t, 1, scopeMetrics.Metrics().Len())

	m := scopeMetrics.Metrics().At(0)
	require.Equal(t, "successRate", m.Name())
	require.Equal(t, "percent", m.Unit())
	require.Equal(t, "Application success rate", m.Description())
	require.Equal(t, 2, m.Gauge().DataPoints().Len())

	appName0, ok := m.Gauge().DataPoints().At(0).Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myAppA", appName0.AsString())
	appName1, ok := m.Gauge().DataPoints().At(1).Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myAppB", appName1.AsString())
}

func TestMetricsBuilder_Build_SortsDatapointsByTimestampAscending(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	// Datapoints are out of order both within a single record and across
	// separate records merged into the same metric.
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388762000,"value":3.0},{"timestamp":1673388760000,"value":1.0}]}`))
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388761000,"value":2.0}]}`))

	md := b.build()
	require.Equal(t, 1, md.ResourceMetrics().Len())

	m := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	dps := m.Gauge().DataPoints()
	require.Equal(t, 3, dps.Len())

	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388760000)), dps.At(0).Timestamp())
	require.InDelta(t, 1.0, dps.At(0).DoubleValue(), .001)
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388761000)), dps.At(1).Timestamp())
	require.InDelta(t, 2.0, dps.At(1).DoubleValue(), .001)
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388762000)), dps.At(2).Timestamp())
	require.InDelta(t, 3.0, dps.At(2).DoubleValue(), .001)
}

func TestMetricsBuilder_UnmarshalRecord_DoesNotMergeConflictingUnits(t *testing.T) {
	b := newMetricsBuilder(zap.NewNop())
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388760000,"value":42.0}]}`))
	b.unmarshalRecord([]byte(`{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"s"},"datapoints":[{"timestamp":1673388761000,"value":0.04}]}`))

	md := b.build()
	require.Equal(t, 1, md.ResourceMetrics().Len())

	metrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	require.Equal(t, 2, metrics.Len())

	metricsByUnit := make(map[string]pmetric.Metric, metrics.Len())
	for i := 0; i < metrics.Len(); i++ {
		metricsByUnit[metrics.At(i).Unit()] = metrics.At(i)
	}
	require.Contains(t, metricsByUnit, "ms")
	require.Contains(t, metricsByUnit, "s")
	require.InDelta(t, 42.0, metricsByUnit["ms"].Gauge().DataPoints().At(0).DoubleValue(), .001)
	require.InDelta(t, 0.04, metricsByUnit["s"].Gauge().DataPoints().At(0).DoubleValue(), .001)
}
