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
	"io/ioutil"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestRedirectTraceLogs(t *testing.T) {
	// Creating a typical monitor logrus entry/logger.
	monitorLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

	// Creating the monitor logrus key in the monitor receiver.
	// The monitor type is known.
	// The logger is assumed to be the standard logrus logger.
	monitorLogrusKey := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

	// Checking that the monitor logger and the assumed logrus key logger are the same.
	if monitorLogger.Logger != monitorLogrusKey.Logger {
		t.Error("Expected the standard logrus logger")
	}

	zapLogger, zapLogs := newObservedLogs(zap.DebugLevel)

	// logrus to zap redirection of monitor logs.
	newLogrusToZap(t).redirect(monitorLogrusKey, zapLogger)

	// Simulating logging a message in the monitor.
	monitorLogger.Logger.Level = logrus.TraceLevel
	monitorLogger.Trace("a trace msg")

	// Checking that the zap logger is logging the message logged by the monitor.
	assertLogMsg(t, zapLogs, "a trace msg")
}

func TestRedirectDebugLogs(t *testing.T) {
	// Creating a typical monitor logrus entry/logger.
	monitorLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

	// Creating the monitor logrus key in the monitor receiver.
	// The monitor type is known.
	// The logger is assumed to be the standard logrus logger.
	receiverMonitorLogger := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

	// Checking that the monitor logger and the assumed logrus key logger are the same.
	if monitorLogger.Logger != receiverMonitorLogger.Logger {
		t.Error("Expected the standard logrus logger")
	}

	zapLogger, zapLogs := newObservedLogs(zap.DebugLevel)

	// logrus to zap redirection of monitor logs.
	newLogrusToZap(t).redirect(receiverMonitorLogger, zapLogger)

	// Simulating logging a message in the monitor.
	monitorLogger.Logger.Level = logrus.DebugLevel
	monitorLogger.Debug("a debug msg")

	// Checking that the zap logger is logging the message logged by the monitor.
	assertLogMsg(t, zapLogs, "a debug msg")
}

func TestRedirectInfoLogs(t *testing.T) {
	// Creating a typical monitor logrus entry/logger.
	monitorLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

	// Creating the monitor logrus key in the monitor receiver.
	// The monitor type is known.
	// The logger is assumed to be the standard logrus logger.
	receiverMonitorLogger := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

	// Checking that the monitor logger and the assumed logrus key logger are the same.
	if monitorLogger.Logger != receiverMonitorLogger.Logger {
		t.Error("Expected the standard logrus logger")
	}

	zapLogger, zapLogs := newObservedLogs(zap.InfoLevel)

	// logrus to zap redirection of monitor logs.
	newLogrusToZap(t).redirect(receiverMonitorLogger, zapLogger)

	// Simulating logging a message in the monitor.
	monitorLogger.Logger.Level = logrus.InfoLevel
	monitorLogger.Info("an info msg")

	// Checking that the zap logger is logging the message logged by the monitor.
	assertLogMsg(t, zapLogs, "an info msg")
}

func TestRedirectWarnLogs(t *testing.T) {
	// Creating a typical monitor logrus entry/logger.
	monitorLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

	// Creating the monitor logrus key in the monitor receiver.
	// The monitor type is known.
	// The logger is assumed to be the standard logrus logger.
	receiverMonitorLogger := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

	// Checking that the monitor logger and the assumed logrus key logger are the same.
	if monitorLogger.Logger != receiverMonitorLogger.Logger {
		t.Error("Expected the standard logrus logger")
	}

	zapLogger, zapLogs := newObservedLogs(zap.WarnLevel)

	// logrus to zap redirection of monitor logs.
	newLogrusToZap(t).redirect(receiverMonitorLogger, zapLogger)

	// Simulating logging a message in the monitor.
	monitorLogger.Logger.Level = logrus.WarnLevel
	monitorLogger.Warn("a warn msg")

	// Checking that the zap logger is logging the message logged by the monitor.
	assertLogMsg(t, zapLogs, "a warn msg")
}

