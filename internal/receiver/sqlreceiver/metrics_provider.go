// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlreceiver

import (
	"context"
	"strconv"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type metricsProvider struct {
	client     dbClient
	metricsCfg []Metric
}

func (mp metricsProvider) scrape(ctx context.Context) (pmetric.Metrics, error) {
	out := pmetric.NewMetrics()
	rows, err := mp.client.metricRows(ctx)
	if err != nil {
		return out, err
	}
	rms := out.ResourceMetrics()
	rm := rms.AppendEmpty()
	sms := rm.ScopeMetrics()
	sm := sms.AppendEmpty()
	ms := sm.Metrics()
	for _, metricCfg := range mp.metricsCfg {
		for _, row := range rows {
			s := row[metricCfg.ValueColumn]
			val, err := strconv.Atoi(s)
			if err != nil {
				return out, err
			}
			m := ms.AppendEmpty()
			m.SetName(metricCfg.MetricName)

			// var attrs pdata.Map
			var dps pmetric.NumberDataPointSlice
			if metricCfg.IsMonotonic {
				m.SetDataType(pmetric.MetricDataTypeSum)
				dps = m.Sum().DataPoints()
			} else {
				m.SetDataType(pmetric.MetricDataTypeGauge)
				dps = m.Gauge().DataPoints()
			}

			dp := dps.AppendEmpty()
			dp.SetIntVal(int64(val))
			attrs := dp.Attributes()
			for _, columnName := range metricCfg.AttributeColumns {
				attrVal, found := row[columnName]
				if found {
					attrs.InsertString(columnName, attrVal)
				}
			}
		}
	}
	return out, nil
}
