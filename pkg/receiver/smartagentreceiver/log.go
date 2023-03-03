// Copyright OpenTelemetry Authors
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
	"io"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logrusToZapLevel = map[logrus.Level]zapcore.Level{
		logrus.DebugLevel: zapcore.DebugLevel,
		logrus.ErrorLevel: zapcore.ErrorLevel,
		logrus.FatalLevel: zapcore.FatalLevel,
		logrus.InfoLevel:  zapcore.InfoLevel,
		logrus.PanicLevel: zapcore.PanicLevel,
		// No zap level equivalent to trace. Mapping trace to debug.
		logrus.TraceLevel: zapcore.DebugLevel,
		logrus.WarnLevel:  zapcore.WarnLevel,
	}

	zapToLogrusLevel = map[zapcore.Level]logrus.Level{
		zapcore.DebugLevel: logrus.DebugLevel,
		zapcore.ErrorLevel: logrus.ErrorLevel,
		zapcore.FatalLevel: logrus.FatalLevel,
		zapcore.InfoLevel:  logrus.InfoLevel,
		zapcore.PanicLevel: logrus.PanicLevel,
		zapcore.WarnLevel:  logrus.WarnLevel,
	}

	// zapLevels must be in lowest to highest sensitivity order
	// given current getLevelFromCore() implementation
	zapLevels = []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.DPanicLevel,
		zapcore.PanicLevel,
		zapcore.FatalLevel,
	}
)

var _ logrus.Hook = (*logrusToZap)(nil)

// logrusToZap provides a logrus.Hook ~singleton that redirects logrus.Logger.Log() calls
// to the desired registered zap.Logger routed by agent-set "monitorType" and "monitorID" field values.
type logrusToZap struct {
	// ~sync.Map(map[monitorLogrus]*zap.Logger)
	loggerMap     *sync.Map
	noopLogger    *logrus.Logger
	defaultLogger *zap.Logger
}

func newLogrusToZap(defaultLogger *zap.Logger) *logrusToZap {
	return &logrusToZap{
		loggerMap:     &sync.Map{},
		defaultLogger: defaultLogger,
		noopLogger: &logrus.Logger{
			Out:       io.Discard,
			Formatter: new(noopFormatter),
			Hooks:     make(logrus.LevelHooks),
		},
	}
}

// redirect prepares the src monitorLogrus to reflect the dst zap.Logger's settings
// and registers it for rerouting in the logrus.Hook's Fire()
func (l *logrusToZap) redirect(src monitorLogrus, dst *zap.Logger) {
	if desiredLogrusLevel, ok := zapToLogrusLevel[getLevelFromCore(dst.Core())]; ok {
		src.Logger.SetLevel(desiredLogrusLevel)
	}

	src.initialize()
	src.addLogrusToZapHook(l)

	// we only register for the first monitorLogrus instance
	// since there should only be one logger per component
	_, _ = l.loggerMap.LoadOrStore(src, dst)
}

func (l *logrusToZap) getZapLogger(src monitorLogrus) *zap.Logger {
	logger := l.defaultLogger
	if l.loggerMap != nil {
		if z, ok := l.loggerMap.Load(src); ok {
			logger = z.(*zap.Logger)
		}
	}
	return logger
}

// Levels is a logrus.Hook method that returns all logrus logging levels so
// that its Fire() is executed for all logrus logging activity.
func (l *logrusToZap) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a logrus.Hook method that is called when logging on the logging levels returned by Levels.
// A zap log entry is created from the supplied logrus entry and written out.
func (l *logrusToZap) Fire(entry *logrus.Entry) error {
	var monitorType string
	var monitorID string

	fields := make([]zapcore.Field, 0)

	for k, v := range entry.Data {
		if k == "monitorType" {
			monitorType = strings.TrimSpace(fmt.Sprintf("%v", v))
		} else if k == "monitorID" {
			monitorID = strings.TrimSpace(fmt.Sprintf("%v", v))
		}
		fields = append(fields, zap.Any(k, v))
	}

	zapLogger := l.getZapLogger(monitorLogrus{
		Logger:      entry.Logger,
		monitorType: monitorType,
		monitorID:   monitorID,
	})

	if ce := zapLogger.Check(logrusToZapLevel[entry.Level], entry.Message); ce != nil {
		ce.Time = entry.Time
		// clear stack so that it's not for parent Check()
		ce.Stack = ""
		if entry.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(entry.Caller.PC, entry.Caller.File, entry.Caller.Line, true)
		}
		ce.Write(fields...)
	}

	// we must set the Entry's logger to noop to prevent writing to the
	// StandardLogger during further processing. This will only be for the lifetime
	// of the hook evaluation chain
	entry.Logger = l.noopLogger

	return nil
}

type monitorLogrus struct {
	// in practice this will always be the logrus.StandardLogger
	// but embed it to support potential others
	*logrus.Logger
	// the monitor-logged "monitorType" field
	monitorType string
	// the monitor-logged "monitorID" field
	monitorID string
}

func (ml *monitorLogrus) initialize() {
	if !ml.ReportCaller {
		ml.SetReportCaller(true)
	}
}

func (ml *monitorLogrus) addLogrusToZapHook(l *logrusToZap) {
	for _, hooks := range ml.Hooks {
		for _, existing := range hooks {
			if existing == l {
				return
			}
		}
	}
	ml.AddHook(l)
}

func getLevelFromCore(core zapcore.Core) zapcore.Level {
	for _, level := range zapLevels {
		if core.Enabled(level) {
			return level
		}
	}
	return zapcore.InfoLevel
}

type noopFormatter struct{}

func (n *noopFormatter) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}
