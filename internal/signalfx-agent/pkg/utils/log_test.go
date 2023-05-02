package utils

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/signalfx/signalfx-agent/pkg/neotest"
)

func TestThrottledLogger(t *testing.T) {
	now = neotest.PinnedNow(neotest.FixedTime())

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

	now = neotest.AdvancedNow(now, 1*time.Second)

	output.Reset()
	logger.ThrottledError(errMsg)
	logger.ThrottledError(otherMsg)
	assert.Equal(t, "", output.String(), "errors aren't logged twice within duration")

	now = neotest.AdvancedNow(now, 11*time.Second)
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

	now = neotest.AdvancedNow(now, 11*time.Second)
	derivedLogger.ThrottledError(errMsg)
	assert.Contains(t, output.String(), "John", "fields weren't copied in derived logger")
}
