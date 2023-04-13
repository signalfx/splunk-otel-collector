// Copyright Splunk, Inc.
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

package httpsinkexporter

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type dataPoint interface {
	pmetric.NumberDataPoint | pmetric.SummaryDataPoint | pmetric.HistogramDataPoint | pmetric.ExponentialHistogramDataPoint
	Attributes() pcommon.Map
}

type dataPointSlice[D dataPoint] interface {
	Len() int
	At(int) D
}

type metric[D dataPoint, S dataPointSlice[D]] interface {
	pmetric.Gauge | pmetric.Sum | pmetric.Histogram | pmetric.ExponentialHistogram | pmetric.Summary
	DataPoints() S
}

func filterMetricByAttr[D dataPoint, S dataPointSlice[D], M metric[D, S]](m M, attrs map[string]string) bool {
	if len(attrs) == 0 {
		return true
	}
	dps := m.DataPoints()
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		dpAttrs := dp.Attributes()
		for k, v := range attrs {
			if val, found := dpAttrs.Get(k); found {
				if val.AsString() == v {
					return true
				}
			}
		}
	}
	return false
}
