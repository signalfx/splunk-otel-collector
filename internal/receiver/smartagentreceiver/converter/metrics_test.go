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

	sfx "github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
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

func pdataMetric() (pdata.Metrics, pdata.Metric) {
	out := pdata.NewMetrics()
	m := out.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty().Metrics().AppendEmpty()
	return out, m
}

func pdataMetrics(dataType pdata.MetricDataType, value interface{}, timeReceived time.Time) pdata.Metrics {
	metrics, metric := pdataMetric()
	metric.SetDataType(dataType)
	metric.SetName("some metric")

	var dps pdata.NumberDataPointSlice

	switch dataType {
	case pdata.MetricDataTypeGauge:
		dps = metric.Gauge().DataPoints()
	case pdata.MetricDataTypeSum:
		metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		dps = metric.Sum().DataPoints()
	}

	var attributes pdata.AttributeMap

	dp := dps.AppendEmpty()
	attributes = dp.Attributes()
	attributes.Upsert("k0", pdata.NewAttributeValueString("v0"))
	attributes.Upsert("k1", pdata.NewAttributeValueString("v1"))
	attributes.Upsert("k2", pdata.NewAttributeValueString("v2"))
	attributes.Sort()
	dp.SetTimestamp(pdata.Timestamp(timeReceived.UnixNano()))
	switch val := value.(type) {
	case int:
		dp.SetIntVal(int64(val))
	case float64:
		dp.SetDoubleVal(val)
	}

	attributes.Clear()
	pdata.NewAttributeMapFromMap(map[string]pdata.AttributeValue{
		"k0": pdata.NewAttributeValueString("v0"),
		"k1": pdata.NewAttributeValueString("v1"),
		"k2": pdata.NewAttributeValueString("v2"),
	}).CopyTo(attributes)
	attributes.Sort()

	return metrics
}

func TestDatapointsToPDataMetrics(t *testing.T) {
	tests := []struct {
		timeReceived    time.Time
		expectedMetrics pdata.Metrics
		name            string
		datapoints      []*sfx.Datapoint
	}{
		{
			name:            "IntGauge",
			datapoints:      []*sfx.Datapoint{sfxDatapoint()},
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeGauge, 13, now),
		},
		{
			name: "DoubleGauge",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Gauge
				pt.Value = sfx.NewFloatValue(13.13)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeGauge, 13.13, now),
		},
		{
			name: "IntCount",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Count
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pdata.Metrics {
				m := pdataMetrics(pdata.MetricDataTypeSum, 13, now)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
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
			expectedMetrics: func() pdata.Metrics {
				m := pdataMetrics(pdata.MetricDataTypeSum, 13.13, now)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
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
			expectedMetrics: func() pdata.Metrics {
				m := pdataMetrics(pdata.MetricDataTypeSum, 13, now)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
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
			expectedMetrics: func() pdata.Metrics {
				m := pdataMetrics(pdata.MetricDataTypeSum, 13.13, now)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Sum()
				d.SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
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
			expectedMetrics: func() pdata.Metrics {
				md := pdataMetrics(pdata.MetricDataTypeGauge, 13, time.Unix(0, 0))
				md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).SetTimestamp(0)
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
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeGauge, 13, now),
			timeReceived:    now,
		},
		{
			name: "empty_dimension_values_accepted",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Dimensions["k0"] = ""
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pdata.Metrics {
				md := pdataMetrics(pdata.MetricDataTypeGauge, 13, now)
				md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes().Update("k0", pdata.NewAttributeValueString(""))
				return md
			}(),
		},
		{
			name:            "nil_datapoints_ignored",
			datapoints:      []*sfx.Datapoint{nil, sfxDatapoint(), nil},
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeGauge, 13, now),
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
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeGauge, 13, now),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			md := sfxDatapointsToPDataMetrics(test.datapoints, test.timeReceived, zap.NewNop())
			sortLabels(tt, md)

			assert.Equal(tt, test.expectedMetrics, md)
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
			_, err := setDataTypeAndPoints(test.datapoint, pdata.NewMetricSlice(), time.Now())
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		})
	}
}

func sortLabels(t *testing.T, metrics pdata.Metrics) {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.InstrumentationLibraryMetrics().Len(); j++ {
			ilm := rm.InstrumentationLibraryMetrics().At(j)
			for k := 0; k < ilm.Metrics().Len(); k++ {
				m := ilm.Metrics().At(k)
				switch m.DataType() {
				case pdata.MetricDataTypeGauge:
					for l := 0; l < m.Gauge().DataPoints().Len(); l++ {
						m.Gauge().DataPoints().At(l).Attributes().Sort()
					}
				case pdata.MetricDataTypeSum:
					for l := 0; l < m.Sum().DataPoints().Len(); l++ {
						m.Sum().DataPoints().At(l).Attributes().Sort()
					}
				default:
					t.Errorf("unexpected datatype: %v", m.DataType())
				}
			}
		}
	}
}
