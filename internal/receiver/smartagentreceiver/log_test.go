// Copyright Splunk, Inc.
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

package smartagentreceiver

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func logAt(entry *logrus.Entry, level logrus.Level, msg string) {
	entry.Logger.Level = level
	switch level {
	case logrus.FatalLevel:
		entry.Fatal(msg)
	case logrus.PanicLevel:
		entry.Panic(msg)
	case logrus.ErrorLevel:
		entry.Error(msg)
	case logrus.WarnLevel:
		entry.Warn(msg)
	case logrus.InfoLevel:
		entry.Info(msg)
	case logrus.DebugLevel:
		entry.Debug(msg)
	case logrus.TraceLevel:
		entry.Trace(msg)
	}
}

func TestRedirectMonitorLogs(t *testing.T) {
	for logrusLevel, zapLevel := range logrusToZapLevel {
		//TODO: handle fatal and panic levels
		if logrusLevel == logrus.FatalLevel || logrusLevel == logrus.PanicLevel {
			continue
		}

		// Simulating the creation of logrus entry/loggers in monitors (monitor1, monitor2).
		monitor1LogrusLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})
		monitor2LogrusLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor2"})

		// Case where logs do not have "monitorType" field
		monitor3LogrusLogger := logrus.WithFields(logrus.Fields{})

		// Checking that the logrus standard logger is the logger in the monitors.
		require.Same(t, logrus.StandardLogger(), monitor1LogrusLogger.Logger, "Expected the standard logrus logger")
		require.Same(t, logrus.StandardLogger(), monitor2LogrusLogger.Logger, "Expected the standard logrus logger")
		require.Same(t, logrus.StandardLogger(), monitor3LogrusLogger.Logger, "Expected the standard logrus logger")

		// Simulating the creation of logrus keys for the monitors where:
		// 1. the monitor types (i.e. monitor1, monitor2) are known.
		// 2. the logger is assumed to be the standard logrus logger.
		monitor1LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}
		monitor2LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor2"}
		monitor3LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor3"}

		monitor1ZapLogger, monitor1ZapLogs := newObservedLogs(zapLevel)
		monitor2ZapLogger, monitor2ZapLogs := newObservedLogs(zapLevel)
		monitor3ZapLogger, monitor3ZapLogs := newObservedLogs(zapLevel)
		defaultZapLogger, defaultLogs := newObservedLogs(zapLevel)

		// Logrus to zap redirections.
		logToZap := newLogrusToZap(func() *zap.Logger { return defaultZapLogger })
		logToZap.redirect(monitor1LogrusKey, monitor1ZapLogger)
		logToZap.redirect(monitor2LogrusKey, monitor2ZapLogger)
		logToZap.redirect(monitor3LogrusKey, monitor3ZapLogger)

		expectedLogrusLevel := logrusLevel
		if logrusLevel == logrus.TraceLevel {
			expectedLogrusLevel = logrus.DebugLevel
		}
		require.Equal(t, expectedLogrusLevel, monitor1LogrusKey.Level)
		require.Equal(t, expectedLogrusLevel, monitor2LogrusKey.Level)
		require.Equal(t, expectedLogrusLevel, monitor3LogrusKey.Level)

		logMsg1 := "a log msg1"
		logMsg2 := "a log msg2"
		logMsg3 := "a log msg3"

		// Logrus logging.
		logAt(monitor1LogrusLogger, logrusLevel, logMsg1)
		logAt(monitor2LogrusLogger, logrusLevel, logMsg2)
		logAt(monitor3LogrusLogger, logrusLevel, logMsg3)

		// Checking zap logs for redirected logrus logs.
		require.Equal(t, 1, monitor1ZapLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", monitor1ZapLogs.Len()))
		require.Equal(t, logMsg1, monitor1ZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg1, monitor1ZapLogs.All()[0].Message))

		require.Equal(t, 1, monitor2ZapLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", monitor2ZapLogs.Len()))
		require.Equal(t, logMsg2, monitor2ZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg2, monitor2ZapLogs.All()[0].Message))

		require.Equal(t, 0, monitor3ZapLogs.Len(), fmt.Sprintf("Expected 0 log message, got %d", monitor3ZapLogs.Len()))
		require.Equal(t, logMsg3, defaultLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg3, defaultLogs.All()[0].Message))

		require.Equal(t, 1, defaultLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", defaultLogs.Len()))

		logToZap.unRedirect(monitor1LogrusKey, monitor1ZapLogger)
		logToZap.unRedirect(monitor2LogrusKey, monitor2ZapLogger)
		logToZap.unRedirect(monitor3LogrusKey, monitor3ZapLogger)
	}
}

