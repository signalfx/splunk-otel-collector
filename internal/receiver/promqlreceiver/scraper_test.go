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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/util/features"
	"github.com/prometheus/prometheus/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type dbAdapter struct {
	*tsdb.DB
}

func (a *dbAdapter) BlockMetas() ([]tsdb.BlockMeta, error) {
	return a.DB.BlockMetas(), nil
}

func (a *dbAdapter) Stats(statsByLabelName string, limit int) (*tsdb.Stats, error) {
	return a.Head().Stats(statsByLabelName, limit), nil
}

func (*dbAdapter) WALReplayStatus() (tsdb.WALReplayStatus, error) {
	return tsdb.WALReplayStatus{}, nil
}

func TestQueryPrometheusAPI(t *testing.T) {
	tempDir := t.TempDir()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	activeQueryTracker, err := promql.NewActiveQueryTracker(tempDir, 1, slog.Default())
	require.NoError(t, err)

	promqlParser := parser.NewParser(parser.Options{})

	opts := promql.EngineOpts{
		Logger:             slog.Default(),
		Reg:                prometheus.DefaultRegisterer,
		MaxSamples:         1000,
		Timeout:            1 * time.Second,
		ActiveQueryTracker: activeQueryTracker,
		LookbackDelta:      1 * time.Hour,
		NoStepSubqueryIntervalFn: func(rangeMillis int64) int64 {
			return rangeMillis
		},
		// EnableAtModifier and EnableNegativeOffset have to be
		// always on for regular PromQL as of Prometheus v2.33.
		EnableAtModifier:     true,
		EnableNegativeOffset: true,
		FeatureRegistry:      features.DefaultRegistry,
		Parser:               promqlParser,
	}

	queryEngine := promql.NewEngine(opts)

	db, err := tsdb.Open(
		tempDir,
		slog.Default(),
		prometheus.DefaultRegisterer,
		tsdb.DefaultOptions(),
		tsdb.NewDBStats(),
	)
	require.NoError(t, err)
	appender := db.AppenderV2(t.Context())
	_, err = appender.Append(storage.SeriesRef(0), promlabels.New(promlabels.Label{Name: "foo", Value: "bar"}, promlabels.Label{Name: model.MetricNameLabel, Value: "up"}, promlabels.Label{Name: model.MetricTypeLabel, Value: string(model.MetricTypeGauge)}), time.Now().Unix(), time.Now().UnixMilli(), 1, nil, nil, storage.AppendV2Options{
		MetricFamilyName: "up",
		Metadata:         metadata.Metadata{Type: model.MetricTypeGauge},
	})
	require.NoError(t, err)
	_, err = appender.Append(storage.SeriesRef(0), promlabels.New(promlabels.Label{Name: "foo", Value: "foobar"}, promlabels.Label{Name: model.MetricNameLabel, Value: "up"}, promlabels.Label{Name: model.MetricTypeLabel, Value: string(model.MetricTypeGauge)}), time.Now().Unix(), time.Now().UnixMilli(), 1, nil, nil, storage.AppendV2Options{
		MetricFamilyName: "up",
		Metadata:         metadata.Metadata{Type: model.MetricTypeGauge},
	})
	require.NoError(t, err)
	require.NoError(t, appender.Commit())
	s := dbAdapter{db}

	externalURL, err := url.Parse("http://localhost:9090")
	require.NoError(t, err)
	webConfig := &web.Options{
		QueryEngine:     queryEngine,
		ExternalURL:     externalURL,
		ListenAddresses: []string{"127.0.0.1:9090"},
		RoutePrefix:     "/",
		Context:         t.Context(),
		EnableLifecycle: true,
		EnableSearch:    true,
		EnableAdminAPI:  true,
		IsAgent:         false,
		Gatherer:        prometheus.DefaultGatherer,
		FeatureRegistry: features.DefaultRegistry,
		Registerer:      prometheus.DefaultRegisterer,
		TSDBDir:         tempDir,
		Storage:         s,
		ReadTimeout:     1 * time.Second,
		Version:         &web.PrometheusVersion{},
		MaxConnections:  200,
		MaxSearchLimit:  200,
	}
	webHandler := web.New(slog.Default(), webConfig)

	listeners, err := webHandler.Listeners()
	require.NoError(t, err)

	webHandler.SetReady(web.Ready)
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = webHandler.Run(t.Context(), listeners, "")
	})

	// now set up our receiver and query:
	f := NewFactory()
	sink := &consumertest.MetricsSink{}
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Queries = []Query{
		{
			Query: "up",
		},
		{
			Query:              "count(up)",
			MetricNameFallback: "myups",
		},
	}
	cfg.ControllerConfig.CollectionInterval = 1 * time.Second
	cfg.ClientConfig.Endpoint = "http://localhost:9090/api/v1/query"
	cfg.ClientConfig.TLS.Insecure = true
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = logger
	r, err := f.CreateMetrics(t.Context(), settings, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(t.Context(), componenttest.NewNopHost()))

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		require.NotEmpty(tt, sink.AllMetrics())
		require.Positive(tt, sink.AllMetrics()[0].MetricCount())
	}, 10*time.Second, 1*time.Second)

	require.NoError(t, r.Shutdown(t.Context()))

	webHandler.Quit()
	require.NoError(t, db.Close())
	require.NoError(t, activeQueryTracker.Close())

	firstMetric := sink.AllMetrics()[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, "up", firstMetric.Name())
	require.Equal(t, pmetric.MetricTypeGauge, firstMetric.Type())
	require.Positive(t, firstMetric.Gauge().DataPoints().Len())
	require.Equal(t, 1.0, firstMetric.Gauge().DataPoints().At(0).DoubleValue())
	require.Equal(t, "bar", firstMetric.Gauge().DataPoints().At(0).Attributes().AsRaw()["foo"])
	secondMetric := sink.AllMetrics()[0].ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, "myups", secondMetric.Name())
	require.Equal(t, pmetric.MetricTypeGauge, secondMetric.Type())
	require.Positive(t, secondMetric.Gauge().DataPoints().Len())
	require.Equal(t, 2.0, secondMetric.Gauge().DataPoints().At(0).DoubleValue())
}

