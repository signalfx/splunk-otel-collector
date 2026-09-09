// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

// Program splunk-connect-for-otlp is a binary listening for OTLP data and exporting it to stdout.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/signalfx/splunk-otel-collector/internal/auth"
	"github.com/signalfx/splunk-otel-collector/internal/exporter/stdoutexporter"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--scheme":
			fmt.Println(Scheme)
		case "--validate-arguments":
		}
	} else if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			fmt.Println("Stack trace:")
			fmt.Println(string(debug.Stack()))
		}
	}()
	config, errCfg := ReadFromStdin()
	if errCfg != nil {
		return errCfg
	}

	logger, errLogger := createLogger()
	if errLogger != nil {
		return errLogger
	}
	logger.Info("Starting OTLP input")

	settings := component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: noop.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		Resource:       pcommon.NewResource(),
	}

	xmlCfg := config.Extract()
	if err := xmlCfg.Validate(); err != nil {
		return fmt.Errorf("cannot start TA due to invalid configuration: %w", err)
	}
	stdoutCfg := stdoutexporter.NewFactory().CreateDefaultConfig().(*stdoutexporter.Config)
	stdoutCfg.Source = xmlCfg.Source
	stdoutCfg.Sourcetype = xmlCfg.Sourcetype

	f := stdoutexporter.NewFactory()
	ctx := context.Background()
	telemetrySettings := exporter.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("stdout"),
	}
	var le exporter.Logs
	var me exporter.Metrics
	var tracesExporter exporter.Traces
	var err error
	if le, err = f.CreateLogs(ctx, telemetrySettings, stdoutCfg); err != nil {
		return err
	}
	if me, err = f.CreateMetrics(ctx, telemetrySettings, stdoutCfg); err != nil {
		return err
	}
	if tracesExporter, err = f.CreateTraces(ctx, telemetrySettings, stdoutCfg); err != nil {
		return err
	}
	logger.Info("Configured exporter")

	rf := otlpreceiver.NewFactory()
	cfg := rf.CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint = fmt.Sprintf("%s:%d", xmlCfg.ListenAddress, xmlCfg.GrpcPort)
	cfg.Protocols.HTTP.GetOrInsertDefault().ServerConfig.NetAddr.Endpoint = fmt.Sprintf("%s:%d", xmlCfg.ListenAddress, xmlCfg.HTTPPort)
	extID := component.MustNewID("splunkauth")
	cfg.Protocols.GRPC.Get().Auth.GetOrInsertDefault().AuthenticatorID = extID
	cfg.Protocols.HTTP.Get().ServerConfig.Auth.GetOrInsertDefault().Config.AuthenticatorID = extID

	if xmlCfg.EnableSSL {
		cfg.Protocols.GRPC.Get().TLS.GetOrInsertDefault().CertFile = xmlCfg.ServerCert
		cfg.Protocols.HTTP.Get().ServerConfig.TLS.GetOrInsertDefault().CertFile = xmlCfg.ServerCert
		cfg.Protocols.GRPC.Get().TLS.GetOrInsertDefault().KeyFile = xmlCfg.ServerKey
		cfg.Protocols.HTTP.Get().ServerConfig.TLS.GetOrInsertDefault().KeyFile = xmlCfg.ServerKey
	}

	otlpSettings := receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}
	if _, err = rf.CreateLogs(ctx, otlpSettings, cfg, le); err != nil {
		return err
	}
	if _, err = rf.CreateMetrics(ctx, otlpSettings, cfg, me); err != nil {
		return err
	}
	var r receiver.Traces
	r, err = rf.CreateTraces(ctx, otlpSettings, cfg, tracesExporter)
	if err != nil {
		return err
	}

	logger.Info("Configured OTLP receiver")

	h := newTtyHost(map[component.ID]component.Component{})
	h.Start()

	var authExtension extension.Extension
	if authExtension, err = auth.New(ctx, settings, xmlCfg.ServerURI, xmlCfg.SessionKey); err != nil {
		return err
	}

	h.Extensions[extID] = authExtension

	if errStart := errors.Join(le.Start(ctx, h), me.Start(ctx, h), tracesExporter.Start(ctx, h), r.Start(ctx, h)); errStart != nil {
		return errStart
	}

	logger.Info("OTLP Input started")

	err = h.Wait()

	return errors.Join(err, r.Shutdown(ctx), le.Shutdown(ctx), tracesExporter.Shutdown(ctx), me.Shutdown(ctx))
}

func createLogger() (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapCfg.OutputPaths = []string{"stderr"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}
	return zapCfg.Build()
}
