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

package zaputil

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
)

type zapWrapper zap.Logger

var _ logrus.Hook = (*zapWrapper)(nil)

var levelsMap = map[logrus.Level]zapcore.Level {
	logrus.PanicLevel: zapcore.PanicLevel,
	logrus.FatalLevel: zapcore.FatalLevel,
	logrus.ErrorLevel: zapcore.ErrorLevel,
	logrus.WarnLevel: zapcore.WarnLevel,
	logrus.InfoLevel: zapcore.InfoLevel,
	logrus.DebugLevel: zapcore.DebugLevel,
	// No zap level equivalent to trace. Mapping trace to debug.
	logrus.TraceLevel: zapcore.DebugLevel,
}

// RedirectLogrusLogs mutes the output of the supplied logrus.Logger and adds a hook to it.
// When fired the hook creates a zap entry using the supplied logrus entry.
func RedirectLogrusLogs(from *logrus.Logger, to *zap.Logger) {
	from.ReportCaller = true
	from.SetOutput(ioutil.Discard)
	from.AddHook((*zapWrapper)(to))
}

// Levels is a logrus.Hook interface method that returns all logrus logging levels.
// The hook is fired when logging on the logging levels returned by Levels.
func (z *zapWrapper) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a logrus.Hook interface method that is called when logging on the logging levels returned by Levels.
// Fire creates a zap entry from the supplied logrus entry then writes it out.
func (z *zapWrapper) Fire(e *logrus.Entry) error {
	if ce := (*zap.Logger)(z).Check(levelsMap[e.Level], e.Message); ce != nil {
		// Updating zap entry ce with logrus entry values.
		ce.Time = e.Time
		ce.Stack = ""
		if e.Caller != nil {
			ce.Caller = zapcore.NewEntryCaller(e.Caller.PC, e.Caller.File, e.Caller.Line, true)
		}

		// Creating zap log fields from logrus fields.
		fields := make([]zapcore.Field, 0)
		for k, v := range e.Data {
			fields = append(fields, zapcore.Field{Key: k, Type: zapcore.StringType, String: fmt.Sprintf("%v", v)})
		}

		ce.Write(fields...)
	}
	return nil
}