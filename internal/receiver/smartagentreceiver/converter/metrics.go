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
	errNoIntValue                     = fmt.Errorf("no valid value for expected IntValue")
	errNoFloatValue                   = fmt.Errorf("no valid value for expected FloatValue")
)

// Based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.15.0/receiver/signalfxreceiver/signalfxv2_to_metricdata.go
// toMetrics() will respect the timestamp of any datapoint that isn't the zero value for time.Time,
// using timeReceived otherwise.
func sfxDatapointsToPDataMetrics(datapoints []*sfx.Datapoint, timeReceived time.Time, logger *zap.Logger) (pdata.Metrics, int) {
	numDropped := 0
	md := pdata.NewMetrics()
	md.ResourceMetrics().Resize(1)
	rm := md.ResourceMetrics().At(0)

	rm.InstrumentationLibraryMetrics().Resize(1)
	ilm := rm.InstrumentationLibraryMetrics().At(0)

	metrics := ilm.Metrics()
	metrics.Resize(len(datapoints))

	i := 0
	for _, datapoint := range datapoints {
		if datapoint == nil {
			continue
		}

		m := metrics.At(i)
		err := setDataType(datapoint, m)
		if err != nil {
			numDropped++
			logger.Debug("SignalFx datapoint type conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.String()))
			continue
		}

		m.SetName(datapoint.Metric)

		switch m.DataType() {
		case pdata.MetricDataTypeIntGauge:
			err = fillIntDatapoint(datapoint, m.IntGauge().DataPoints(), timeReceived)
		case pdata.MetricDataTypeIntSum:
			err = fillIntDatapoint(datapoint, m.IntSum().DataPoints(), timeReceived)
		case pdata.MetricDataTypeGauge:
			err = fillDoubleDatapoint(datapoint, m.Gauge().DataPoints(), timeReceived)
		case pdata.MetricDataTypeSum:
			err = fillDoubleDatapoint(datapoint, m.Sum().DataPoints(), timeReceived)
		}

		if err != nil {
			numDropped++
			logger.Debug("SignalFx datapoint datum conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.Metric))
			continue
		}

		i++
	}

	metrics.Resize(i)

	return md, numDropped

}

func setDataType(datapoint *sfx.Datapoint, m pdata.Metric) error {
	sfxMetricType := datapoint.MetricType
	if sfxMetricType == sfx.Timestamp {
		return errUnsupportedMetricTypeTimestamp
	}

	var isFloat bool
	switch datapoint.Value.(type) {
	case sfx.IntValue:
	case sfx.FloatValue:
		isFloat = true
	default:
		return fmt.Errorf("unsupported value type %T: %v", datapoint.Value, datapoint.Value)
	}

	switch sfxMetricType {
	case sfx.Gauge, sfx.Enum, sfx.Rate:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeGauge)
		} else {
			m.SetDataType(pdata.MetricDataTypeIntGauge)
		}
	case sfx.Count:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeSum)
			m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.Sum().SetIsMonotonic(true)
		} else {
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.IntSum().SetIsMonotonic(true)
		}
	case sfx.Counter:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeSum)
			m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.Sum().SetIsMonotonic(true)
		} else {
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.IntSum().SetIsMonotonic(true)
		}
	default:
		return fmt.Errorf("unsupported metric type %T: %v", sfxMetricType, sfxMetricType)
	}

	return nil
}

func fillIntDatapoint(datapoint *sfx.Datapoint, dps pdata.IntDataPointSlice, timeReceived time.Time) error {
	var intValue sfx.IntValue
	var ok bool
	if intValue, ok = datapoint.Value.(sfx.IntValue); !ok {
		return errNoIntValue
	}

	timestamp := datapoint.Timestamp
	if timestamp.IsZero() {
		timestamp = timeReceived
	}

	dps.Resize(1)
	dp := dps.At(0)
	dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.UnixNano())))
	dp.SetValue(intValue.Int())
	fillInLabels(datapoint.Dimensions, dp.LabelsMap())

	return nil
}

func fillDoubleDatapoint(datapoint *sfx.Datapoint, dps pdata.DoubleDataPointSlice, timeReceived time.Time) error {
	var floatValue sfx.FloatValue
	var ok bool
	if floatValue, ok = datapoint.Value.(sfx.FloatValue); !ok {
		return errNoFloatValue
	}

	timestamp := datapoint.Timestamp
	if timestamp.IsZero() {
		timestamp = timeReceived
	}

	dps.Resize(1)
	dp := dps.At(0)
	dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.UnixNano())))
	dp.SetValue(floatValue.Float())
	fillInLabels(datapoint.Dimensions, dp.LabelsMap())

	return nil

}

func fillInLabels(dimensions map[string]string, labels pdata.StringMap) {
	labels.Clear()
	labels.EnsureCapacity(len(dimensions))
	for k, v := range dimensions {
		labels.Insert(k, v)
	}
}
