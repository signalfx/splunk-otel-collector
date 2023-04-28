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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	monitorIDDim = "monitorID"
)

var (
	errUnsupportedMetricTypeTimestamp = fmt.Errorf("unsupported metric type timestamp")
)

// Based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.15.0/receiver/signalfxreceiver/signalfxv2_to_metricdata.go
// toMetrics() will respect the timestamp of any datapoint that isn't the zero value for time.Time,
// using timeReceived otherwise.
func sfxDatapointsToPDataMetrics(datapoints []*sfx.Datapoint, timeReceived time.Time, logger *zap.Logger) pmetric.Metrics {
	md := pmetric.NewMetrics()
	ilm := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()

	metrics := ilm.Metrics()
	metrics.EnsureCapacity(len(datapoints))

	numDropped := 0
	for _, datapoint := range datapoints {
		if datapoint == nil {
			continue
		}

		if err := setDataTypeAndPoints(datapoint, metrics, timeReceived); err != nil {
			numDropped++
			logger.Debug("SignalFx datapoint type conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.String()))
		}
	}

	if numDropped > 0 {
		logger.Debug("SendDatapoints has dropped points", zap.Int("numDropped", numDropped))
	}

	return md
}

func setDataTypeAndPoints(datapoint *sfx.Datapoint, ms pmetric.MetricSlice, timeReceived time.Time) error {
	var m pmetric.Metric
	sfxMetricType := datapoint.MetricType
	if sfxMetricType == sfx.Timestamp {
		return errUnsupportedMetricTypeTimestamp
	}
	switch datapoint.Value.(type) {
	case sfx.IntValue, sfx.FloatValue:
		break
	default:
		return fmt.Errorf("unsupported value type %T: %v", datapoint.Value, datapoint.Value)
	}

	// isolated collectd plugins will set the "monitorID" dimension. We need to
	// delete this dimension if it matches the meta value to prevent high cardinality values
	// (especially a concern with receiver creator that uses endpoint IDs).
	if mmID, metaSet := datapoint.Meta[monitorIDDim]; metaSet {
		if dmID, dimSet := datapoint.Dimensions[monitorIDDim]; dimSet && dmID == mmID {
			delete(datapoint.Dimensions, monitorIDDim)
		}
	}

	switch sfxMetricType {
	case sfx.Gauge, sfx.Enum, sfx.Rate:
		m = ms.AppendEmpty()
		m.SetEmptyGauge()
		fillNumberDatapoint(datapoint.Value, datapoint.Timestamp, datapoint.Dimensions, m.Gauge().DataPoints(), timeReceived)
	case sfx.Count:
		m = ms.AppendEmpty()
		m.SetEmptySum()
		m.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
		m.Sum().SetIsMonotonic(true)
		fillNumberDatapoint(datapoint.Value, datapoint.Timestamp, datapoint.Dimensions, m.Sum().DataPoints(), timeReceived)
	case sfx.Counter:
		m = ms.AppendEmpty()
		m.SetEmptySum()
		m.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		m.Sum().SetIsMonotonic(true)
		fillNumberDatapoint(datapoint.Value, datapoint.Timestamp, datapoint.Dimensions, m.Sum().DataPoints(), timeReceived)
	default:
		return fmt.Errorf("unsupported metric type %T: %v", sfxMetricType, sfxMetricType)
	}

	m.SetName(datapoint.Metric)
	return nil
}

func fillNumberDatapoint(value sfx.Value, timestamp time.Time, dimensions map[string]string, dps pmetric.NumberDataPointSlice, timeReceived time.Time) {
	if timestamp.IsZero() {
		timestamp = timeReceived
	}

	dp := dps.AppendEmpty()
	dp.SetTimestamp(pcommon.Timestamp(uint64(timestamp.UnixNano())))
	switch val := value.(type) {
	case sfx.IntValue:
		dp.SetIntValue(val.Int())
	case sfx.FloatValue:
		dp.SetDoubleValue(val.Float())
	}
	attributes := dp.Attributes()
	attributes.EnsureCapacity(len(dimensions))
	for k, v := range dimensions {
		attributes.PutStr(k, v)
	}
}
