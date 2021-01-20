package smartagentreceiver

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapHooks struct {
	ZapLogger *zap.Logger
}

var levelMap = map[logrus.Level]zapcore.Level {
	logrus.PanicLevel: zapcore.PanicLevel,
	logrus.FatalLevel: zapcore.FatalLevel,
	logrus.ErrorLevel: zapcore.ErrorLevel,
	logrus.WarnLevel: zapcore.WarnLevel,
	logrus.InfoLevel: zapcore.InfoLevel,
	logrus.DebugLevel: zapcore.DebugLevel,
	// No trace level in zap thus mapping trace to debug.
	logrus.TraceLevel: zapcore.DebugLevel,
}

func (hooks *ZapHooks) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hooks *ZapHooks) Fire(logrusEntry *logrus.Entry) error {
	if checkedEntry := hooks.ZapLogger.Check(levelMap[logrusEntry.Level], logrusEntry.Message); checkedEntry != nil {
		updateCheckedEntry(checkedEntry, logrusEntry)
		zapFields := toZapFields(logrusEntry.Data)
		checkedEntry.Write(zapFields...)
	}
	return nil
}

// updateCheckedEntry assigns logrus log entry fields to zap log entry fields.
// logrus is missing fields equivalent to zapcore.Entry.Defined, zapcore.Entry.Stack, zapcore.Entry.LoggerName
func updateCheckedEntry(checkedEntry *zapcore.CheckedEntry, logrusEntry *logrus.Entry) {
	checkedEntry.Time = logrusEntry.Time
	if logrusEntry.Caller != nil {
		checkedEntry.Caller = zapcore.NewEntryCaller(logrusEntry.Caller.PC, logrusEntry.Caller.File, logrusEntry.Caller.Line, true)
	}
	checkedEntry.Stack = ""
}

func toZapFields(fields logrus.Fields) []zapcore.Field {
	zapFields := make([]zapcore.Field, 0)
	for k, v := range fields {
		zapFields = append(zapFields,zapcore.Field{Key: k, Interface: v} )
	}
	return zapFields
}
