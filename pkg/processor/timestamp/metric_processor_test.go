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

package timestamp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func Test_newMetricAttributesProcessor_Gauge(t *testing.T) {
	now := time.Now().UTC()
	metrics := pmetric.NewMetrics()
	m := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	gauge := m.SetEmptyGauge()
	dp := gauge.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newMetricAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newMetrics, err := proc(context.Background(), metrics)
	require.NoError(t, err)
	require.Equal(t, 1, newMetrics.MetricCount())
	result := newMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.Gauge().DataPoints().At(0).Timestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.Gauge().DataPoints().At(0).StartTimestamp().AsTime())
}

func Test_newMetricAttributesProcessor_Sum(t *testing.T) {
	now := time.Now().UTC()
	metrics := pmetric.NewMetrics()
	m := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	sum := m.SetEmptySum()
	dp := sum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newMetricAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newMetrics, err := proc(context.Background(), metrics)
	require.NoError(t, err)
	require.Equal(t, 1, newMetrics.MetricCount())
	result := newMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.Sum().DataPoints().At(0).Timestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.Sum().DataPoints().At(0).StartTimestamp().AsTime())
}

func Test_newMetricAttributesProcessor_Histogram(t *testing.T) {
	now := time.Now().UTC()
	metrics := pmetric.NewMetrics()
	m := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	sum := m.SetEmptyHistogram()
	dp := sum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newMetricAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newMetrics, err := proc(context.Background(), metrics)
	require.NoError(t, err)
	require.Equal(t, 1, newMetrics.MetricCount())
	result := newMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.Histogram().DataPoints().At(0).Timestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.Histogram().DataPoints().At(0).StartTimestamp().AsTime())
}

func Test_newMetricAttributesProcessor_ExponentionalHistogram(t *testing.T) {
	now := time.Now().UTC()
	metrics := pmetric.NewMetrics()
	m := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	sum := m.SetEmptyExponentialHistogram()
	dp := sum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newMetricAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newMetrics, err := proc(context.Background(), metrics)
	require.NoError(t, err)
	require.Equal(t, 1, newMetrics.MetricCount())
	result := newMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.ExponentialHistogram().DataPoints().At(0).Timestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.ExponentialHistogram().DataPoints().At(0).StartTimestamp().AsTime())
}

func Test_newMetricAttributesProcessor_Summary(t *testing.T) {
	now := time.Now().UTC()
	metrics := pmetric.NewMetrics()
	m := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	sum := m.SetEmptySummary()
	dp := sum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newMetricAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newMetrics, err := proc(context.Background(), metrics)
	require.NoError(t, err)
	require.Equal(t, 1, newMetrics.MetricCount())
	result := newMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.Summary().DataPoints().At(0).Timestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.Summary().DataPoints().At(0).StartTimestamp().AsTime())
}