func TestRedirectErrorLogs(t *testing.T) {
	// Creating a typical monitor logrus entry/logger.
	monitorLogger := logrus.WithFields(logrus.Fields{"monitorType": "monitor1"})

	// Creating the monitor logrus key in the monitor receiver.
	// The monitor type is known.
	// The logger is assumed to be the standard logrus logger.
	receiverMonitorLogger := logrusKey{Logger: logrus.StandardLogger(), monitorType: "monitor1"}

	// Checking that the monitor logger and the assumed logrus key logger are the same.
	if monitorLogger.Logger != receiverMonitorLogger.Logger {
		t.Error("Expected the standard logrus logger")
	}

	zapLogger, zapLogs := newObservedLogs(zap.ErrorLevel)

	// logrus to zap redirection of monitor logs.
	newLogrusToZap(t).redirect(receiverMonitorLogger, zapLogger)

	// Simulating logging a message in the monitor.
	monitorLogger.Logger.Level = logrus.ErrorLevel
	monitorLogger.Error("an error msg")

	// Checking that the zap logger is logging the message logged by the monitor.
	assertLogMsg(t, zapLogs, "an error msg")
}

func TestLevelsMapShouldIncludeAllLogrusLogLevels(t *testing.T) {
	for _, level := range logrus.AllLevels {
		if _, ok := levelsMap[level]; !ok {
			t.Errorf("Expected logrus log level %s not found", level.String())
		}
	}
}

func TestRedirectShouldReturnAllLogrusLogLevels(t *testing.T) {
	hook := newLogrusToZap(t)
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		if logrus.AllLevels[i] != levels[i] {
			t.Errorf("Expected logrus log level %s not found", logrus.AllLevels[i].String())
		}
	}
}

func TestRedirectShouldSetSrcReportCallerTrueOnRedirectCalls(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	newLogrusToZap(t).redirect(src, zap.NewNop())
	if !src.ReportCaller {
		t.Errorf("Expected the source logrus logger to report caller after redirection")
	}
}

func TestRedirectShouldSetSrcLoggerOutDiscardOnRedirectCalls(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	newLogrusToZap(t).redirect(src, zap.NewNop())
	if src.Out != ioutil.Discard {
		t.Errorf("Expected the source logrus logger to be 'discarded' after redirection")
	}
}

func TestRedirectShouldUniquelyAddHooksToSrcLoggerOnRedirectCalls(t *testing.T) {
	src := logrusKey{Logger: logrus.New(), monitorType: "monitor1"}
	if got := len(src.Hooks); got != 0 {
		t.Errorf("Expected 0 hooks, got %d", got)
	}

	rdr := newLogrusToZap(t)
	// These multiple redirect calls should add hook once to log levels.
	rdr.redirect(src, zap.NewNop())
	rdr.redirect(src, zap.NewNop())
	rdr.redirect(src, zap.NewNop())

	for _, level := range logrus.AllLevels {
		if got := len(src.Hooks[level]); got != 1 {
			t.Errorf("Expected 1 hook for logrus log level %s, got %d", level.String(), got)
		}
		if src.Hooks[level][0] != rdr {
			t.Errorf("Expected hook hook0 at index 0 for logrus log level %s not found", level.String())
		}
	}
}

func newLogrusToZap(t *testing.T) *logrusToZap {
	return &logrusToZap{
		loggerMap:      make(map[logrusKey][]*zap.Logger),
		mu:             sync.Mutex{},
		catchallLogger: zaptest.NewLogger(t),
	}
}

func assertLogMsg(t *testing.T, logs *observer.ObservedLogs, msg string) {
	numLogs, entry := logs.Len(), logs.All()[0]
	if numLogs != 1 || entry.Message != msg {
		t.Errorf("Invalid log entry %v", entry)
	}
}

func newObservedLogs(level zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	return zap.New(core), logs
}
