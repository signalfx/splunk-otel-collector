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

package internal

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
	Reporter Reporter
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
	handler := newHandler(ctx, config.Reporter, config, config.Mc)
	mx := mux.NewRouter()
	mx.HandleFunc(config.Path, handler)
	server, err := config.HTTPServerSettings.ToServer(config.Host, config.TelemetrySettings, handler)
	if err != nil {
		return nil, err
	}
	return &PrometheusRemoteWriteServer{
		Server:   server,
		Reporter: config.Reporter,
	}, nil
}

func (prw *PrometheusRemoteWriteServer) Close() error {
	return prw.Server.Close()
}

func (prw *PrometheusRemoteWriteServer) ListenAndServe() error {
	prw.Reporter.OnDebugf("Starting prometheus simple write server")
	err := prw.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func newHandler(_ context.Context, _ Reporter, _ *ServerConfig, _ chan pmetric.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// THIS IS A STUB FUNCTION.  You can see another branch with how I'm thinking this will look if you're curious
		w.WriteHeader(http.StatusBadGateway)
	}
}
