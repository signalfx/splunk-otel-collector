package forwarder

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/signalfx/golib/v3/datapoint/dpsink"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/golib/v3/web"
	"github.com/signalfx/ingest-protocols/protocol/signalfx"
)

type pathSetupFunc = func(*mux.Router, http.Handler)

func (m *Monitor) startListening(ctx context.Context, listenAddr string, timeout time.Duration, sink signalfx.Sink) (sfxclient.Collector, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("cannot open listening address %s: %w", listenAddr, err)
	}
	router := mux.NewRouter()

	httpChain := web.NextConstructor(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next web.ContextHandler) {
		next.ServeHTTPC(tryToExtractRemoteAddressToContext(ctx, r), rw, r)
	})

	jaegerMetrics := m.setupHandler(ctx, router, signalfx.JaegerV1, sink, func(sink signalfx.Sink) signalfx.ErrorReader {
		return signalfx.NewJaegerThriftTraceDecoderV1(m.golibLogger, sink)
	}, httpChain, setupPathFunc(signalfx.SetupThriftByPaths, signalfx.DefaultTracePathV1))

	protobufDatapoints := m.setupHandler(ctx, router, "protobufv2", sink, func(sink signalfx.Sink) signalfx.ErrorReader {
		return &signalfx.ProtobufDecoderV2{Sink: sink, Logger: m.golibLogger}
	}, httpChain, setupPathFunc(signalfx.SetupProtobufV2ByPaths, "/v2/datapoint"))

	jsonDatapoints := m.setupHandler(ctx, router, "jsonv2", sink, func(sink signalfx.Sink) signalfx.ErrorReader {
		return &signalfx.JSONDecoderV2{Sink: sink, Logger: m.golibLogger}
	}, httpChain, setupPathFunc(signalfx.SetupJSONByPaths, "/v2/datapoint"))

	zipkinMetrics := m.setupHandler(ctx, router, signalfx.ZipkinV1, sink, func(sink signalfx.Sink) signalfx.ErrorReader {
		return &signalfx.JSONTraceDecoderV1{Logger: m.golibLogger, Sink: sink}
	}, httpChain, setupPathFuncN(signalfx.SetupJSONByPathsN, signalfx.DefaultTracePathV1, signalfx.ZipkinTracePathV1, signalfx.ZipkinTracePathV2))

	router.NotFoundHandler = http.HandlerFunc(m.notFoundHandler)

	server := http.Server{
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	go func() { _ = server.Serve(listener) }()

	go func() {
		<-ctx.Done()
		err := server.Close()
		if err != nil {
			m.logger.WithError(err).Error("Could not close SignalFx forwarding server")
		}
	}()
	return sfxclient.NewMultiCollector(jsonDatapoints, protobufDatapoints, jaegerMetrics, zipkinMetrics), nil
}

func setupPathFunc(setupFunc func(*mux.Router, http.Handler, string), path string) pathSetupFunc {
	return func(r *mux.Router, h http.Handler) {
		setupFunc(r, h, path)
	}
}

func setupPathFuncN(setupFunc func(*mux.Router, http.Handler, ...string), paths ...string) pathSetupFunc {
	return func(r *mux.Router, h http.Handler) {
		setupFunc(r, h, paths...)
	}
}

func (m *Monitor) setupHandler(ctx context.Context, router *mux.Router, chainType string, sink signalfx.Sink, getReader func(signalfx.Sink) signalfx.ErrorReader, httpChain web.NextConstructor, pathSetup pathSetupFunc) sfxclient.Collector {
	handler, internalMetrics := signalfx.SetupChain(ctx, sink, chainType, getReader, httpChain, m.golibLogger, &dpsink.Counter{})
	pathSetup(router, handler)
	return internalMetrics
}

func (m *Monitor) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	errMsg := "Datapoint or span request received on invalid path"
	m.logger.ThrottledError(fmt.Sprintf("%s: %s", errMsg, r.URL.Path))

	errMsg = fmt.Sprintf(
		"%s. Supported paths: /v2/datapoint, %s, %s, and %s.\n", errMsg,
		signalfx.DefaultTracePathV1, signalfx.ZipkinTracePathV1, signalfx.ZipkinTracePathV2,
	)

	w.WriteHeader(404)
	_, _ = w.Write([]byte(errMsg))
}
