// Copyright 2021, OpenTelemetry Authors
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

package smartagentreceiver

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	levelsMap = map[logrus.Level]zapcore.Level{
		logrus.FatalLevel: zapcore.FatalLevel,
		logrus.PanicLevel: zapcore.PanicLevel,
		logrus.ErrorLevel: zapcore.ErrorLevel,
		logrus.WarnLevel:  zapcore.WarnLevel,
		logrus.InfoLevel:  zapcore.InfoLevel,
		logrus.DebugLevel: zapcore.DebugLevel,
		// No zap level equivalent to trace. Mapping trace to debug.
		logrus.TraceLevel: zapcore.DebugLevel,
	}
)

var _ logrus.Hook = (*logrusToZap)(nil)

// logrusToZap stores a mapping of logrus to zap loggers in loggerMap and hooks to the logrus loggers.
type logrusToZap struct {
	loggerMap     map[logrusKey][]*zap.Logger
	mu            sync.Mutex
	noopLogger    *logrus.Logger
	defaultLogger *zap.Logger
}

type noopFormatter struct{}

type logrusKey struct {
	*logrus.Logger
	monitorType string
}

func (l *noopFormatter) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}

func (l *logrusKey) reportCaller() {
	if !l.ReportCaller {
		l.SetReportCaller(true)
	}
}

func (l *logrusKey) addHookUnique(newHook logrus.Hook) {
	for _, hooks := range l.Hooks {
		for _, hook := range hooks {
			if hook == newHook {
				return
			}
		}
	}
	l.AddHook(newHook)
}

func (l *logrusKey) removeHook(remove logrus.Hook, levels ...logrus.Level) {
	if levels == nil {
		levels = logrus.AllLevels
	}

	keep := make(logrus.LevelHooks)
	for _, level := range levels {
		keep[level] = make([]logrus.Hook, 0)
		for _, hook := range l.Hooks[level] {
			if hook != remove {
				keep[level] = append(keep[level], hook)
			}
		}
	}

	l.ReplaceHooks(keep)
}

func newLogrusToZap() *logrusToZap {
	logger, err := newDefaultLoggerCfg().Build()
	if err != nil {
		log.Fatalf("Cannot initialize the default zap logger: %v", err)
	}
	defer logger.Sync()

	return &logrusToZap{
		loggerMap:     make(map[logrusKey][]*zap.Logger),
		mu:            sync.Mutex{},
		defaultLogger: logger,
		noopLogger: &logrus.Logger{
			Out:       ioutil.Discard,
			Formatter: new(noopFormatter),
			Hooks:     make(logrus.LevelHooks),
		},
	}
}

func newDefaultLoggerCfg() *zap.Config {
	defaultLoggerCfg := zap.NewProductionConfig()
	defaultLoggerCfg.Encoding = "console"
	defaultLoggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	defaultLoggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	defaultLoggerCfg.InitialFields = map[string]interface{}{
		"component_kind": "receiver",
		"component_type": "smartagent",
	}
	return &defaultLoggerCfg
}

func (l *logrusToZap) redirect(src logrusKey, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.loggerMap[src]; !ok {
		l.loggerMap[src] = make([]*zap.Logger, 0)
	}

	src.reportCaller()
	src.addHookUnique(l)

	l.loggerMap[src] = append(l.loggerMap[src], dst)
}

func (l *logrusToZap) unRedirect(src logrusKey, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	vLen := len(l.loggerMap[src])
	if vLen == 0 || vLen == 1 {
		src.removeHook(l)
		delete(l.loggerMap, src)
		return
	}

	keep := make([]*zap.Logger, 0)
	for _, logger := range l.loggerMap[src] {
		if logger != dst {
			keep = append(keep, logger)
		}
	}
	l.loggerMap[src] = keep
}

func (l *logrusToZap) loggerMapValue0(src logrusKey) (*zap.Logger, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loggerMap == nil {
		return nil, false
	}

	loggers, inMap := l.loggerMap[src]

	if len(loggers) > 0 {
		return loggers[0], inMap
	}

	return nil, inMap
}

// Levels is a logrus.Hook implementation that returns all logrus logging levels.
func (l *logrusToZap) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a logrus.Hook implementation that is called when logging on the logging levels returned by Levels.
// A zap log entry is created from the supplied logrus entry and written out.
func (l *logrusToZap) Fire(e *logrus.Entry) error {
	var monitorType string

	fields := make([]zapcore.Field, 0)

	// Creating zap entry fields from logrus entry fields.
	for k, v := range e.Data {
		if k == "monitorType" {
			monitorType = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
		fields = append(fields, zap.Any(k, v))
	}

	if monitorType == "" {
		l.defaultLogger.Warn("Cannot find zap logger for monitor. The log field monitorType is missing or blank.")
		return nil
	}

	logger, _ := l.loggerMapValue0(logrusKey{e.Logger, monitorType})
	if logger == nil {
		l.defaultLogger.Warn(fmt.Sprintf("Cannot find zap logger for monitorType %s", monitorType))
		return nil
	}

	if ce := logger.Check(levelsMap[e.Level], e.Message); ce != nil {
		ce.Time = e.Time
		ce.Stack = ""
		if e.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(e.Caller.PC, e.Caller.File, e.Caller.Line, true)
		}
		ce.Write(fields...)
	}

	e.Logger = l.noopLogger

	return nil
}
