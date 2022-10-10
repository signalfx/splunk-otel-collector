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

// This test confirms that the major assumption of the logrusToZap
// mechanism is accurate for the current logrus dep version.
// If it fails a workaround will need to be determined before adopting the
// breaking version.
func TestAllStdBasedLogrusLoggersWrapStdInstance(t *testing.T) {
	std := logrus.StandardLogger()
	one := logrus.WithField("one", "one")
	two := one.WithField("one", "one")
	three := two.WithField("one", "one")
	require.Same(t, std, one.Logger)
	require.Same(t, std, two.Logger)
	require.Same(t, std, three.Logger)
}

func TestLevelsMapShouldIncludeAllLogrusLevels(t *testing.T) {
	for _, level := range logrus.AllLevels {
		_, ok := logrusToZapLevel[level]
		require.True(t, ok, fmt.Sprintf("Expected log level %q not found", level.String()))
	}
}

func TestLevelsShouldReturnAllLogrusLevels(t *testing.T) {
	hook := newLogrusToZap(zap.NewNop())
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		require.Equal(t, logrus.AllLevels[i], levels[i], fmt.Sprintf("Expected log level %s not found", logrus.AllLevels[i].String()))
	}
}

func TestRedirectShouldSetMonitorLogrusLoggerReportCallerTrue(t *testing.T) {
	defer unredirect()
	src := monitorLogrus{Logger: logrus.New(), monitorType: "monitor1"}
	dst := zap.NewNop()
	logToZap := newLogrusToZap(zap.NewNop())
	logToZap.redirect(src, dst)
	require.True(t, src.ReportCaller, "Expected the logrus key logger to report caller")
}

func TestRedirectShouldUniquelyAddHooksToMonitorLogrusLogger(t *testing.T) {
	unredirect()
	src := monitorLogrus{Logger: logrus.New(), monitorType: "monitor1"}
	require.Equal(t, 0, len(src.Hooks), fmt.Sprintf("Expected 0 hooks, got %d", len(src.Hooks)))

	logToZap := newLogrusToZap(zap.NewNop())
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

			logToZap := newLogrusToZap(logger)
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

func TestRedirectMonitorLogs(t *testing.T) {
	defer setup()()
	for _, lLevel := range logrus.AllLevels {
		logrusLevel := lLevel
		zapLevel := logrusToZapLevel[logrusLevel]
		t.Run(logrusLevel.String(), func(t *testing.T) {
			defer unredirect()
			defaultZap, defaultLogs := newObservedZap(zapLevel)
			logToZap := newLogrusToZap(defaultZap)

			expectedLogrusLevel := logrusLevel
			if logrusLevel == logrus.TraceLevel {
				expectedLogrusLevel = logrus.DebugLevel
			}

			logrus1 := logrus.WithField("monitorType", "monitor1").WithField("monitorID", "id1")
			monitorLogrus1 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "monitor1", monitorID: "id1"}
			zap1, zap1Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus1, zap1)
			require.Equal(t, expectedLogrusLevel, monitorLogrus1.Level)
			require.Same(t, zap1, logToZap.getZapLogger(monitorLogrus1))

			logrus2 := logrus.WithField("monitorType", "monitor2").WithField("monitorID", "id2")
			monitorLogrus2 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "monitor2", monitorID: "id2"}
			zap2, zap2Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus2, zap2)
			require.Equal(t, expectedLogrusLevel, monitorLogrus2.Level)
			require.Same(t, zap2, logToZap.getZapLogger(monitorLogrus2))

			logrus3 := logrus.WithField("irrelevant", "irrelevant")
			monitorLogrus3 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "without-match"}
			zap3, zap3Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus3, zap3)
			require.Equal(t, expectedLogrusLevel, monitorLogrus3.Level)
			require.Same(t, zap3, logToZap.getZapLogger(monitorLogrus3))

			asserter := getAsserter(logrusLevel)
			msg1 := fmt.Sprintf("%v - a log msg", zapLevel.String())
			asserter(t, func() { logAt(logrus1, logrusLevel, msg1) })
			msg2 := fmt.Sprintf("%v - another log msg", zapLevel.String())
			asserter(t, func() { logAt(logrus2, logrusLevel, msg2) })
			msg3 := fmt.Sprintf("%v - yet another log msg", zapLevel.String())
			asserter(t, func() { logAt(logrus3, logrusLevel, msg3) })

			zap1.Sync()
			require.Equal(t, 1, zap1Logs.Len())
			require.Equal(t, msg1, zap1Logs.All()[0].Message)

			zap2.Sync()
			require.Equal(t, 1, zap2Logs.Len())
			require.Equal(t, msg2, zap2Logs.All()[0].Message)

			zap3.Sync()
			defaultZap.Sync()
			require.Equal(t, 0, zap3Logs.Len())
			require.Equal(t, 1, defaultLogs.Len())
			require.Equal(t, msg3, defaultLogs.All()[0].Message)
		})
	}
}

