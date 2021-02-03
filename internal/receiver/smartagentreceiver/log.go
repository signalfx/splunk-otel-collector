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
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logrusKey struct {
	*logrus.Logger
	monitorType string
}

// logrusToZap stores a mapping of logrus to zap loggers in loggerMap and hooks to the logrus loggers.
type logrusToZap struct {
	loggerMap map[logrusKey][]*zap.Logger
	mu        sync.Mutex
}

var _ logrus.Hook = (*logrusToZap)(nil)

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

func (l *logrusKey) reportCaller() {
	if !l.ReportCaller {
		l.SetReportCaller(true)
	}
}

func (l *logrusKey) discardOutput() {
	if l.Out != ioutil.Discard {
		l.SetOutput(ioutil.Discard)
	}
}

func (l *logrusKey) hookUnique(newHook logrus.Hook) {
	for _, hooks := range l.Hooks {
		for _, hook := range hooks {
			if hook == newHook {
				return
			}
		}
	}
	l.AddHook(newHook)
}

func (l *logrusToZap) redirect(src logrusKey, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.loggerMap[src]; !ok {
		l.loggerMap[src] = make([]*zap.Logger, 0)
	}

	// mutating the logrus logger as a side effect.
	src.reportCaller()
	src.discardOutput()
	src.hookUnique(l)

	l.loggerMap[src] = append(l.loggerMap[src], dst)
}

func (l *logrusToZap) unRedirect(src logrusKey, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loggerMap == nil {
		return
	}

	vLen := len(l.loggerMap[src])
	if vLen == 0 || vLen == 1 {
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

func (l *logrusToZap) loggerMapValue0(src logrusKey) *zap.Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loggerMap == nil || len(l.loggerMap[src]) == 0 {
		return nil
	}

	return l.loggerMap[src][0]
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
			monitorType = fmt.Sprintf("%v", v)
		}
		fields = append(fields, zap.Any(k, e.Data))
	}

	if monitorType == "" {
		log.Panic("Expected non-zero log field monitorType")
	}

	zapLogger := l.loggerMapValue0(logrusKey{e.Logger, monitorType})
	if zapLogger == nil {
		return nil
	}

	if ce := zapLogger.Check(levelsMap[e.Level], e.Message); ce != nil {
		ce.Time = e.Time
		ce.Stack = ""
		if e.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(e.Caller.PC, e.Caller.File, e.Caller.Line, true)
		}
		ce.Write(fields...)
	}

	return nil
}
