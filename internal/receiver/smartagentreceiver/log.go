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
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// srcLogger is the monitor's logrus logger.
// It is the source in a redirect to the monitor's receiver zap logger.
// Monitors use the same exported standard logrus logger.
// The logrus.Logger / monitorType combination is for unique identification of srcLogger.
// srcLogger(s) of multiple instances of the same monitor are indistinguishable.
type srcLogger struct {
	*logrus.Logger
	monitorType string
}

// dstLogger is the monitor's receiver zap logger.
// It is the destination in a redirect to the monitor's receiver zap logger.
type dstLogger *zap.Logger

// logRedirects maintains a map of srcLogger to dstLogger redirects.
// It hooks to srcLogger(s) and redirects log entries to dstLogger
type logRedirects struct {
	// The map values are slices to model destination(s) of indistinguishable sources.
	// However, redirections only apply to the first dstLogger.
	redirects   map[srcLogger][]dstLogger
	mu          sync.Mutex
	dstCatchall *zap.Logger
}

var _ logrus.Hook = (*logRedirects)(nil)

var (
	levelsMap = map[logrus.Level]zapcore.Level{
		logrus.PanicLevel: zapcore.PanicLevel,
		logrus.FatalLevel: zapcore.FatalLevel,
		logrus.ErrorLevel: zapcore.ErrorLevel,
		logrus.WarnLevel:  zapcore.WarnLevel,
		logrus.InfoLevel:  zapcore.InfoLevel,
		logrus.DebugLevel: zapcore.DebugLevel,
		// No zap level equivalent to trace. Mapping trace to debug.
		logrus.TraceLevel: zapcore.DebugLevel,
	}
)

func (l *srcLogger) reportCaller() {
	if !l.ReportCaller {
		l.SetReportCaller(true)
	}
}

func (l *srcLogger) discardOutput() {
	if l.Out != ioutil.Discard {
		l.SetOutput(ioutil.Discard)
	}
}

func (l *srcLogger) hook(hook *logRedirects) {
	added := false
	for _, hooks := range l.Hooks {
		for _, h := range hooks {
			if hook == h.(*logRedirects) {
				added = true
				break
			}
		}
	}
	if !added {
		l.AddHook(hook)
	}
}

func (l *logRedirects) redirect(src srcLogger, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.redirects[src]; !ok {
		l.redirects[src] = make([]dstLogger, 0)
	}

	src.reportCaller()
	src.discardOutput()
	src.hook(l)

	l.redirects[src] = append(l.redirects[src], dst)
}

func (l *logRedirects) unRedirect(src srcLogger, dst *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.redirects[src]; !ok {
		return
	}

	keep := make([]dstLogger, 0)

	for _, logger := range l.redirects[src] {
		if logger != dst {
			keep = append(keep, logger)
		}
	}

	l.redirects[src] = keep
}

// get1stDstLogger returns the first destination zap logger.
func (l *logRedirects) get1stDstLogger(src srcLogger) *zap.Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if loggers, ok := l.redirects[src]; !ok || len(loggers) == 0 {
		return nil
	}

	return l.redirects[src][0]
}

// Levels is a logrus.Hook interface method that returns all logrus logging levels.
// The hook is fired when logging on the logging levels returned by Levels.
func (l *logRedirects) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a logrus.Hook interface method that is called when logging on the logging levels returned by Levels.
// Fire creates a dstLogger zap entry from the supplied srcLogger logrus entry.
func (l *logRedirects) Fire(e *logrus.Entry) error {
	var monitorType string

	fields := make([]zapcore.Field, 0)

	// Creating zap entry fields and getting the monitor type from the logrus entry.
	for k, v := range e.Data {
		vStr := fmt.Sprintf("%v", v)
		if k == "monitorType" {
			monitorType = vStr
		}
		fields = append(fields, zapcore.Field{Key: k, Type: zapcore.StringType, String: vStr})
	}

	// Getting the first destination zap logger for the given monitor logrus logger.
	dstLogger1st := l.get1stDstLogger(srcLogger{e.Logger, monitorType})
	if dstLogger1st == nil {
		fields = append(fields, zapcore.Field{Key: "monitorType", Type: zapcore.StringType, String: monitorType})
		fields = append(fields, zapcore.Field{Key: "redirect_error", Type: zapcore.StringType, String: "Could not find zap logger in receiver for the monitorType. Using the catchall zap logger instead."})
		dstLogger1st = l.dstCatchall
	}

	// Creating zap entry from the logrus entry.
	if ce := dstLogger1st.Check(levelsMap[e.Level], e.Message); ce != nil {
		ce.Time = e.Time
		ce.Stack = ""
		if e.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(e.Caller.PC, e.Caller.File, e.Caller.Line, true)
		}
		ce.Write(fields...)
	}

	return nil
}
