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
	"fmt"
	"time"

	sfx "github.com/signalfx/golib/v3/datapoint"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

var (
	errUnsupportedMetricTypeTimestamp = fmt.Errorf("unsupported metric type timestamp")
)

// Based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.15.0/receiver/signalfxreceiver/signalfxv2_to_metricdata.go
// toMetrics() will respect the timestamp of any datapoint that isn't the zero value for time.Time,
// using timeReceived otherwise.
func sfxDatapointsToPDataMetrics(datapoints []*sfx.Datapoint, timeReceived time.Time, logger *zap.Logger) pdata.Metrics {
	md := pdata.NewMetrics()
	ilm := md.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()

	metrics := ilm.Metrics()
	metrics.EnsureCapacity(len(datapoints))

	numDropped := 0
	for _, datapoint := range datapoints {
		if datapoint == nil {
			continue
		}

		m, err := setDataTypeAndPoints(datapoint, metrics, timeReceived)
		if err != nil {
			numDropped++
			logger.Debug("SignalFx datapoint type conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.String()))
			continue
		}

		m.SetName(datapoint.Metric)
	}

	if numDropped > 0 {
		logger.Debug("SendDatapoints has dropped points", zap.Int("numDropped", numDropped))
	}

	return md
}

func setDataTypeAndPoints(datapoint *sfx.Datapoint, ms pdata.MetricSlice, timeReceived time.Time) (pdata.Metric, error) {
	var m pdata.Metric
	sfxMetricType := datapoint.MetricType
	if sfxMetricType == sfx.Timestamp {
		return m, errUnsupportedMetricTypeTimestamp
	}

	switch sfxMetricType {
	case sfx.Gauge, sfx.Enum, sfx.Rate:
		switch val := datapoint.Value.(type) {
		case sfx.IntValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeIntGauge)
			fillIntDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.IntGauge().DataPoints(), timeReceived)
		case sfx.FloatValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeGauge)
			fillDoubleDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.Gauge().DataPoints(), timeReceived)
		default:
			return m, fmt.Errorf("unsupported value type %T: %v", datapoint.Value, datapoint.Value)
		}
	case sfx.Count:
		switch val := datapoint.Value.(type) {
		case sfx.IntValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.IntSum().SetIsMonotonic(true)
			fillIntDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.IntSum().DataPoints(), timeReceived)
		case sfx.FloatValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeSum)
			m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.Sum().SetIsMonotonic(true)
			fillDoubleDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.Sum().DataPoints(), timeReceived)
		default:
			return m, fmt.Errorf("unsupported value type %T: %v", datapoint.Value, datapoint.Value)
		}
	case sfx.Counter:
		switch val := datapoint.Value.(type) {
		case sfx.IntValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.IntSum().SetIsMonotonic(true)
			fillIntDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.IntSum().DataPoints(), timeReceived)
		case sfx.FloatValue:
			m = ms.AppendEmpty()
			m.SetDataType(pdata.MetricDataTypeSum)
			m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.Sum().SetIsMonotonic(true)
			fillDoubleDatapoint(val, datapoint.Timestamp, datapoint.Dimensions, m.Sum().DataPoints(), timeReceived)
		default:
			return m, fmt.Errorf("unsupported value type %T: %v", datapoint.Value, datapoint.Value)
		}
	default:
		return m, fmt.Errorf("unsupported metric type %T: %v", sfxMetricType, sfxMetricType)
	}

	return m, nil
}

func fillIntDatapoint(intValue sfx.IntValue, timestamp time.Time, dimensions map[string]string, dps pdata.IntDataPointSlice, timeReceived time.Time) {
	if timestamp.IsZero() {
		timestamp = timeReceived
	}

	dp := dps.AppendEmpty()
	dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.UnixNano())))
	dp.SetValue(intValue.Int())
	fillInLabels(dimensions, dp.LabelsMap())
}

func fillDoubleDatapoint(floatValue sfx.FloatValue, timestamp time.Time, dimensions map[string]string, dps pdata.NumberDataPointSlice, timeReceived time.Time) {
	if timestamp.IsZero() {
		timestamp = timeReceived
	}

	dp := dps.AppendEmpty()
	dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.UnixNano())))
	dp.SetValue(floatValue.Float())
	fillInLabels(dimensions, dp.LabelsMap())
}

func fillInLabels(dimensions map[string]string, labels pdata.StringMap) {
	labels.Clear()
	labels.EnsureCapacity(len(dimensions))
	for k, v := range dimensions {
		labels.Insert(k, v)
	}
}
