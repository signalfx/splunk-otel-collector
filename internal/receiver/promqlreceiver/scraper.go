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
	api "github.com/prometheus/prometheus/web/api/v1"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type scraper struct {
	queries    []string
	httpClient *http.Client
	cfg        *Config
	settings   receiver.Settings
}

func (s *scraper) Start(ctx context.Context, host component.Host) error {
	var err error
	s.httpClient, err = s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	return err
}

func (s *scraper) Shutdown(_ context.Context) error {
	s.httpClient.CloseIdleConnections()
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

func (s *scraper) runOneQuery(ctx context.Context, endpointURL *url.URL, q string, m pmetric.Metrics) error {
	queryString := url.Values{}
	queryString.Add("query", q)
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

	resp, err := s.httpClient.Get(queryURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var promResponse api.Response
	if err = json.Unmarshal(b, &promResponse); err != nil {
		return err
	}
	if promResponse.Status != "success" {
		return fmt.Errorf("response status %q: %s", promResponse.Status, promResponse.Error)
	}

	qData := promResponse.Data.(api.QueryData)
	switch qData.ResultType {
	case parser.ValueTypeVector:
		rm := m.ResourceMetrics().AppendEmpty()
		convertVector(qData.Result.(promql.Vector), rm.ScopeMetrics().AppendEmpty())
	default:
		panic("unsupported response type " + qData.ResultType)
	}

	return nil
}

func convertVector(vector promql.Vector, scopeMetric pmetric.ScopeMetrics) {
	mfMap := make(map[string]*pmetric.Metric)

	for _, sample := range vector {
		metricName := sample.Metric.Get(model.MetricNameLabel)
		if metricName == "" {
			// PromQL operations like sum() drop the original metric name.
			metricName = "promql_result"
		}

		attrs := pcommon.NewMap()
		sample.Metric.Range(func(l labels.Label) {
			if l.Name == model.MetricNameLabel {
				return
			}
			if l.Name == model.MetricTypeLabel {
				return
			}
			attrs.PutStr(l.Name, l.Value)
		})

		metric := mfMap[metricName]
		if metric == nil {
			nm := pmetric.NewMetric()
			metric = &nm
			metric.SetName(metricName)
			mfMap[metricName] = metric
		}
		switch sample.Metric.Get(model.MetricTypeLabel) {
		case string(model.MetricTypeGauge):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyGauge()
			}
			dp := metric.Gauge().DataPoints().AppendEmpty()
			dp.SetDoubleValue(sample.F)
			dp.SetTimestamp(pcommon.Timestamp(sample.T))
			attrs.CopyTo(dp.Attributes())
		case string(model.MetricTypeCounter):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptySum()
			}
			dp := metric.Sum().DataPoints().AppendEmpty()
			dp.SetDoubleValue(sample.F)
			dp.SetTimestamp(pcommon.Timestamp(sample.T))
			attrs.CopyTo(dp.Attributes())
		case string(model.MetricTypeHistogram), string(model.MetricTypeGaugeHistogram):
			if metric.Type() == pmetric.MetricTypeEmpty {
				metric.SetEmptyHistogram()
			}
			dp := metric.Histogram().DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.Timestamp(sample.T))
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

	for _, m := range mfMap {
		m.MoveTo(scopeMetric.Metrics().AppendEmpty())
	}
}
