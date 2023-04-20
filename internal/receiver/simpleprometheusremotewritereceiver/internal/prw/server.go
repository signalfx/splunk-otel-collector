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

package prw

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/simpleprometheusremotewritereceiver/internal/transport"
)

type PrometheusRemoteWriteServer struct {
	*http.Server
	*handler
}

type ServerConfig struct {
	transport.Reporter
	Mc           chan pmetric.Metrics
	Addr         confignet.NetAddr
	Path         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewPrometheusRemoteWriteServer(ctx context.Context, config *ServerConfig) (*PrometheusRemoteWriteServer, error) {
	handler := newHandler(ctx, config.Reporter, config.Path, config.Mc)
	server := &http.Server{
		Handler:      handler,
		Addr:         config.Addr.Endpoint,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}
	return &PrometheusRemoteWriteServer{
		handler: handler,
		Server:  server,
	}, nil
}

func (prw *PrometheusRemoteWriteServer) Close() error {
	return prw.Server.Close()
}

func (prw *PrometheusRemoteWriteServer) ListenAndServe() error {
	prw.reporter.OnDebugf("Starting prometheus simple write server")
	err := prw.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

type handler struct {
	ctx      context.Context
	reporter transport.Reporter
	mc       chan pmetric.Metrics
	path     string
}

func newHandler(ctx context.Context, reporter transport.Reporter, path string, mc chan pmetric.Metrics) *handler {
	return &handler{
		ctx:      ctx,
		path:     path,
		reporter: reporter,
		mc:       mc,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// THIS IS A STUB FUNCTION.  You can see another branch with how I'm thinking this will look if you're curious
	w.WriteHeader(http.StatusBadGateway)
}
