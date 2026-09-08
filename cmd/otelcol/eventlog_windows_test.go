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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service/telemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWindowsEventLogCoreMapsLevelsAndFields(t *testing.T) {
	elog := &recordingWindowsEventLog{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(io.Discard),
		zapcore.DebugLevel,
	)
	logger := zap.New(withWindowsEventLogCore(elog)(core))

	logger.Debug("debug message")
	logger.With(zap.String("persistent", "field")).Info("info message", zap.String("inline", "value"))
	logger.Warn("warning message")
	logger.Error("error message")
	logger.DPanic("dpanic message")

	require.Len(t, elog.events, 5)
	expected := []eventLogEntry{
		{level: eventLogInfo, eid: 1},
		{level: eventLogInfo, eid: 1},
		{level: eventLogWarning, eid: 2},
		{level: eventLogError, eid: 3},
		{level: eventLogError, eid: 3},
	}
	for i := range expected {
		assert.Equal(t, expected[i].level, elog.events[i].level)
		assert.Equal(t, expected[i].eid, elog.events[i].eid)
	}
	assert.Contains(t, elog.events[1].message, "info message")
	assert.Contains(t, elog.events[1].message, "persistent")
	assert.Contains(t, elog.events[1].message, "field")
	assert.Contains(t, elog.events[1].message, "inline")
	assert.Contains(t, elog.events[1].message, "value")
}

func TestWindowsEventLogCorePreservesConfiguredLevel(t *testing.T) {
	elog := &recordingWindowsEventLog{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(io.Discard),
		zapcore.WarnLevel,
	)
	logger := zap.New(withWindowsEventLogCore(elog)(core))

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")

	require.Len(t, elog.events, 1)
	assert.Equal(t, eventLogWarning, elog.events[0].level)
	assert.Contains(t, elog.events[0].message, "warning message")
}

func TestWindowsEventLogTelemetryFactoryWrapsOnlyConsoleOutputs(t *testing.T) {
	tests := []struct {
		name                string
		outputs             []string
		expectedOptionCount int
	}{
		{
			name:                "default stderr output",
			outputs:             []string{"stderr"},
			expectedOptionCount: 1,
		},
		{
			name:                "stdout and stderr outputs",
			outputs:             []string{"stdout", "stderr"},
			expectedOptionCount: 1,
		},
		{
			name:    "file output",
			outputs: []string{`C:\logs\otelcol.log`},
		},
		{
			name:    "console and file outputs",
			outputs: []string{"stderr", `C:\logs\otelcol.log`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapCfg := zap.NewProductionConfig()
			zapCfg.OutputPaths = tt.outputs
			factory := telemetry.NewFactory(
				func() component.Config { return struct{}{} },
				telemetry.WithCreateLogger(func(_ context.Context, settings telemetry.LoggerSettings, _ component.Config) (*zap.Logger, component.ShutdownFunc, error) {
					logger, err := settings.BuildZapLogger(zapCfg)
					return logger, nil, err
				}),
			)

			var optionCount int
			wrappedFactory := windowsEventLogTelemetryFactory{
				Factory: factory,
				elog:    &recordingWindowsEventLog{},
			}
			logger, _, err := wrappedFactory.CreateLogger(t.Context(), telemetry.LoggerSettings{
				BuildZapLogger: func(_ zap.Config, opts ...zap.Option) (*zap.Logger, error) {
					optionCount = len(opts)
					return zap.NewNop(), nil
				},
			}, struct{}{})

			require.NoError(t, err)
			require.NotNil(t, logger)
			assert.Equal(t, tt.expectedOptionCount, optionCount)
		})
	}
}

type eventLogLevel string

const (
	eventLogInfo    eventLogLevel = "info"
	eventLogWarning eventLogLevel = "warning"
	eventLogError   eventLogLevel = "error"
)

type eventLogEntry struct {
	level   eventLogLevel
	message string
	eid     uint32
}

type recordingWindowsEventLog struct {
	events []eventLogEntry
}

func (r *recordingWindowsEventLog) Info(eid uint32, msg string) error {
	r.events = append(r.events, eventLogEntry{level: eventLogInfo, eid: eid, message: msg})
	return nil
}

func (r *recordingWindowsEventLog) Warning(eid uint32, msg string) error {
	r.events = append(r.events, eventLogEntry{level: eventLogWarning, eid: eid, message: msg})
	return nil
}

func (r *recordingWindowsEventLog) Error(eid uint32, msg string) error {
	r.events = append(r.events, eventLogEntry{level: eventLogError, eid: eid, message: msg})
	return nil
}