func TestRedirectMonitorLogsWithNilLoggerMap(t *testing.T) {
	for logrusLevel, zapLevel := range logrusToZapLevel {
		//TODO: handle fatal and panic levels
		if logrusLevel == logrus.FatalLevel || logrusLevel == logrus.PanicLevel {
			continue
		}

		// Simulating the creation of logrus entry/logger in monitor1.
		monitor1LogrusLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

		// Simulating the creation of logrus key for monitor1 in the smart agent receiver where:
		// 1. the monitor types (i.e. monitor1) are known.
		// 2. the logger is assumed to be the standard logrus logger.
		monitor1LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

		monitor1ZapLogger, monitor1ZapLogs := newObservedLogs(zapLevel)
		defaultZapLogger, defaultZapLogs := newObservedLogs(zapLevel)

		// Logrus to zap redirection.
		logToZap := newLogrusToZap(func() *zap.Logger { return defaultZapLogger })
		logToZap.redirect(monitor1LogrusKey, monitor1ZapLogger)

		require.Len(t, logToZap.loggerMap, 1)
		require.Len(t, logToZap.loggerMap[monitor1LogrusKey], 1)
		require.Equal(t, monitor1ZapLogger, logToZap.loggerMap[monitor1LogrusKey][0])

		// Setting loggerMap to nil
		logToZap.loggerMap = nil

		logMsg := "a log msg"

		// Logrus logging.
		logAt(monitor1LogrusLogger, logrusLevel, logMsg)

		// Logrus logs should not be redirected to monitor1ZapLogger.
		assert.Equal(t, 0, monitor1ZapLogs.Len(), fmt.Sprintf("Expected 0 log message, got %d", monitor1ZapLogs.Len()))

		// Logrus logs should be redirected to defaultZapLogger when loggerMap is nil.
		assert.Equal(t, 1, defaultZapLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", monitor1ZapLogs.Len()))
		require.Equal(t, logMsg, defaultZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg, defaultZapLogs.All()[0].Message))

		logToZap.unRedirect(monitor1LogrusKey, monitor1ZapLogger)
	}
}

func TestRedirectSameMonitorManyInstancesLogs(t *testing.T) {
	for logrusLevel, zapLevel := range logrusToZapLevel {
		//TODO: handle fatal and panic levels
		if logrusLevel == logrus.FatalLevel || logrusLevel == logrus.PanicLevel {
			continue
		}

		// Simulating the creation of logrus entry/loggers in instances of a monitor (monitor1).
		instance1LogrusLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})
		instance2LogrusLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

		// Simulating the creation of logrus keys for the instances of monitor1 in the smart agent receiver where:
		// 1. the monitor type (i.e. monitor1) is known.
		// 2. the logger is assumed to be the standard logrus logger.
		instance1LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}
		instance2LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

		// Checking that the logrus standard logger is the logger for both monitor1 instances.
		require.Same(t, logrus.StandardLogger(), instance1LogrusLogger.Logger, "Expected the standard logrus logger")
		require.Same(t, logrus.StandardLogger(), instance2LogrusLogger.Logger, "Expected the standard logrus logger")
		// Checking that the logrus keys are equal.
		require.Equal(t, instance1LogrusKey, instance2LogrusKey, "Expected the standard logrus logger")

		instance1ZapLogger, instance1ZapLogs := newObservedLogs(zapLevel)
		instance2ZapLogger, instance2ZapLogs := newObservedLogs(zapLevel)

		// Simulating logrus to zap redirections in in the smart agent receiver.
		logToZap := newLogrusToZap(zap.NewNop)
		logToZap.redirect(instance1LogrusKey, instance1ZapLogger)
		logToZap.redirect(instance2LogrusKey, instance2ZapLogger)

		logMsg1 := "a log msg1"
		logMsg2 := "a log msg2"

		// Simulating logging messages in the instances of monitor1.
		logAt(instance1LogrusLogger, logrusLevel, logMsg1)
		logAt(instance2LogrusLogger, logrusLevel, logMsg2)

		// Asserting that messages logged by all instances get logged by the zap logger of the first instance.
		require.Equal(t, 2, instance1ZapLogs.Len(), fmt.Sprintf("Expected 2 log message, got %d", instance1ZapLogs.Len()))
		// Asserting that no messages get logged by the zap logger of the other instances.
		require.Equal(t, 0, instance2ZapLogs.Len(), fmt.Sprintf("Expected 0 log message, got %d", instance2ZapLogs.Len()))

		// Asserting messages logged by the zap logger of the first instance.
		require.Equal(t, logMsg1, instance1ZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg1, instance1ZapLogs.All()[0].Message))
		require.Equal(t, logMsg2, instance1ZapLogs.All()[1].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg2, instance1ZapLogs.All()[1].Message))

		logToZap.unRedirect(instance1LogrusKey, instance1ZapLogger)
		logToZap.unRedirect(instance2LogrusKey, instance2ZapLogger)
	}
}

