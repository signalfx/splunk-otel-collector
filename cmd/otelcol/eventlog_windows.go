// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package main

import (
	"context"
	"fmt"
	"slices"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/service/telemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc/eventlog"
)

const windowsEventLogSource = "splunk-otel-collector"

// runInteractiveWithWindowsEventLog restores the Windows service logging behavior,
// such as event log level, when the Collector is running as a child of launcher
func runInteractiveWithWindowsEventLog(settings otelcol.CollectorSettings) error {
	elog, err := eventlog.Open(windowsEventLogSource)
	if err != nil {
		return fmt.Errorf("failed to open Windows Event Log source %q: %w", windowsEventLogSource, err)
	}
	defer func() {
		_ = elog.Close()
	}()

	getFactories := settings.Factories
	settings.Factories = func() (otelcol.Factories, error) {
		factories, factoriesErr := getFactories()
		if factoriesErr != nil {
			return factories, factoriesErr
		}
		baseTelemetryFactory := factories.Telemetry
		factories.Telemetry = windowsEventLogTelemetryFactory{
			Factory: baseTelemetryFactory,
			elog:    elog,
		}
		return factories, nil
	}

	return runInteractive(settings)
}

// windowsEventLogTelemetryFactory overrides only final logger construction and
// delegates all other telemetry behavior to the configured factory.
type windowsEventLogTelemetryFactory struct {
	telemetry.Factory
	elog windowsEventLog
}

// CreateLogger wraps the Zap logger builder to mirror the upstream log handling
// where the logger writes to the Windows Event Log if no file output is specified.
func (f windowsEventLogTelemetryFactory) CreateLogger(
	ctx context.Context,
	settings telemetry.LoggerSettings,
	cfg component.Config,
) (*zap.Logger, component.ShutdownFunc, error) {
	buildZapLogger := settings.BuildZapLogger
	if buildZapLogger == nil {
		buildZapLogger = zap.Config.Build
	}
	settings.BuildZapLogger = func(zapCfg zap.Config, opts ...zap.Option) (*zap.Logger, error) {
		for _, output := range zapCfg.OutputPaths {
			if output != "stdout" && output != "stderr" {
				return buildZapLogger(zapCfg, opts...)
			}
		}
		opts = slices.Insert(opts, 0, zap.WrapCore(withWindowsEventLogCore(f.elog)))
		return buildZapLogger(zapCfg, opts...)
	}
	return f.Factory.CreateLogger(ctx, settings, cfg)
}

type windowsEventLog interface {
	Info(eid uint32, msg string) error
	Warning(eid uint32, msg string) error
	Error(eid uint32, msg string) error
}

var _ zapcore.Core = (*windowsEventLogCore)(nil)

// windowsEventLogCore mirrors the Windows Event Log core in the upstream
// OpenTelemetry Collector's otelcol/collector_windows.go.
type windowsEventLogCore struct {
	core    zapcore.Core
	elog    windowsEventLog
	encoder zapcore.Encoder
}

func (w windowsEventLogCore) Enabled(level zapcore.Level) bool {
	return w.core.Enabled(level)
}

func (w windowsEventLogCore) With(fields []zapcore.Field) zapcore.Core {
	enc := w.encoder.Clone()
	for _, field := range fields {
		field.AddTo(enc)
	}
	return windowsEventLogCore{
		core:    w.core,
		elog:    w.elog,
		encoder: enc,
	}
}

func (w windowsEventLogCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if w.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, w)
	}
	return checkedEntry
}

func (w windowsEventLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buf, err := w.encoder.EncodeEntry(entry, fields)
	if err != nil {
		_ = w.elog.Warning(2, fmt.Sprintf("failed encoding log entry %v\r\n", err))
		return err
	}
	msg := buf.String()
	buf.Free()

	switch entry.Level {
	case zapcore.FatalLevel, zapcore.PanicLevel, zapcore.DPanicLevel, zapcore.ErrorLevel:
		// golang.org/x/sys/windows/svc/eventlog does not support Critical events.
		return w.elog.Error(3, msg)
	case zapcore.WarnLevel:
		return w.elog.Warning(2, msg)
	case zapcore.InfoLevel:
		return w.elog.Info(1, msg)
	}
	return w.elog.Info(1, msg)
}

func (w windowsEventLogCore) Sync() error {
	return w.core.Sync()
}

// withWindowsEventLogCore replaces the configured console core with the same
// Event Log encoding and severity mapping used by the upstream service handler.
func withWindowsEventLogCore(elog windowsEventLog) func(zapcore.Core) zapcore.Core {
	return func(core zapcore.Core) zapcore.Core {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.LineEnding = "\r\n"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return windowsEventLogCore{core: core, elog: elog, encoder: zapcore.NewConsoleEncoder(encoderConfig)}
	}
}