// promAPIResponse marshals the given result into the same envelope the Prometheus HTTP API
// returns from /api/v1/query and /api/v1/query_range.
func promAPIResponse(t *testing.T, resultType parser.ValueType, result any) []byte {
	t.Helper()
	resultBytes, err := json.Marshal(result)
	require.NoError(t, err)
	b, err := json.Marshal(struct {
		Status string `json:"status"`
		Data   struct {
			ResultType parser.ValueType `json:"resultType"`
			Result     json.RawMessage  `json:"result"`
		} `json:"data"`
	}{
		Status: "success",
		Data: struct {
			ResultType parser.ValueType `json:"resultType"`
			Result     json.RawMessage  `json:"result"`
		}{ResultType: resultType, Result: resultBytes},
	})
	require.NoError(t, err)
	return b
}

func newTestScraper(t *testing.T, handler http.HandlerFunc) *scraper {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Queries = []Query{{Query: "up"}}
	cfg.ClientConfig.Endpoint = srv.URL

	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = zaptest.NewLogger(t)

	s := &scraper{queries: cfg.Queries, cfg: cfg, settings: settings}
	require.NoError(t, s.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, s.Shutdown(t.Context())) })
	return s
}

func TestScrapeMetricsVectorGauge(t *testing.T) {
	v := model.Vector{
		&model.Sample{
			Metric:    model.Metric{model.MetricNameLabel: "up", model.MetricTypeLabel: "gauge", "foo": "bar"},
			Value:     42,
			Timestamp: model.TimeFromUnix(1700000000),
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, metrics.MetricCount())

	m := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, "up", m.Name())
	require.Equal(t, pmetric.MetricTypeGauge, m.Type())
	dp := m.Gauge().DataPoints().At(0)
	require.InDelta(t, 42, dp.DoubleValue(), 0)
	require.Equal(t, uint64(1700000000)*1e9, uint64(dp.Timestamp()))
	require.Equal(t, "bar", dp.Attributes().AsRaw()["foo"])
}

