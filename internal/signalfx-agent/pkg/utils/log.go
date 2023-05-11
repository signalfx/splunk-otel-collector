package utils

import (
	"fmt"
	stdlog "log"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/signalfx/golib/v3/log"
	"github.com/sirupsen/logrus"

	apmlog "github.com/signalfx/signalfx-agent/pkg/apm/log"
)

func NewAPMShim(log logrus.FieldLogger) apmlog.Logger {
	return apmShim{log: log}
}

type apmShim struct {
	log logrus.FieldLogger
}

func (a apmShim) Debug(msg string) {
	a.log.Debug(msg)
}

func (a apmShim) Warn(msg string) {
	a.log.Warn(msg)
}

func (a apmShim) Error(msg string) {
	a.log.Error(msg)
}

func (a apmShim) Info(msg string) {
	a.log.Info(msg)
}

func (a apmShim) Panic(msg string) {
	a.log.Panic(msg)
}

func (a apmShim) WithFields(fields apmlog.Fields) apmlog.Logger {
	return apmShim{log: a.log.WithFields(logrus.Fields(fields))}
}

func (a apmShim) WithError(err error) apmlog.Logger {
	return apmShim{log: a.log.WithError(err)}
}

var _ apmlog.Logger = &apmShim{}

// LogrusGolibShim makes a Logrus logger conform to the golib Log interface
type LogrusGolibShim struct {
	logrus.FieldLogger
}

var _ log.Logger = (*LogrusGolibShim)(nil)

// Log conforms to the golib Log interface
func (l *LogrusGolibShim) Log(keyvals ...interface{}) {
	fields := logrus.Fields{}

	var currentKey *log.Key
	messages := []interface{}{}

	for k := range keyvals {
		switch v := keyvals[k].(type) {
		case log.Key:
			currentKey = &v
		default:
			if currentKey != nil {
				switch *currentKey {
				case log.Msg:
					messages = append(messages, v)
				default:
					fields[string(*currentKey)] = v
				}
				currentKey = nil
			} else {
				messages = append(messages, v)
			}
		}
	}

	fieldlog := logrus.WithFields(fields)

	if _, ok := fields[string(log.Err)]; ok {
		fieldlog.Error(messages...)
	} else {
		fieldlog.Info(messages...)
	}
}

var _ logrus.StdLogger = &ThrottledLogger{}

// For unit testing exposure
var now = time.Now

// ThrottledLogger throttles error and warning messages sent through it via the
// special ThrottledError method (other standard level methods are not
// throttled).  This doesn't technically conform to the Logrus FieldLogger
// interface because some of the chained methods return *Entry and we can't
// wrap those but must propagate the original instance to keep state.  It
// should, however, behave functionally the same.
type ThrottledLogger struct {
	logrus.FieldLogger

	errorsSeen   *lru.Cache
	warningsSeen *lru.Cache
	duration     time.Duration
}

// NewThrottledLogger returns an initialized ThrottleLogger.  The duration
// specifies the maximum frequency with which a specific error message will be
// logged.  All other duplicate messages within this duration will be ignored.
func NewThrottledLogger(logger logrus.FieldLogger, duration time.Duration) *ThrottledLogger {
	// We don't need room for many entries since in the most common case it
	// will only be one or a small handful of error messages being repeated.
	errorsSeen, err := lru.New(10)
	if err != nil {
		panic("could not create throttled logger LRU cache")
	}

	warningsSeen, err := lru.New(10)
	if err != nil {
		panic("could not create throttled logger LRU cache")
	}

	return &ThrottledLogger{
		FieldLogger:  logger,
		duration:     duration,
		errorsSeen:   errorsSeen,
		warningsSeen: warningsSeen,
	}
}

func (tl *ThrottledLogger) copy(newLogger logrus.FieldLogger) *ThrottledLogger {
	return &ThrottledLogger{
		FieldLogger: newLogger,
		errorsSeen:  tl.errorsSeen,
		duration:    tl.duration,
	}
}

// WithField is functionally equivalent to the logrus version of this method
func (tl *ThrottledLogger) WithField(key string, value interface{}) *ThrottledLogger {
	return tl.copy(tl.FieldLogger.WithField(key, value))
}

// WithFields is functionally equivalent to the logrus version of this method
func (tl *ThrottledLogger) WithFields(fields logrus.Fields) *ThrottledLogger {
	return tl.copy(tl.FieldLogger.WithFields(fields))
}

// WithError is functionally equivalent to the logrus version of this method
func (tl *ThrottledLogger) WithError(err error) *ThrottledLogger {
	return tl.copy(tl.FieldLogger.WithError(err))
}

// ThrottledError logs an error message, throttled.  Make the throttling
// explicit in the function name instead of implicit to the logger type since
// most error messages should be logged at full blast without having to use a
// different logger instance.
func (tl *ThrottledLogger) ThrottledError(args ...interface{}) {
	key := fmt.Sprint(args...)
	rightNow := now()

	if lastSeen, present := tl.errorsSeen.Get(key); present {
		if lastSeen.(*time.Time).Add(tl.duration).After(rightNow) {
			return
		}
	}
	tl.errorsSeen.Add(key, &rightNow)

	tl.FieldLogger.Error(args...)
}

// ThrottledWarning logs a warning message, throttled.  Make the throttling
// explicit in the function name instead of implicit to the logger type since
// some warning messages should be logged at full blast without having to use a
// different logger instance.
func (tl *ThrottledLogger) ThrottledWarning(args ...interface{}) {
	key := fmt.Sprint(args...)
	rightNow := now()

	if lastSeen, present := tl.warningsSeen.Get(key); present {
		if lastSeen.(*time.Time).Add(tl.duration).After(rightNow) {
			return
		}
	}
	tl.warningsSeen.Add(key, &rightNow)

	tl.FieldLogger.Warn(args...)
}

func NewStdLogWithLogrus(logger logrus.FieldLogger) *stdlog.Logger {
	return stdlog.New(&StdLogLogrusWriter{Logger: logger}, "", 0)
}

// This exploits the documented fact that the standard log pkg sends each log
// entry as a single io.Writer.Write call: https://golang.org/pkg/log/#Logger
type StdLogLogrusWriter struct {
	Logger logrus.FieldLogger
}

func (w *StdLogLogrusWriter) Write(b []byte) (int, error) {
	n := len(b)
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	w.Logger.Warning(string(b))
	return n, nil
}
