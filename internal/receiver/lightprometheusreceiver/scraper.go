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

package lightprometheusreceiver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.uber.org/zap"
)

type scraper struct {
	client   *http.Client
	settings component.TelemetrySettings
	cfg      *Config
	name     string
}

func newScraper(
	settings receiver.CreateSettings,
	cfg *Config,
) *scraper {
	e := &scraper{
		settings: settings.TelemetrySettings,
		cfg:      cfg,
		name:     settings.ID.Name(),
	}

	return e
}

func (s *scraper) start(_ context.Context, host component.Host) error {
	var err error
	s.client, err = s.cfg.ToClient(host, s.settings)
	return err
}

type fetcher func() (io.ReadCloser, expfmt.Format, error)

func (s *scraper) scrape(context.Context) (pmetric.Metrics, error) {
	fetch := func() (io.ReadCloser, expfmt.Format, error) {
		req, err := http.NewRequest("GET", s.cfg.Endpoint, nil)
		if err != nil {
			return nil, expfmt.FmtUnknown, err
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, expfmt.FmtUnknown, err
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return nil, expfmt.FmtUnknown, fmt.Errorf("light prometheus %s returned status %d: %s", s.cfg.Endpoint, resp.StatusCode, string(body))
		}
		return resp.Body, expfmt.ResponseFormat(resp.Header), nil
	}
	return s.fetchPrometheusMetrics(fetch)
}

func (s *scraper) fetchPrometheusMetrics(fetch fetcher) (pmetric.Metrics, error) {
	metricFamilies, err := s.doFetch(fetch)
	m := pmetric.NewMetrics()
	if err != nil {
		return m, err
	}

	u, err := url.Parse(s.cfg.Endpoint)
	if err != nil {
		return m, err
	}
	rm := m.ResourceMetrics().AppendEmpty()
	res := rm.Resource()
	if s.cfg.ResourceAttributes.ServiceName.Enabled {
		res.Attributes().PutStr(conventions.AttributeServiceName, s.name)
	}
	if s.cfg.ResourceAttributes.NetHostName.Enabled {
		res.Attributes().PutStr(conventions.AttributeNetHostName, u.Host)
	}
	if s.cfg.ResourceAttributes.ServiceInstanceID.Enabled {
		res.Attributes().PutStr(conventions.AttributeServiceInstanceID, u.Host)
	}
	if s.cfg.ResourceAttributes.NetHostPort.Enabled {
		res.Attributes().PutStr(conventions.AttributeNetHostPort, u.Port())
	}
	if s.cfg.ResourceAttributes.HTTPScheme.Enabled {
		res.Attributes().PutStr(conventions.AttributeHTTPScheme, u.Scheme)
	}
	s.convertMetricFamilies(metricFamilies, rm)
	return m, nil
}

func (s *scraper) doFetch(fetch fetcher) ([]*dto.MetricFamily, error) {
	body, expformat, err := fetch()
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var decoder expfmt.Decoder
	// some "text" responses are missing \n from the last line
	if expformat != expfmt.FmtProtoDelim {
		decoder = expfmt.NewDecoder(io.MultiReader(body, strings.NewReader("\n")), expformat)
	} else {
		decoder = expfmt.NewDecoder(body, expformat)
	}

	var mfs []*dto.MetricFamily

	for {
		var mf dto.MetricFamily
		err := decoder.Decode(&mf)

		if err == io.EOF {
			return mfs, nil
		} else if err != nil {
			return nil, err
		}

		mfs = append(mfs, &mf)
	}
}

