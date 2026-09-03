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
	"log/slog"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
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
	ref, err := appender.Append(storage.SeriesRef(0), labels.New(labels.Label{Name: "foo", Value: "bar"}, labels.Label{Name: model.MetricNameLabel, Value: "up"}, labels.Label{Name: model.MetricTypeLabel, Value: string(model.MetricTypeGauge)}), time.Now().UnixMilli(), time.Now().Unix(), 1, nil, nil, storage.AppendV2Options{
		MetricFamilyName: "up",
		Metadata:         metadata.Metadata{Type: model.MetricTypeGauge},
	})
	require.NoError(t, err)
	require.NoError(t, appender.Commit())
	appender = db.AppenderV2(t.Context())
	_, err = appender.Append(ref, labels.New(labels.Label{Name: "foo", Value: "bar"}, labels.Label{Name: model.MetricNameLabel, Value: "up"}, labels.Label{Name: model.MetricTypeLabel, Value: string(model.MetricTypeGauge)}), time.Now().Unix(), time.Now().UnixMilli(), 1, nil, nil, storage.AppendV2Options{
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
	cfg.Queries = []string{"up"}
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
}
