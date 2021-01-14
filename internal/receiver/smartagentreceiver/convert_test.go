// Copyright 2021, OpenTelemetry Authors
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
package smartagentreceiver

import (
	"testing"
	"time"

	sfx "github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/zap"
)

// based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/receiver/signalfxreceiver/signalfxv2_to_metricdata_test.go

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
	out.ResourceMetrics().Resize(1)
	rm := out.ResourceMetrics().At(0)
	rm.InstrumentationLibraryMetrics().Resize(1)
	ilm := rm.InstrumentationLibraryMetrics().At(0)
	ms := ilm.Metrics()

	ms.Resize(1)
	m := ms.At(0)
	return out, m
}

func pdataMetrics(dataType pdata.MetricDataType, val interface{}) pdata.Metrics {
	metrics, metric := pdataMetric()
	metric.SetDataType(dataType)
	metric.SetName("some metric")

	var dps interface{}

	switch dataType {
	case pdata.MetricDataTypeIntGauge:
		metric.IntGauge().InitEmpty()
		dps = metric.IntGauge().DataPoints()
	case pdata.MetricDataTypeIntSum:
		metric.IntSum().InitEmpty()
		metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		dps = metric.IntSum().DataPoints()
	case pdata.MetricDataTypeDoubleGauge:
		metric.DoubleGauge().InitEmpty()
		dps = metric.DoubleGauge().DataPoints()
	case pdata.MetricDataTypeDoubleSum:
		metric.DoubleSum().InitEmpty()
		metric.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		dps = metric.DoubleSum().DataPoints()
	}

	var labels pdata.StringMap

	switch dataType {
	case pdata.MetricDataTypeIntGauge, pdata.MetricDataTypeIntSum:
		dps.(pdata.IntDataPointSlice).Resize(1)
		dp := dps.(pdata.IntDataPointSlice).At(0)
		labels = dp.LabelsMap()
		dp.SetTimestamp(pdata.TimestampUnixNano(now.UnixNano()))
		dp.SetValue(int64(val.(int)))
	case pdata.MetricDataTypeDoubleGauge, pdata.MetricDataTypeDoubleSum:
		dps.(pdata.DoubleDataPointSlice).Resize(1)
		dp := dps.(pdata.DoubleDataPointSlice).At(0)
		labels = dp.LabelsMap()
		dp.SetTimestamp(pdata.TimestampUnixNano(now.UnixNano()))
		dp.SetValue(val.(float64))
	}

	labels.InitFromMap(map[string]string{
		"k0": "v0",
		"k1": "v1",
		"k2": "v2",
	})
	labels.Sort()

	return metrics
}