func (s *scraper) convertMetricFamilies(metricFamilies []*dto.MetricFamily, rm pmetric.ResourceMetrics) {
	now := pcommon.NewTimestampFromTime(time.Now())

	sm := rm.ScopeMetrics().AppendEmpty()
	for _, family := range metricFamilies {
		newMetric := sm.Metrics().AppendEmpty()
		newMetric.SetName(family.GetName())
		newMetric.SetDescription(family.GetHelp())
		switch *family.Type {
		case dto.MetricType_COUNTER:
			sum := newMetric.SetEmptySum()
			sum.SetIsMonotonic(true)
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			for _, fm := range family.GetMetric() {
				dp := sum.DataPoints().AppendEmpty()
				dp.SetTimestamp(now)
				dp.SetDoubleValue(fm.GetCounter().GetValue())
				for _, l := range fm.GetLabel() {
					if l.GetValue() != "" {
						dp.Attributes().PutStr(l.GetName(), l.GetValue())
					}
				}
			}
		case dto.MetricType_GAUGE:
			gauge := newMetric.SetEmptyGauge()
			for _, fm := range family.Metric {
				dp := gauge.DataPoints().AppendEmpty()
				dp.SetDoubleValue(fm.GetGauge().GetValue())
				dp.SetTimestamp(now)
				for _, l := range fm.GetLabel() {
					if l.GetValue() != "" {
						dp.Attributes().PutStr(l.GetName(), l.GetValue())
					}
				}
			}
		case dto.MetricType_HISTOGRAM, dto.MetricType_GAUGE_HISTOGRAM:
			histogram := newMetric.SetEmptyHistogram()
			histogram.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			for _, fm := range family.Metric {
				dp := histogram.DataPoints().AppendEmpty()
				dp.SetTimestamp(now)

				// Translate histogram buckets from Prometheus to the OTLP schema.
				// The bucket counts in Prometheus are cumulative, while in OTLP they are not.
				// Also, Prometheus has an extra bucket at the end for +Inf, OTLP assumes that implicitly.
				buckets := fm.GetHistogram().GetBucket()
				if len(buckets) > 0 {
					dp.BucketCounts().Append(buckets[0].GetCumulativeCount())
				}
				if len(buckets) > 1 {
					dp.ExplicitBounds().Append(buckets[0].GetUpperBound())
					lastCumulativeCount := buckets[0].GetCumulativeCount()
					for i := 1; i < len(buckets)-1; i++ {
						currentCumulativeCount := buckets[i].GetCumulativeCount()
						dp.BucketCounts().Append(currentCumulativeCount - lastCumulativeCount)
						lastCumulativeCount = currentCumulativeCount
						dp.ExplicitBounds().Append(buckets[i].GetUpperBound())
					}
					dp.BucketCounts().Append(buckets[len(buckets)-1].GetCumulativeCount() - lastCumulativeCount)
				}

				dp.SetSum(fm.GetHistogram().GetSampleSum())
				dp.SetCount(fm.GetHistogram().GetSampleCount())

				for _, l := range fm.GetLabel() {
					if l.GetValue() != "" {
						dp.Attributes().PutStr(l.GetName(), l.GetValue())
					}
				}
			}
		case dto.MetricType_SUMMARY:
			sum := newMetric.SetEmptySummary()
			for _, fm := range family.Metric {
				dp := sum.DataPoints().AppendEmpty()
				dp.SetTimestamp(now)
				for _, q := range fm.GetSummary().GetQuantile() {
					newQ := dp.QuantileValues().AppendEmpty()
					newQ.SetValue(q.GetValue())
					newQ.SetQuantile(q.GetQuantile())
				}
				dp.SetSum(fm.GetSummary().GetSampleSum())
				dp.SetCount(fm.GetSummary().GetSampleCount())
				for _, l := range fm.GetLabel() {
					if l.GetValue() != "" {
						dp.Attributes().PutStr(l.GetName(), l.GetValue())
					}
				}
			}
		case dto.MetricType_UNTYPED:
			gauge := newMetric.SetEmptyGauge()
			for _, fm := range family.Metric {
				dp := gauge.DataPoints().AppendEmpty()
				dp.SetDoubleValue(fm.GetUntyped().GetValue())
				dp.SetTimestamp(now)
				for _, l := range fm.GetLabel() {
					if l.GetValue() != "" {
						dp.Attributes().PutStr(l.GetName(), l.GetValue())
					}
				}
			}
		default:
			s.settings.Logger.Warn("Unknown metric family", zap.Any("family", family.Type))
		}
	}
}
