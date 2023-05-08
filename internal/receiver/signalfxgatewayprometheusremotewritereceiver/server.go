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

package signalfxgatewayprometheusremotewritereceiver

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/storage/remote"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type prometheusRemoteWriteServer struct {
	*http.Server
	*serverConfig
	closeChannel *sync.Once
}

type serverConfig struct {
	Reporter reporter
	component.Host
	Mc chan<- pmetric.Metrics
	component.TelemetrySettings
	Path   string
	Parser *prometheusRemoteOtelParser
	confighttp.HTTPServerSettings
}

func newPrometheusRemoteWriteServer(config *serverConfig) (*prometheusRemoteWriteServer, error) {
	mx := mux.NewRouter()
	handler := newHandler(config.Parser, config, config.Mc)
	mx.HandleFunc(config.Path, handler)
	mx.Host(config.Endpoint)
	// TODO is this really where it's thrown?  oh prolly in handler... .wait though lots of nils
	// otelcollector             | go.opentelemetry.io/collector/config/confighttp.(*HTTPServerSettings).ToServer(0xc000dd2d80, {0x61fe798, 0xc002437730}, {0x0, {0x0, 0x0}, {0x0, 0x0}, 0x0}, {0x61c0fa0, ...}, ...)
	server, err := config.HTTPServerSettings.ToServer(config.Host, config.TelemetrySettings, mx)
	// Currently this is not set, in favor of the pattern where they always explicitly pass the listener
	server.Addr = config.Endpoint
	if err != nil {
		return nil, err
	}
	return &prometheusRemoteWriteServer{
		Server:       server,
		serverConfig: config,
		closeChannel: &sync.Once{},
	}, nil
}

func (prw *prometheusRemoteWriteServer) close() error {
	defer prw.closeChannel.Do(func() { close(prw.Mc) })
	return prw.Server.Close()
}

func (prw *prometheusRemoteWriteServer) listenAndServe() error {
	prw.Reporter.OnDebugf("Starting prometheus simple write server")
	listener, err := prw.serverConfig.ToListener()
	if err != nil {
		return err
	}
	defer listener.Close()
	err = prw.Server.Serve(listener)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func newHandler(parser *prometheusRemoteOtelParser, sc *serverConfig, mc chan<- pmetric.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc.Reporter.OnDebugf("Processing write request %s", r.RequestURI)
		req, err := remote.DecodeWriteRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(req.Timeseries) == 0 && len(req.Metadata) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		results, err := parser.fromPrometheusWriteRequestMetrics(req)
		if nil != err {
			http.Error(w, err.Error(), http.StatusBadRequest)
			sc.Reporter.OnDebugf("prometheus_translation", err)
			return
		}
		mc <- results
		w.WriteHeader(http.StatusAccepted)
	}
}
