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
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Monitor logger
type monitorLogger struct {
	*logrus.Logger
	monitorType string
}

// Mapping of monitor loggers to receiver loggers.
type loggersMapWrap struct {
	loggersMap map[monitorLogger][]*zap.Logger
	mu         sync.Mutex
}

var _ logrus.Hook = (*loggersMapWrap)(nil)

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

func (l *loggersMapWrap) add(key monitorLogger, val *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.loggersMap[key]; !ok {
		l.loggersMap[key] = make([]*zap.Logger, 0)
	}

	l.loggersMap[key] = append(l.loggersMap[key], val)
}

func (l *loggersMapWrap) remove(key monitorLogger, val *zap.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.loggersMap[key]; !ok {
		return
	}

	keep := make([]*zap.Logger, 0)

	for _, logger := range l.loggersMap[key] {
		if logger != val {
			keep = append(keep, logger)
		}
	}

	l.loggersMap[key] = keep
}

// get0 returns the first zap logger.
func (l *loggersMapWrap) get0(key monitorLogger) (*zap.Logger, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if loggers, ok := l.loggersMap[key]; !ok || len(loggers) == 0 {
		return nil, fmt.Errorf("missing zap logger for monitor %s", key.monitorType)
	}

	return l.loggersMap[key][0], nil
}

// Levels is a logrus.Hook interface method that returns all logrus logging levels.
// The hook is fired when logging on the logging levels returned by Levels.
func (l *loggersMapWrap) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a logrus.Hook interface method that is called when logging on the logging levels returned by Levels.
// Fire creates a zap entry from the supplied logrus entry then writes to it.
func (l *loggersMapWrap) Fire(e *logrus.Entry) error {
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

	key := monitorLogger{e.Logger, monitorType}
	// Getting the zap logger for the given key.
	l0, err := l.get0(key)
	if err != nil {
		return err
	}

	// Creating zap entry from the logrus entry.
	if ce := l0.Check(levelsMap[e.Level], e.Message); ce != nil {
		ce.Time = e.Time
		ce.Stack = ""
		if e.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(e.Caller.PC, e.Caller.File, e.Caller.Line, true)
		}
		ce.Write(fields...)
	}
	return nil
}