func TestScrapeMetricsVectorCounter(t *testing.T) {
	v := model.Vector{
		&model.Sample{
			Metric:    model.Metric{model.MetricNameLabel: "requests_total", model.MetricTypeLabel: "counter"},
			Value:     7,
			Timestamp: model.TimeFromUnix(1700000000),
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)

	m := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, pmetric.MetricTypeSum, m.Type())
	require.True(t, m.Sum().IsMonotonic())
	require.Equal(t, pmetric.AggregationTemporalityCumulative, m.Sum().AggregationTemporality())
	require.InDelta(t, 7, m.Sum().DataPoints().At(0).DoubleValue(), 0)
}

func TestScrapeMetricsVectorHistogram(t *testing.T) {
	v := model.Vector{
		&model.Sample{
			Metric: model.Metric{model.MetricNameLabel: "latency", model.MetricTypeLabel: "histogram"},
			Histogram: &model.SampleHistogram{
				Count: 10,
				Sum:   25.3,
				Buckets: model.HistogramBuckets{
					{Boundaries: 3, Lower: 0.1, Upper: 0.2, Count: 5},
					{Boundaries: 3, Lower: 0.2, Upper: 0.4, Count: 5},
				},
			},
			Timestamp: model.TimeFromUnix(1700000000),
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)

	m := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, pmetric.MetricTypeHistogram, m.Type())
	dp := m.Histogram().DataPoints().At(0)
	require.Equal(t, uint64(10), dp.Count())
	require.InDelta(t, 25.3, dp.Sum(), 0.0001)
	require.Equal(t, []float64{0.1, 0.2, 0.4}, dp.ExplicitBounds().AsRaw())
	require.Equal(t, []uint64{5, 5}, dp.BucketCounts().AsRaw())
}

func TestScrapeMetricsMatrixCounter(t *testing.T) {
	matrix := model.Matrix{
		&model.SampleStream{
			Metric: model.Metric{model.MetricNameLabel: "requests_total", model.MetricTypeLabel: "counter"},
			Values: []model.SamplePair{
				{Timestamp: model.TimeFromUnix(1700000000), Value: 1},
				{Timestamp: model.TimeFromUnix(1700000001), Value: 2},
			},
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeMatrix, matrix))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)

	m := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, pmetric.MetricTypeSum, m.Type())
	require.Equal(t, 2, m.Sum().DataPoints().Len())
	require.InDelta(t, 1, m.Sum().DataPoints().At(0).DoubleValue(), 0)
	require.InDelta(t, 2, m.Sum().DataPoints().At(1).DoubleValue(), 0)
}

func TestScrapeMetricsMatrixHistogram(t *testing.T) {
	matrix := model.Matrix{
		&model.SampleStream{
			Metric: model.Metric{model.MetricNameLabel: "latency", model.MetricTypeLabel: "gaugehistogram"},
			Histograms: []model.SampleHistogramPair{
				{
					Timestamp: model.TimeFromUnix(1700000000),
					Histogram: &model.SampleHistogram{
						Count:   4,
						Sum:     8,
						Buckets: model.HistogramBuckets{{Boundaries: 3, Lower: 0, Upper: 1, Count: 4}},
					},
				},
			},
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeMatrix, matrix))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)

	m := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	require.Equal(t, pmetric.MetricTypeHistogram, m.Type())
	require.Equal(t, uint64(4), m.Histogram().DataPoints().At(0).Count())
}

