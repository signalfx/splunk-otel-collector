// Copyright 2021 Splunk, Inc.
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
	"io/ioutil"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
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
	for logrusLevel, zapLevel := range levelsMap {
		//TODO: handle fatal and panic levels
		if logrusLevel == logrus.FatalLevel || logrusLevel == logrus.PanicLevel {
			continue
		}

		// Simulating the creation of logrus entry/loggers in monitors (monitor1, monitor2).
		monitor1Logger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})
		monitor2Logger := logrus.WithFields(logrus.Fields{"monitorType": "monitor2"})

		// Checking that the logrus standard logger is the logger for monitor1 and monitor2.
		require.Same(t, logrus.StandardLogger(), monitor1Logger.Logger, "Expected the standard logrus logger")
		require.Same(t, logrus.StandardLogger(), monitor2Logger.Logger, "Expected the standard logrus logger")

		// Simulating the creation of logrus keys for monitor1 and monitor2 in the smart agent receiver where:
		// 1. the monitor types (i.e. monitor1, monitor2) are known.
		// 2. the logger is assumed to be the standard logrus logger.
		monitor1LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}
		monitor2LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor2"}

		monitor1ZapLogger, monitor1ZapLogs := newObservedLogs(zapLevel)
		monitor2ZapLogger, monitor2ZapLogs := newObservedLogs(zapLevel)

		// Simulating logrus to zap redirections in receiver.
		logToZap := newLogrusToZap()
		logToZap.redirect(monitor1LogrusKey, monitor1ZapLogger)
		logToZap.redirect(monitor2LogrusKey, monitor2ZapLogger)

		logMsg1 := "a log msg1"
		logMsg2 := "a log msg2"

		// Simulating logging a message in the monitor1.
		logAt(monitor1Logger, logrusLevel, logMsg1)
		logAt(monitor2Logger, logrusLevel, logMsg2)

		// Checking the zap logger is logging the same number of messages logged by monitor1.
		require.Equal(t, 1, monitor1ZapLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", monitor1ZapLogs.Len()))
		require.Equal(t, 1, monitor2ZapLogs.Len(), fmt.Sprintf("Expected 1 log message, got %d", monitor2ZapLogs.Len()))

		// Checking that the zap logger is logging the message logged by monitor1.
		require.Equal(t, logMsg1, monitor1ZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg1, monitor1ZapLogs.All()[0].Message))
		require.Equal(t, logMsg2, monitor2ZapLogs.All()[0].Message, fmt.Sprintf("Expected message '%s', got '%s'", logMsg2, monitor2ZapLogs.All()[0].Message))

		logToZap.unRedirect(monitor1LogrusKey, monitor1ZapLogger)
		logToZap.unRedirect(monitor2LogrusKey, monitor2ZapLogger)
	}
}

func TestRedirectSameMonitorManyInstancesLogs(t *testing.T) {
	for logrusLevel, zapLevel := range levelsMap {
		//TODO: handle fatal and panic levels
		if logrusLevel == logrus.FatalLevel || logrusLevel == logrus.PanicLevel {
			continue
		}

		// Simulating the creation of logrus entry/loggers in instances of a monitor (monitor1).
		instance1Logger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})
		instance2Logger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

		// Simulating the creation of logrus keys for the instances of monitor1 in the smart agent receiver where:
		// 1. the monitor type (i.e. monitor1) is known.
		// 2. the logger is assumed to be the standard logrus logger.
		instance1LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}
		instance2LogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

		// Checking that the logrus standard logger is the logger for both monitor1 instances.
		require.Same(t, logrus.StandardLogger(), instance1Logger.Logger, "Expected the standard logrus logger")
		require.Same(t, logrus.StandardLogger(), instance2Logger.Logger, "Expected the standard logrus logger")
		// Checking that the logrus keys are equal.
		require.Equal(t, instance1LogrusKey, instance2LogrusKey, "Expected the standard logrus logger")

		instance1ZapLogger, instance1ZapLogs := newObservedLogs(zapLevel)
		instance2ZapLogger, instance2ZapLogs := newObservedLogs(zapLevel)

		// Simulating logrus to zap redirections in in the smart agent receiver.
		logToZap := newLogrusToZap()
		logToZap.redirect(instance1LogrusKey, instance1ZapLogger)
		logToZap.redirect(instance2LogrusKey, instance2ZapLogger)

		logMsg1 := "a log msg1"
		logMsg2 := "a log msg2"

		// Simulating logging messages in the instances of monitor1.
		logAt(instance1Logger, logrusLevel, logMsg1)
		logAt(instance2Logger, logrusLevel, logMsg2)

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
		_, ok := levelsMap[level]
		require.True(t, ok, fmt.Sprintf("Expected log level %s not found", level.String()))
	}
}

func TestLevelsShouldReturnAllLogrusLevels(t *testing.T) {
	hook := newLogrusToZap()
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		require.Equal(t, logrus.AllLevels[i], levels[i], fmt.Sprintf("Expected log level %s not found", logrus.AllLevels[i].String()))
	}
}

func TestRedirectShouldSetLogrusKeyLoggerReportCallerTrue(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	dst := zap.NewNop()
	logToZap := newLogrusToZap()
	logToZap.redirect(src, dst)
	require.True(t, src.ReportCaller, "Expected the logrus key logger to report caller")
	logToZap.unRedirect(src, dst)
}

func TestRedirectShouldDiscardLogrusKeyLoggerOutput(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	dst := zap.NewNop()
	logToZap := newLogrusToZap()
	logToZap.redirect(src, dst)
	require.Equal(t, ioutil.Discard, src.Out, "Expected the logrus key logger output to be 'discarded'")
	logToZap.unRedirect(src, dst)
}

func TestRedirectShouldUniquelyAddHooksToLogrusKeyLogger(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	require.Equal(t, 0, len(src.Hooks), fmt.Sprintf("Expected 0 hooks, got %d", len(src.Hooks)))

	logToZap := newLogrusToZap()
	dst := zap.NewNop()
	// These multiple redirect calls should add hook once to log levels.
	logToZap.redirect(src, dst)
	logToZap.redirect(src, dst)
	logToZap.redirect(src, dst)

	for _, level := range logrus.AllLevels {
		got := len(src.Hooks[level])
		require.Equal(t, 1, got, fmt.Sprintf("Expected 1 hook for log level %s, got %d", level.String(), got))
		require.Equal(t, logToZap, src.Hooks[level][0], fmt.Sprintf("Expected hook for log level %s not found", level.String()))
	}
	logToZap.unRedirect(src, dst)
}

func newLogrusToZap() *logrusToZap {
	return &logrusToZap{
		loggerMap: make(map[logrusKey][]*zap.Logger),
		mu:        sync.Mutex{},
	}
}

func newObservedLogs(level zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	return zap.New(core), logs
}