func TestLevelsMapShouldIncludeAllLogrusLevels(t *testing.T) {
	for _, level := range logrus.AllLevels {
		_, ok := logrusToZapLevel[level]
		require.True(t, ok, fmt.Sprintf("Expected log level %s not found", level.String()))
	}
}

func TestLevelsShouldReturnAllLogrusLevels(t *testing.T) {
	hook := newLogrusToZap(zap.NewNop)
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		require.Equal(t, logrus.AllLevels[i], levels[i], fmt.Sprintf("Expected log level %s not found", logrus.AllLevels[i].String()))
	}
}

func TestRedirectShouldSetLogrusKeyLoggerReportCallerTrue(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	dst := zap.NewNop()
	logToZap := newLogrusToZap(zap.NewNop)
	logToZap.redirect(src, dst)
	require.True(t, src.ReportCaller, "Expected the logrus key logger to report caller")
	logToZap.unRedirect(src, dst)
}

func TestRedirectShouldUniquelyAddHooksToLogrusKeyLogger(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	require.Equal(t, 0, len(src.Hooks), fmt.Sprintf("Expected 0 hooks, got %d", len(src.Hooks)))

	logToZap := newLogrusToZap(zap.NewNop)
	// These multiple redirect calls should add hook once to log levels.
	logToZap.redirect(src, zap.NewNop())
	logToZap.redirect(src, zap.NewNop())
	logToZap.redirect(src, zap.NewNop())

	for _, level := range logrus.AllLevels {
		got := len(src.Hooks[level])
		require.Equal(t, 1, got, fmt.Sprintf("Expected 1 hook for log level %s, got %d", level.String(), got))
		require.Equal(t, logToZap, src.Hooks[level][0], fmt.Sprintf("Expected hook for log level %s not found", level.String()))
	}
}

func TestLogrusToZapLevel(t *testing.T) {
	tests := []struct {
		name           string
		zapConfig      *zap.Config
		levelsEnabled  []zapcore.Level
		levelsDisabled []zapcore.Level
	}{
		{
			name: "debug",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.DebugLevel),
			},
			levelsEnabled: zapLevels,
		},
		{
			name: "info",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.InfoLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
				zapcore.DPanicLevel,
				zapcore.PanicLevel,
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.DebugLevel,
			},
		},
		{
			name: "warn",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.WarnLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
				zapcore.DPanicLevel,
				zapcore.PanicLevel,
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.InfoLevel,
				zapcore.DebugLevel,
			},
		},
		{
			name: "error",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.ErrorLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.ErrorLevel,
				zapcore.DPanicLevel,
				zapcore.PanicLevel,
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.DebugLevel,
			},
		},
		{
			name: "dpanic",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.DPanicLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.DPanicLevel,
				zapcore.PanicLevel,
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.DebugLevel,
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
			},
		},
		{
			name: "panic",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.PanicLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.PanicLevel,
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.DebugLevel,
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
				zapcore.DPanicLevel,
			},
		},
		{
			name: "fatal",
			zapConfig: &zap.Config{
				Level: zap.NewAtomicLevelAt(zap.FatalLevel),
			},
			levelsEnabled: []zapcore.Level{
				zapcore.FatalLevel,
			},
			levelsDisabled: []zapcore.Level{
				zapcore.DebugLevel,
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
				zapcore.PanicLevel,
				zapcore.DPanicLevel,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.zapConfig.Encoding = "json"
			logger, err := test.zapConfig.Build()
			require.NoError(t, err)

			logToZap := newLogrusToZap(loggerProvider(logger.Core()))
			defaultLoggerCore := logToZap.defaultLogger.Core()

			for _, level := range test.levelsEnabled {
				require.True(t, defaultLoggerCore.Enabled(level), fmt.Sprintf("level:%s", level))
			}

			for _, level := range test.levelsDisabled {
				require.False(t, defaultLoggerCore.Enabled(level), fmt.Sprintf("level:%s", level))
			}
		})
	}
}

func newObservedLogs(level zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	return zap.New(core), logs
}