func TestScrapeMetricsMissingMetricNameUsesConfiguredName(t *testing.T) {
	v := model.Vector{
		&model.Sample{
			// PromQL aggregations like sum() drop __name__.
			Metric:    model.Metric{model.MetricTypeLabel: "gauge"},
			Value:     1,
			Timestamp: model.TimeFromUnix(1700000000),
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	}))
	t.Cleanup(srv.Close)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Queries = []Query{{Query: "sum(up)", MetricNameFallback: "custom_name"}}
	cfg.ClientConfig.Endpoint = srv.URL
	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = zaptest.NewLogger(t)
	s := &scraper{queries: cfg.Queries, cfg: cfg, settings: settings}
	require.NoError(t, s.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, s.Shutdown(t.Context())) })

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)
	require.Equal(t, "custom_name", metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestScrapeMetricsMissingMetricNameFallsBackToDefault(t *testing.T) {
	v := model.Vector{
		&model.Sample{
			Metric:    model.Metric{model.MetricTypeLabel: "gauge"},
			Value:     1,
			Timestamp: model.TimeFromUnix(1700000000),
		},
	}
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)
	require.Equal(t, defaultMetricName, metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestScrapeMetricsNoneResultType(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"none","result":null}}`))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)
	require.Equal(t, 0, metrics.MetricCount())
}

func TestScrapeMetricsStringAndScalarResultTypesAreLoggedNotErrored(t *testing.T) {
	for _, resultType := range []parser.ValueType{parser.ValueTypeString, parser.ValueTypeScalar} {
		t.Run(string(resultType), func(t *testing.T) {
			s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `{"status":"success","data":{"resultType":%q,"result":[0,"1"]}}`, resultType)
			})

			metrics, err := s.ScrapeMetrics(t.Context())
			require.NoError(t, err)
			require.Equal(t, 0, metrics.MetricCount())
		})
	}
}

func TestScrapeMetricsUnsupportedResultTypeIsLoggedNotErrored(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"unknown","result":null}}`))
	})

	metrics, err := s.ScrapeMetrics(t.Context())
	require.NoError(t, err)
	require.Equal(t, 0, metrics.MetricCount())
}

func TestScrapeMetricsAPIErrorStatus(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","errorType":"bad_data","error":"invalid query"}`))
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.ErrorContains(t, err, "invalid query")
}

func TestScrapeMetricsMalformedResponseBody(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsMalformedResultBody(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":"not an array"}}`))
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsMalformedMatrixResultBody(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":"not an array"}}`))
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsMalformedDataEnvelope(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":"not an object"}`))
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Queries = []Query{{Query: "up"}}
	cfg.ClientConfig.Endpoint = srv.URL
	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = zaptest.NewLogger(t)
	s := &scraper{queries: cfg.Queries, cfg: cfg, settings: settings}
	require.NoError(t, s.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, s.Shutdown(t.Context())) })

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsHTTPServerError(t *testing.T) {
	s := newTestScraper(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := s.ScrapeMetrics(t.Context())
	require.Error(t, err)
}

func TestScrapeMetricsMultipleQueriesAggregateErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") == "bad" {
			_, _ = w.Write([]byte(`{"status":"error","error":"boom"}`))
			return
		}
		v := model.Vector{&model.Sample{Metric: model.Metric{model.MetricNameLabel: "up", model.MetricTypeLabel: "gauge"}, Value: 1, Timestamp: model.TimeFromUnix(1700000000)}}
		_, _ = w.Write(promAPIResponse(t, parser.ValueTypeVector, v))
	}))
	t.Cleanup(srv.Close)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Queries = []Query{{Query: "up"}, {Query: "bad"}}
	cfg.ClientConfig.Endpoint = srv.URL
	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = zaptest.NewLogger(t)
	s := &scraper{queries: cfg.Queries, cfg: cfg, settings: settings}
	require.NoError(t, s.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, s.Shutdown(t.Context())) })

	metrics, err := s.ScrapeMetrics(t.Context())
	require.ErrorContains(t, err, "boom")
	require.Equal(t, 1, metrics.MetricCount())
}

func TestScrapeMetricsInvalidEndpointURL(t *testing.T) {
	settings := receivertest.NewNopSettings(component.MustNewType("promql"))
	settings.TelemetrySettings.Logger = zaptest.NewLogger(t)
	cfg := &Config{Queries: []Query{{Query: "up"}}}
	cfg.ClientConfig.Endpoint = "http://[::1]:namedport"
	s := &scraper{queries: cfg.Queries, cfg: cfg, settings: settings}
	require.NoError(t, s.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, s.Shutdown(context.Background())) })

	_, err := s.ScrapeMetrics(context.Background())
	require.Error(t, err)
}
