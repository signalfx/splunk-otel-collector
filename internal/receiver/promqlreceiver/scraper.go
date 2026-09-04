// Copyright Splunk, Inc.
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

package promqlreceiver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/util/stats"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

const defaultMetricName = "promql_result"

type scraper struct {
	queries    []Query
	httpClient *http.Client
	cfg        *Config
	settings   receiver.Settings
}

type response struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data,omitempty"`
	ErrorType string          `json:"errorType,omitempty"`
	Error     string          `json:"error,omitempty"`
	Warnings  []string        `json:"warnings,omitempty"`
	Infos     []string        `json:"infos,omitempty"`
}

type queryData struct {
	Stats      stats.QueryStats `json:"stats,omitempty"`
	ResultType parser.ValueType `json:"resultType"`
	Result     json.RawMessage  `json:"result"`
}

func (s *scraper) Start(ctx context.Context, host component.Host) error {
	var err error
	s.httpClient, err = s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	return err
}

func (s *scraper) Shutdown(_ context.Context) error {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
	}
	s.httpClient = nil
	return nil
}

func (s *scraper) ScrapeMetrics(ctx context.Context) (pmetric.Metrics, error) {
	endpointURL, err := url.Parse(s.cfg.ClientConfig.Endpoint)
	m := pmetric.NewMetrics()
	if err != nil {
		return m, err
	}
	var errs []error

	for _, q := range s.queries {
		if err := s.runOneQuery(ctx, endpointURL, q, m); err != nil {
			errs = append(errs, err)
		}
	}

	return m, errors.Join(errs...)
}

func (s *scraper) runOneQuery(ctx context.Context, endpointURL *url.URL, q Query, m pmetric.Metrics) error {
	queryString := url.Values{}
	queryString.Add("query", q.Query)
	queryURL := &url.URL{
		Scheme:      endpointURL.Scheme,
		Opaque:      endpointURL.Opaque,
		User:        endpointURL.User,
		Host:        endpointURL.Host,
		Path:        endpointURL.Path,
		Fragment:    endpointURL.Fragment,
		RawQuery:    queryString.Encode(),
		RawPath:     endpointURL.RawPath,
		RawFragment: endpointURL.RawFragment,
		ForceQuery:  endpointURL.ForceQuery,
		OmitHost:    endpointURL.OmitHost,
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL.String(), http.NoBody)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var b []byte
	if b, err = io.ReadAll(resp.Body); err != nil {
		return err
	}
	var promResponse response
	if err := json.Unmarshal(b, &promResponse); err != nil {
		return err
	}
	if promResponse.Status != "success" {
		return fmt.Errorf("response status %q: %s", promResponse.Status, promResponse.Error)
	}

	var qData queryData
	if err := json.Unmarshal(promResponse.Data, &qData); err != nil {
		return err
	}

	switch qData.ResultType {
	case parser.ValueTypeVector:
		var v promql.Vector
		if err := json.Unmarshal(qData.Result, &v); err != nil {
			return err
		}
		if sm, ok := convertVector(v, q.MetricName); ok {
			rm := m.ResourceMetrics().AppendEmpty()
			sm.MoveTo(rm.ScopeMetrics().AppendEmpty())
		}
	case parser.ValueTypeNone:
		// no results
	case parser.ValueTypeString:
		s.settings.Logger.Error("PromQL response of type string is not supported:", zap.String("query", q.Query))
	case parser.ValueTypeScalar:
		s.settings.Logger.Error("PromQL response of type scalar is not supported:", zap.String("query", q.Query))
	case parser.ValueTypeMatrix:
		var matrix promql.Matrix
		if err := json.Unmarshal(qData.Result, &m); err != nil { //nolint:staticcheck // staticcheck fails to check promql.Matrix
			return err
		}
		if sm, ok := convertMatrix(matrix, q.MetricName); ok {
			rm := m.ResourceMetrics().AppendEmpty()
			sm.MoveTo(rm.ScopeMetrics().AppendEmpty())
		}
	default:
		s.settings.Logger.Error("PromQL response unsupported:", zap.String("query", q.Query), zap.String("resultType", string(qData.ResultType)))
	}

	return nil
}

