package smartagentreceiver

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newLogRedirector(t *testing.T) *logRedirector {
	return &logRedirector{
		redirects:         make(map[srcLogger][]dstLogger),
		mu:                sync.Mutex{},
		dstCatchallLogger: zaptest.NewLogger(t),
	}
}

func TestLevelsMapShouldIncludeAllLogrusLogLevels(t *testing.T) {
	for _, level := range logrus.AllLevels {
		if _, ok := levelsMap[level]; !ok {
			t.Errorf("Expected logrus log level %s not found", level.String())
		}
	}
}

func TestLogRedirectorShouldReturnAllLogrusLogLevels(t *testing.T) {
	hook := newLogRedirector(t)
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		if logrus.AllLevels[i] != levels[i] {
			t.Errorf("Expected logrus log level %s not found", logrus.AllLevels[i].String())
		}
	}
}

func TestLogRedirectorShouldSetSrcReportCallerTrueOnRedirectCalls(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	newLogRedirector(t).redirect(src, zap.NewNop())
	if !src.ReportCaller {
		t.Errorf("Expected the source logrus logger to report caller after redirection")
	}
}

func TestLogRedirectorShouldSetSrcLoggerOutDiscardOnRedirectCalls(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	newLogRedirector(t).redirect(src, zap.NewNop())
	if src.Out != ioutil.Discard {
		t.Errorf("Expected the source logrus logger to be 'discarded' after redirection")
	}
}

func TestLogRedirectorShouldUniquelyAddHooksToSrcLoggerOnRedirectCalls(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	if got := len(src.Hooks); got != 0 {
		t.Errorf("Expected 0 hooks, got %d", got)
	}

	redirector := newLogRedirector(t)
	// These multiple redirect calls should add hook once to log levels.
	redirector.redirect(src, zap.NewNop())
	redirector.redirect(src, zap.NewNop())
	redirector.redirect(src, zap.NewNop())

	for _, level := range logrus.AllLevels {
		if got := len(src.Hooks[level]); got != 1 {
			t.Errorf("Expected 1 hook for logrus log level %s, got %d", level.String(), got)
		}
		if src.Hooks[level][0] != redirector {
			t.Errorf("Expected hook hook0 at index 0 for logrus log level %s not found", level.String())
		}
	}
}
