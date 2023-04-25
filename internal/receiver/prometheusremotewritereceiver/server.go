// Copyright 2020, OpenTelemetry Authors
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

package prometheusremotewritereceiver

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type PrometheusRemoteWriteServer struct {
	*http.Server
	*ServerConfig
}

type ServerConfig struct {
	Reporter
	component.Host
	Mc chan pmetric.Metrics
	component.TelemetrySettings
	Path string
	confighttp.HTTPServerSettings
}

func NewPrometheusRemoteWriteServer(ctx context.Context, config *ServerConfig) (*PrometheusRemoteWriteServer, error) {
	mx := mux.NewRouter()
	handler := newHandler(ctx, config.Reporter, config, config.Mc)
	mx.HandleFunc(config.Path, handler)
	mx.Host(config.Endpoint)
	server, err := config.HTTPServerSettings.ToServer(config.Host, config.TelemetrySettings, handler)
	// Currently this is not set, in favor of the pattern where they always explicitly pass the listener
	server.Addr = config.Endpoint
	if err != nil {
		return nil, err
	}
	return &PrometheusRemoteWriteServer{
		Server:       server,
		ServerConfig: config,
	}, nil
}

func (prw *PrometheusRemoteWriteServer) Close() error {
	return prw.Server.Close()
}

func (prw *PrometheusRemoteWriteServer) ListenAndServe() error {
	prw.Reporter.OnDebugf("Starting prometheus simple write server")
	listener, err := prw.ServerConfig.ToListener()
	if err != nil {
		return err
	}
	defer func() {
		if e := listener.Close(); e != nil {
			prw.Reporter.OnDebugf("error in listener: %s", e)
		}
	}()
	err = prw.Server.Serve(listener)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func newHandler(ctx context.Context, reporter Reporter, _ *ServerConfig, _ chan pmetric.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// THIS IS A STUB FUNCTION.  You can see another branch with how I'm thinking this will look if you're curious
		ctx2 := reporter.StartMetricsOp(ctx)
		reporter.OnMetricsProcessed(ctx2, 0, nil)
		w.WriteHeader(http.StatusNoContent)
	}
}
