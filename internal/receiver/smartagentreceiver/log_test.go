package smartagentreceiver

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newLogRedirect(t *testing.T) *logRedirect {
	return &logRedirect{
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

func TestLogRedirectShouldReturnAllLogrusLogLevels(t *testing.T) {
	hook := newLogRedirect(t)
	levels := hook.Levels()
	for i := range logrus.AllLevels {
		if logrus.AllLevels[i] != levels[i] {
			t.Errorf("Expected logrus log level %s not found", logrus.AllLevels[i].String())
		}
	}
}

func TestLogRedirectShouldSetSrcReportCallerTrue(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	newLogRedirect(t).redirect(src, zap.NewNop())
	if !src.ReportCaller {
		t.Errorf("Expected the source logrus logger to report caller after redirection")
	}
}

func TestLogRedirectShouldSetSrcLoggerOutDiscard(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	newLogRedirect(t).redirect(src, zap.NewNop())
	if src.Out != ioutil.Discard {
		t.Errorf("Expected the source logrus logger to be 'discarded' after redirection")
	}
}

func TestSrcLoggerHookOnce(t *testing.T) {
	src := srcLogger{Logger: logrus.New(), monitorType: "monitor1"}
	if got := len(src.Hooks); got != 0 {
		t.Errorf("Expected 0 hooks, got %d", got)
	}

	hook0 := newLogRedirect(t)
	// hooking hook0 twice to all logrus log levels in src
	src.hookOnce(hook0)
	src.hookOnce(hook0)
	hook1 := newLogRedirect(t)
	// hooking hook1 thrice to all logrus logs in src
	src.hookOnce(hook1)
	src.hookOnce(hook1)
	src.hookOnce(hook1)

	for _, level := range logrus.AllLevels {
		if got := len(src.Hooks[level]); got != 2 {
			t.Errorf("Expected 2 hook for logrus log level %s, got %d", level.String(), got)
		}
		if src.Hooks[level][0] != hook0 {
			t.Errorf("Expected hook hook0 at index 0 for logrus log level %s not found", level.String())
		}
		if src.Hooks[level][1] != hook1 {
			t.Errorf("Expected hook hook1 at index 1 for logrus log level %s not found", level.String())
		}
	}
}