func TestToMetrics(t *testing.T) {
	tests := []struct {
		name            string
		datapoints      []*sfx.Datapoint
		expectedMetrics pdata.Metrics
		expectedDropped int
	}{
		{
			name:            "IntGauge",
			datapoints:      []*sfx.Datapoint{sfxDatapoint()},
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeIntGauge, 13),
		},
		{
			name: "DoubleGauge",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Gauge
				pt.Value = sfx.NewFloatValue(13.13)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeDoubleGauge, 13.13),
		},
		{
			name: "IntCount",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.MetricType = sfx.Count
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pdata.Metrics {
				m := pdataMetrics(pdata.MetricDataTypeIntSum, 13)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).IntSum()
				d.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
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
				m := pdataMetrics(pdata.MetricDataTypeDoubleSum, 13.13)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).DoubleSum()
				d.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
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
				m := pdataMetrics(pdata.MetricDataTypeIntSum, 13)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).IntSum()
				d.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
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
				m := pdataMetrics(pdata.MetricDataTypeDoubleSum, 13.13)
				d := m.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).DoubleSum()
				d.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
				d.SetIsMonotonic(true)
				return m
			}(),
		},
		{
			name: "with_zero_timestamp",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Timestamp = time.Unix(0, 0)
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pdata.Metrics {
				md := pdataMetrics(pdata.MetricDataTypeIntGauge, 13)
				md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).IntGauge().DataPoints().At(0).SetTimestamp(0)
				return md
			}(),
		},
		{
			name: "empty_dimension_values_accepted",
			datapoints: func() []*sfx.Datapoint {
				pt := sfxDatapoint()
				pt.Dimensions["k0"] = ""
				return []*sfx.Datapoint{pt}
			}(),
			expectedMetrics: func() pdata.Metrics {
				md := pdataMetrics(pdata.MetricDataTypeIntGauge, 13)
				md.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics().At(0).IntGauge().DataPoints().At(0).LabelsMap().Update("k0", "")
				return md
			}(),
		},
		{
			name:            "nil_datapoints_ignored",
			datapoints:      []*sfx.Datapoint{nil, sfxDatapoint(), nil},
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeIntGauge, 13),
			expectedDropped: 0,
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

				return []*sfx.Datapoint{
					pt0, pt1, sfxDatapoint(), pt2}
			}(),
			expectedMetrics: pdataMetrics(pdata.MetricDataTypeIntGauge, 13),
			expectedDropped: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			converter := Converter{logger: zap.NewNop()}
			md, dropped := converter.toMetrics(test.datapoints)
			sortLabels(tt, md)

			assert.Equal(tt, test.expectedMetrics, md)
			assert.Equal(tt, test.expectedDropped, dropped)
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
			_, metric := pdataMetric()
			err := setDataType(test.datapoint, metric)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		})
	}
}

func TestFillIntDatapointWithInvalidValue(t *testing.T) {
	datapoint := sfxDatapoint()
	datapoint.MetricType = sfx.Gauge
	datapoint.Value = sfx.NewIntValue(123)

	_, metric := pdataMetric()
	setDataType(datapoint, metric)
	gauge := metric.IntGauge()
	gauge.InitEmpty()

	datapoint.Value = sfx.NewFloatValue(123.45)
	err := fillIntDatapoint(datapoint, gauge.DataPoints())
	require.Error(t, err)
	assert.EqualError(t, err, "no valid value for expected IntValue")
}

func TestFillDoubleDatapointWithInvalidValue(t *testing.T) {
	datapoint := sfxDatapoint()
	datapoint.MetricType = sfx.Gauge
	datapoint.Value = sfx.NewFloatValue(123.45)

	_, metric := pdataMetric()
	setDataType(datapoint, metric)
	gauge := metric.DoubleGauge()
	gauge.InitEmpty()

	datapoint.Value = sfx.NewIntValue(123)
	err := fillDoubleDatapoint(datapoint, gauge.DataPoints())
	require.Error(t, err)
	assert.EqualError(t, err, "no valid value for expected FloatValue")
}

func sortLabels(t *testing.T, metrics pdata.Metrics) {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.InstrumentationLibraryMetrics().Len(); j++ {
			ilm := rm.InstrumentationLibraryMetrics().At(j)
			for k := 0; k < ilm.Metrics().Len(); k++ {
				m := ilm.Metrics().At(k)
				switch m.DataType() {
				case pdata.MetricDataTypeIntGauge:
					for l := 0; l < m.IntGauge().DataPoints().Len(); l++ {
						m.IntGauge().DataPoints().At(l).LabelsMap().Sort()
					}
				case pdata.MetricDataTypeIntSum:
					for l := 0; l < m.IntSum().DataPoints().Len(); l++ {
						m.IntSum().DataPoints().At(l).LabelsMap().Sort()
					}
				case pdata.MetricDataTypeDoubleGauge:
					for l := 0; l < m.DoubleGauge().DataPoints().Len(); l++ {
						m.DoubleGauge().DataPoints().At(l).LabelsMap().Sort()
					}
				case pdata.MetricDataTypeDoubleSum:
					for l := 0; l < m.DoubleSum().DataPoints().Len(); l++ {
						m.DoubleSum().DataPoints().At(l).LabelsMap().Sort()
					}
				default:
					t.Errorf("unexpected datatype: %v", m.DataType())
				}
			}
		}
	}
}
