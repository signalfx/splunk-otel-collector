// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	sfx "github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/signalfxreceiver/signalfxv2_to_metricdata_test.go

var now = time.Now()

func sfxDatapoint() *sfx.Datapoint {
	return &sfx.Datapoint{
		Metric:     "some metric",
		Timestamp:  now,
		Value:      sfx.NewIntValue(13),
		MetricType: sfx.Gauge,
		Dimensions: map[string]string{
			"k0": "v0",
			"k1": "v1",
			"k2": "v2",
		},
	}
}

func pdataMetric() (pmetric.Metrics, pmetric.Metric) {
	out := pmetric.NewMetrics()
	m := out.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	return out, m
}

func pdataMetrics(dataType pmetric.MetricType, value any, timeReceived time.Time) pmetric.Metrics {
	metrics, metric := pdataMetric()
	switch dataType {
	case pmetric.MetricTypeGauge:
		metric.SetEmptyGauge()
	case pmetric.MetricTypeSum:
		metric.SetEmptySum()
	case pmetric.MetricTypeHistogram:
		metric.SetEmptyHistogram()
	case pmetric.MetricTypeExponentialHistogram:
		metric.SetEmptyExponentialHistogram()
	case pmetric.MetricTypeSummary:
		metric.SetEmptySummary()
	}
	metric.SetName("some metric")

	var dps pmetric.NumberDataPointSlice

	switch dataType {
	case pmetric.MetricTypeGauge:
		dps = metric.Gauge().DataPoints()
	case pmetric.MetricTypeSum:
		metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		dps = metric.Sum().DataPoints()
	}

	var attributes pcommon.Map

	dp := dps.AppendEmpty()
	attributes = dp.Attributes()
	attributes.PutStr("k0", "v0")
	attributes.PutStr("k1", "v1")
	attributes.PutStr("k2", "v2")
	dp.SetTimestamp(pcommon.Timestamp(timeReceived.UnixNano()))
	switch val := value.(type) {
	case int:
		dp.SetIntValue(int64(val))
	case float64:
		dp.SetDoubleValue(val)
	}

	attributes.Clear()
	attributes.FromRaw(map[string]any{
		"k0": "v0",
		"k1": "v1",
		"k2": "v2",
	})

	return metrics
}

func TestDatapointsToPDataMetrics(t *testing.T) {
	tests := []struct {
		timeReceived    time.Time
		expectedMetrics pmetric.Metrics
		name            string
		datapoints      []*sfx.Datapoint
	}{
		{
			name:            "IntGauge",
			datapoints:      []*sfx.Datapoint{sfxDatapoint()},
			expectedMetrics: pdataMetrics(pmetric.MetricTypeGauge, 13, now),
		},
		{
			name: "DoubleGauge",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Gauge
				pt.Value = sfx.NewFloatValue(13.13)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: pdataMetrics(pmetric.MetricTypeGauge, 13.13, now),
		},
		{
			name: "IntCount",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Count
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				m := pdataMetrics(pmetric.MetricTypeSum, 13, now)
				d := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
				d.SetIsMonotonic(true)
				return m
			}(),
		},
		{
			name: "DoubleCount",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Count
				pt.Value = sfx.NewFloatValue(13.13)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				m := pdataMetrics(pmetric.MetricTypeSum, 13.13, now)
				d := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
				d.SetIsMonotonic(true)
				return m
			}(),
		},
		{
			name: "IntCounter",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Counter
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				m := pdataMetrics(pmetric.MetricTypeSum, 13, now)
				d := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				d.SetIsMonotonic(true)
				return m
			}(),
		},
		{
			name: "DoubleCounter",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Counter
				pt.Value = sfx.NewFloatValue(13.13)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				m := pdataMetrics(pmetric.MetricTypeSum, 13.13, now)
				d := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				d.SetIsMonotonic(true)
				return m
			}(),
		},
		{
			name: "with_epoch_timestamp",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Timestamp = time.Unix(0, 0)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				md := pdataMetrics(pmetric.MetricTypeGauge, 13, time.Unix(0, 0))
				md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).SetTimestamp(0)
				return md
			}(),
		},
		{
			name: "with_zero_value_timestamp",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Timestamp = time.Time{}
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: pdataMetrics(pmetric.MetricTypeGauge, 13, now),
			timeReceived:    now,
		},
		{
			name: "empty_dimension_values_accepted",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Dimensions["k0"] = ""
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				md := pdataMetrics(pmetric.MetricTypeGauge, 13, now)
				md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().PutStr("k0", "")
				return md
			}(),
		},
		{
			name:            "nil_datapoints_ignored",
			datapoints:      []*sfx.Datapoint{nil, sfxDatapoint(), nil},
			expectedMetrics: pdataMetrics(pmetric.MetricTypeGauge, 13, now),
		},
		{
			name: "drops_invalid_datapoints",
			datapoints: func() []*sfx.Datapoint {
				// nil value
				pt0 := sfxDatapoint()
				pt0.Value = nil

				// timestamps aren't supported
				pt1 := sfxDatapoint()
				pt1.MetricType = sfx.Timestamp

				// unknown enum value
				pt2 := sfxDatapoint()
				pt2.MetricType = sfx.Counter + 100

				return []*sfx.Datapoint{pt0, pt1, sfxDatapoint(), pt2}
			}(),
			expectedMetrics: pdataMetrics(pmetric.MetricTypeGauge, 13, now),
		},
		{
			name: "undesired monitorID dimension",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Meta = map[any]any{"monitorID": "undesired.value"}
				pt.Dimensions["monitorID"] = "undesired.value"
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				return pdataMetrics(pmetric.MetricTypeGauge, 13, now)
			}(),
		},
		{
			name: "desired monitorID dimension",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Meta = map[any]any{"monitorID": "undesired.value"}
				pt.Dimensions["monitorID"] = "desired.value"
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pmetric.Metrics {
				md := pdataMetrics(pmetric.MetricTypeGauge, 13, now)
				md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().PutStr("monitorID", "desired.value")
				return md
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			md := sfxDatapointsToPDataMetrics(test.datapoints, test.timeReceived, zap.NewNop())
			require.NoError(t, pmetrictest.CompareMetrics(test.expectedMetrics, md))
		})
	}
}

func TestSetDataTypeWithInvalidDatapoints(t *testing.T) {
	tests := []struct {
		name          string
		datapoint     *sfx.Datapoint
		expectedError string
	}{
		{
			name: "timestamp_as_MetricType",
			datapoint: func() *sfx.Datapoint {
				datapoint := sfxDatapoint()
				datapoint.MetricType = sfx.Timestamp
				return datapoint
			}(),
			expectedError: "unsupported metric type timestamp",
		},
		{
			name: "string_as_datapoint_value",
			datapoint: func() *sfx.Datapoint {
				datapoint := sfxDatapoint()
				datapoint.Value = sfx.NewStringValue("disallowed")
				return datapoint
			}(),
			expectedError: "unsupported value type datapoint.strWire: disallowed",
		},
		{
			name: "nonexistent_MetricType",
			datapoint: func() *sfx.Datapoint {
				datapoint := sfxDatapoint()
				datapoint.MetricType = sfx.Counter - 10000
				return datapoint
			}(),
			expectedError: "unsupported metric type datapoint.MetricType: MetricType(-",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			err := setDataTypeAndPoints(test.datapoint, pmetric.NewMetricSlice(), time.Now())
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		})
	}
}