func convertMatrix(matrix promql.Matrix, name string) (pmetric.ScopeMetrics, bool) {
	mfMap := make(map[string]*pmetric.Metric)

	for _, series := range matrix {
		var metricName string
		var metricType string
		var metricUnit string
		attrs := pcommon.NewMap()
		series.Metric.Range(func(l labels.Label) {
			if l.Name == model.MetricNameLabel {
				metricName = l.Value
				return
			}
			if l.Name == model.MetricTypeLabel {
				metricType = l.Value
				return
			}
			if l.Name == model.MetricUnitLabel {
				metricUnit = l.Value
				return
			}
			attrs.PutStr(l.Name, l.Value)
		})

		if metricName == "" {
			// PromQL operations like sum() drop the original metric name.
			metricName = name
			if metricName == "" {
				metricName = defaultMetricName
			}
		}

		metric := mfMap[metricName]
		if metric == nil {
			nm := pmetric.NewMetric()
			metric = &nm
			metric.SetName(metricName)
			metric.SetUnit(metricUnit)
			mfMap[metricName] = metric
		}

		switch metricType {
		case string(model.MetricTypeGauge):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyGauge()
			}
			for _, sample := range series.Floats {
				dp := metric.Gauge().DataPoints().AppendEmpty()
				dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
				dp.SetDoubleValue(sample.F)
				attrs.CopyTo(dp.Attributes())
			}
		case string(model.MetricTypeCounter):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptySum()
				metric.Sum().SetIsMonotonic(true)
				metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			}
			for _, sample := range series.Floats {
				dp := metric.Sum().DataPoints().AppendEmpty()
				dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
				dp.SetDoubleValue(sample.F)
				attrs.CopyTo(dp.Attributes())
			}
		case string(model.MetricTypeHistogram), string(model.MetricTypeGaugeHistogram):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyHistogram()
			}
			for _, sample := range series.Histograms {
				dp := metric.Histogram().DataPoints().AppendEmpty()
				dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
				dp.SetSum(sample.H.Sum)
				dp.SetCount(uint64(sample.H.Count))

				firstBucket := true
				iter := sample.H.AllBucketIterator()
				for iter.Next() {
					b := iter.At()
					if firstBucket {
						firstBucket = false
						dp.ExplicitBounds().Append(b.Lower)
					}
					dp.BucketCounts().Append(uint64(b.Count))
					dp.ExplicitBounds().Append(b.Upper)
				}
				attrs.CopyTo(dp.Attributes())
			}
		}
	}

	if len(mfMap) == 0 {
		return pmetric.NewScopeMetrics(), false
	}

	scopeMetric := pmetric.NewScopeMetrics()

	for _, m := range mfMap {
		m.MoveTo(scopeMetric.Metrics().AppendEmpty())
	}

	return scopeMetric, true
}

func convertVector(vector promql.Vector, name string) (pmetric.ScopeMetrics, bool) {
	mfMap := make(map[string]*pmetric.Metric)

	for _, sample := range vector {
		var metricName string
		var metricType string
		var metricUnit string
		attrs := pcommon.NewMap()
		sample.Metric.Range(func(l labels.Label) {
			if l.Name == model.MetricNameLabel {
				metricName = l.Value
				return
			}
			if l.Name == model.MetricTypeLabel {
				metricType = l.Value
				return
			}
			if l.Name == model.MetricUnitLabel {
				metricUnit = l.Value
				return
			}
			attrs.PutStr(l.Name, l.Value)
		})

		if metricName == "" {
			// PromQL operations like sum() drop the original metric name.
			metricName = name
			if metricName == "" {
				metricName = defaultMetricName
			}
		}

		metric := mfMap[metricName]
		if metric == nil {
			nm := pmetric.NewMetric()
			metric = &nm
			metric.SetName(metricName)
			metric.SetUnit(metricUnit)
			mfMap[metricName] = metric
		}
		switch metricType {
		case string(model.MetricTypeGauge):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyGauge()
			}
			dp := metric.Gauge().DataPoints().AppendEmpty()
			dp.SetDoubleValue(sample.F)
			dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
			attrs.CopyTo(dp.Attributes())
		case string(model.MetricTypeCounter):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptySum()
				metric.Sum().SetIsMonotonic(true)
				metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			}
			dp := metric.Sum().DataPoints().AppendEmpty()
			dp.SetDoubleValue(sample.F)
			dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
			attrs.CopyTo(dp.Attributes())
		case string(model.MetricTypeHistogram), string(model.MetricTypeGaugeHistogram):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyHistogram()
			}
			dp := metric.Histogram().DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.Timestamp(sample.T * 1e+6)) //nolint:gosec // disable G115 // convert from ms to ns
			dp.SetSum(sample.H.Sum)
			dp.SetCount(uint64(sample.H.Count))
			iter := sample.H.AllBucketIterator()
			for iter.Next() {
				b := iter.At()
				dp.BucketCounts().Append(uint64(b.Count))
			}
			attrs.CopyTo(dp.Attributes())
		}
	}

	if len(mfMap) == 0 {
		return pmetric.NewScopeMetrics(), false
	}

	scopeMetric := pmetric.NewScopeMetrics()

	for _, m := range mfMap {
		m.MoveTo(scopeMetric.Metrics().AppendEmpty())
	}

	return scopeMetric, true
}
