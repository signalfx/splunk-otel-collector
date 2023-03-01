// Copyright  Splunk, Inc.
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

package timestampprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

func newMetricAttributesProcessor(_ *zap.Logger, offsetFn func(timestamp pcommon.Timestamp) pcommon.Timestamp) processorhelper.ProcessMetricsFunc {
	return func(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
		for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
			rs := metrics.ResourceMetrics().At(i)
			for j := 0; j < rs.ScopeMetrics().Len(); j++ {
				ss := rs.ScopeMetrics().At(j)
				for k := 0; k < ss.Metrics().Len(); k++ {
					metric := ss.Metrics().At(k)
					switch metric.Type() {
					case pmetric.MetricTypeGauge:
						for l := 0; l < metric.Gauge().DataPoints().Len(); l++ {
							dp := metric.Gauge().DataPoints().At(l)
							dp.SetStartTimestamp(offsetFn(dp.StartTimestamp()))
							dp.SetTimestamp(offsetFn(dp.Timestamp()))
							for m := 0; m < dp.Exemplars().Len(); m++ {
								e := dp.Exemplars().At(m)
								e.SetTimestamp(offsetFn(e.Timestamp()))
							}
						}
					case pmetric.MetricTypeHistogram:
						for l := 0; l < metric.Histogram().DataPoints().Len(); l++ {
							dp := metric.Histogram().DataPoints().At(l)
							dp.SetStartTimestamp(offsetFn(dp.StartTimestamp()))
							dp.SetTimestamp(offsetFn(dp.Timestamp()))
							for m := 0; m < dp.Exemplars().Len(); m++ {
								e := dp.Exemplars().At(m)
								e.SetTimestamp(offsetFn(e.Timestamp()))
							}
						}
					case pmetric.MetricTypeEmpty:
					case pmetric.MetricTypeSum:
						for l := 0; l < metric.Sum().DataPoints().Len(); l++ {
							dp := metric.Sum().DataPoints().At(l)
							dp.SetStartTimestamp(offsetFn(dp.StartTimestamp()))
							dp.SetTimestamp(offsetFn(dp.Timestamp()))
							for m := 0; m < dp.Exemplars().Len(); m++ {
								e := dp.Exemplars().At(m)
								e.SetTimestamp(offsetFn(e.Timestamp()))
							}
						}
					case pmetric.MetricTypeExponentialHistogram:
						for l := 0; l < metric.ExponentialHistogram().DataPoints().Len(); l++ {
							dp := metric.ExponentialHistogram().DataPoints().At(l)
							dp.SetStartTimestamp(offsetFn(dp.StartTimestamp()))
							dp.SetTimestamp(offsetFn(dp.Timestamp()))
							for m := 0; m < dp.Exemplars().Len(); m++ {
								e := dp.Exemplars().At(m)
								e.SetTimestamp(offsetFn(e.Timestamp()))
							}
						}
					case pmetric.MetricTypeSummary:
						for l := 0; l < metric.Summary().DataPoints().Len(); l++ {
							dp := metric.Summary().DataPoints().At(l)
							dp.SetStartTimestamp(offsetFn(dp.StartTimestamp()))
							dp.SetTimestamp(offsetFn(dp.Timestamp()))
						}
					default:
						return pmetric.Metrics{}, fmt.Errorf("unsupported metric type: %v", metric.Type())
					}
				}
			}
		}
		return metrics, nil
	}
}