func TestRedirectMonitorLogsWithMissingMapEntryUsesDefaultLogger(t *testing.T) {
	defer setup()()
	for _, lLevel := range logrus.AllLevels {
		logrusLevel := lLevel
		zapLevel := logrusToZapLevel[logrusLevel]
		t.Run(logrusLevel.String(), func(t *testing.T) {
			defer unredirect()
			defaultZap, defaultZapLogs := newObservedZap(zapLevel)
			logToZap := newLogrusToZap(defaultZap)

			logrus1 := logrus.WithField("monitorType", "monitor1").WithField("monitorID", "id1")
			monitorLogrus1 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "monitor1", monitorID: "id1"}
			zap1, zap1Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus1, zap1)
			require.Same(t, zap1, logToZap.getZapLogger(monitorLogrus1))

			// remove zap1 from map
			z, ok := logToZap.loggerMap.LoadAndDelete(monitorLogrus1)
			require.True(t, ok)
			require.Equal(t, zap1, z.(*zap.Logger))
			require.Same(t, defaultZap, logToZap.getZapLogger(monitorLogrus1))

			asserter := getAsserter(logrusLevel)
			msg1 := fmt.Sprintf("%v - a log msg", zapLevel.String())
			asserter(t, func() { logAt(logrus1, logrusLevel, msg1) })

			zap1.Sync()
			defaultZap.Sync()
			assert.Equal(t, 0, zap1Logs.Len())
			require.Equal(t, 1, defaultZapLogs.Len())
			require.Equal(t, msg1, defaultZapLogs.All()[0].Message)
		})
	}
}

func TestRedirectSameMonitorManyInstancesLogs(t *testing.T) {
	defer setup()()
	for _, lLevel := range logrus.AllLevels {
		logrusLevel := lLevel
		zapLevel := logrusToZapLevel[logrusLevel]
		t.Run(logrusLevel.String(), func(t *testing.T) {
			defer unredirect()
			logToZap := newLogrusToZap(zap.NewNop())

			logrus1 := logrus.WithField("monitorType", "monitor1").WithField("monitorID", "id1")
			logrus2 := logrus.WithField("monitorType", "monitor1").WithField("monitorID", "id2")
			monitorLogrus1 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "monitor1", monitorID: "id1"}
			monitorLogrus2 := monitorLogrus{Logger: logrus.StandardLogger(), monitorType: "monitor1", monitorID: "id2"}

			zap1, zap1Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus1, zap1)

			zap2, zap2Logs := newObservedZap(zapLevel)
			logToZap.redirect(monitorLogrus2, zap2)

			asserter := getAsserter(logrusLevel)
			msg1 := fmt.Sprintf("%v - a log msg", zapLevel.String())
			msg2 := fmt.Sprintf("%v - another log msg", zapLevel.String())
			asserter(t, func() { logAt(logrus1, logrusLevel, msg1) })
			asserter(t, func() { logAt(logrus2, logrusLevel, msg2) })

			zap1.Sync()
			zap2.Sync()
			require.Equal(t, 1, zap1Logs.Len())
			require.Equal(t, msg1, zap1Logs.All()[0].Message)
			require.Equal(t, 1, zap2Logs.Len())
			require.Equal(t, msg2, zap2Logs.All()[0].Message)
		})
	}
}

func setup() func() {
	stdLogger := logrus.StandardLogger()
	// don't let Fatal() actually exit the process
	orig := stdLogger.ExitFunc
	stdLogger.ExitFunc = func(code int) {}
	return func() {
		stdLogger.ExitFunc = orig
	}
}

func getAsserter(level logrus.Level) func(t require.TestingT, f assert.PanicTestFunc, msgAndArgs ...any) {
	asserter := require.NotPanics
	if level == logrus.FatalLevel || level == logrus.PanicLevel {
		asserter = require.Panics
	}
	return asserter
}

func newObservedZap(level zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	return zap.New(core).WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic)), logs
}

func logAt(entry *logrus.Entry, level logrus.Level, msg string) {
	origLevel := entry.Logger.Level
	defer func() {
		entry.Logger.Level = origLevel
	}()
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

func unredirect() {
	logger := logrus.StandardLogger()
	keep := make(logrus.LevelHooks)
	for _, level := range logrus.AllLevels {
		keep[level] = make([]logrus.Hook, 0)
		for _, hook := range logger.Hooks[level] {
			if _, ok := hook.(*logrusToZap); !ok {
				keep[level] = append(keep[level], hook)
			}
		}
	}
	logger.ReplaceHooks(keep)
}
