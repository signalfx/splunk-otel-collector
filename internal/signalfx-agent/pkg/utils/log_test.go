package utils

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// FixedTime is always at 2016-01-05T15:04:05-06:00
var fixedTime func() time.Time

func init() {
	f, err := time.Parse(time.RFC3339, "2016-01-05T15:04:05-06:00")
	if err != nil {
		panic("unable to parse time")
	}
	fixedTime = func() time.Time { return f }
}

// pinnedNow returns a function that is compatible with time.Now that always
// returns the given t Time.
func pinnedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// advancedNow returns a now time.Now-compatible function that returns the time
// returned from oldNow advanced by the given Duration d.
func advancedNow(oldNow func() time.Time, d time.Duration) func() time.Time {
	return pinnedNow(oldNow().Add(d))
}

func TestThrottledLogger(t *testing.T) {
	now = pinnedNow(fixedTime())

	var output bytes.Buffer
	rootLogger := logrus.New()

	rootLogger.Out = &output
	logger := NewThrottledLogger(rootLogger.WithFields(logrus.Fields{
		"component": "tests",
	}), 10*time.Second)

	errMsg := "This is an error"
	otherMsg := "Another error"

	logger.ThrottledError(errMsg)
	logger.ThrottledError(otherMsg)

	assert.Contains(t, output.String(), errMsg, "first error wasn't logged")
	assert.Contains(t, output.String(), otherMsg, "second error wasn't logged")

	now = advancedNow(now, 1*time.Second)

	output.Reset()
	logger.ThrottledError(errMsg)
	logger.ThrottledError(otherMsg)
	assert.Equal(t, "", output.String(), "errors aren't logged twice within duration")

	now = advancedNow(now, 11*time.Second)
	logger.ThrottledError(errMsg)
	assert.Contains(t, output.String(), errMsg, "first error wasn't logged after duration expired")

	output.Reset()
	logger.ThrottledError(errMsg)
	assert.Equal(t, "", output.String(), "first error didn't get suppressed")

	logger.ThrottledError(otherMsg)
	assert.Contains(t, output.String(), otherMsg, "second error wasn't logged after duration expired")

	output.Reset()
	derivedLogger := logger.WithFields(logrus.Fields{"name": "John"})

	derivedLogger.ThrottledError(errMsg)
	assert.Equal(t, "", output.String(), "first error didn't get suppressed in derived logger")

	now = advancedNow(now, 11*time.Second)
	derivedLogger.ThrottledError(errMsg)
	assert.Contains(t, output.String(), "John", "fields weren't copied in derived logger")
}
