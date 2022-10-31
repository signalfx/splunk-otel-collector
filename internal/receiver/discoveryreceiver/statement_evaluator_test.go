// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discoveryreceiver

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver/statussources"
)

func TestStatementEvaluation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		match Match
	}{
		{name: "strict", match: Match{Strict: "desired.statement"}},
		{name: "regexp", match: Match{Regexp: "^d[esired]{6}.statement$"}},
		{name: "expr", match: Match{Expr: "msg == 'desired.statement'"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			match := tc.match
			match.Record = &LogRecord{
				Body: "desired body content",
				Attributes: map[string]string{
					"one": "one.value", "two": "two.value",
				},
			}
			for _, status := range discovery.StatusTypes {
				t.Run(string(status), func(t *testing.T) {
					for _, level := range []string{"debug", "info", "warn", "error", "fatal", "dpanic", "panic"} {
						t.Run(level, func(t *testing.T) {
							for _, firstOnly := range []bool{true, false} {
								match.FirstOnly = firstOnly
								t.Run(fmt.Sprintf("FirstOnly:%v", firstOnly), func(t *testing.T) {
									observerID := config.NewComponentIDWithName("an.observer", "observer.name")
									cfg := &Config{
										Receivers: map[config.ComponentID]ReceiverEntry{
											config.NewComponentIDWithName("a.receiver", "receiver.name"): {
												Rule:   "a.rule",
												Status: &Status{Statements: map[discovery.StatusType][]Match{status: {match}}},
											},
										},
										WatchObservers: []config.ComponentID{observerID},
									}
									require.NoError(t, cfg.Validate())

									plogs := make(chan plog.Logs)

									logger := zaptest.NewLogger(t)
									cStore := newCorrelationStore(logger, time.Hour)
									cStore.UpdateEndpoint(
										observer.Endpoint{ID: "endpoint.id"},
										addedState, observerID,
									)

									se, err := newStatementEvaluator(logger, cfg, plogs, cStore)
									require.NoError(t, err)

									evaluatedLogger := se.evaluatedLogger.With(
										zap.String("name", `a.receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`),
									)

									numExpected := 1
									if !firstOnly {
										numExpected = 3
									}

									emitted := plog.NewLogs()
									wg := sync.WaitGroup{}
									wg.Add(numExpected)

									go func() {
										timer := time.NewTimer(50 * time.Millisecond)
										for i := 0; i < numExpected; i++ {
											select {
											case logs := <-plogs:
												if emitted.LogRecordCount() == 0 {
													emitted = logs
												} else {
													logs.ResourceLogs().MoveAndAppendTo(emitted.ResourceLogs())
												}
												wg.Done()
											case <-timer.C:
												return
											}
										}
									}()

									logMethod := map[string]func(string, ...zap.Field){
										"debug":  evaluatedLogger.Debug,
										"info":   evaluatedLogger.Info,
										"warn":   evaluatedLogger.Warn,
										"error":  evaluatedLogger.Error,
										"fatal":  evaluatedLogger.Fatal,
										"dpanic": evaluatedLogger.DPanic,
										"panic":  evaluatedLogger.Panic,
									}[level]

									for _, statement := range []string{
										"undesired.statement",
										"another.undesired.statement",
										"desired.statement",
										"desired.statement",
										"desired.statement",
									} {
										panicCheck := require.NotPanics
										if level == "panic" {
											panicCheck = require.Panics
										}
										panicCheck(t, func() {
											logMethod(statement)
										})
									}

									require.Eventually(t, func() bool {
										wg.Wait()
										return true
									}, 50*time.Millisecond, time.Millisecond)

									for i := 0; i < numExpected; i++ {
										rl := emitted.ResourceLogs().At(i)
										rAttrs := rl.Resource().Attributes()
										require.Equal(t, map[string]any{
											"discovery.endpoint.id":   "endpoint.id",
											"discovery.event.type":    "statement.match",
											"discovery.observer.id":   "an.observer/observer.name",
											"discovery.receiver.name": "receiver.name",
											"discovery.receiver.rule": "a.rule",
											"discovery.receiver.type": "a.receiver",
										}, rAttrs.AsRaw())

										sLogs := rl.ScopeLogs()
										require.Equal(t, 1, sLogs.Len())
										sl := sLogs.At(0)
										lrs := sl.LogRecords()
										require.Equal(t, 1, lrs.Len())
										lr := sl.LogRecords().At(0)

										lrAttrs := lr.Attributes().AsRaw()

										require.Contains(t, lrAttrs, "caller")
										_, expectedFile, _, _ := runtime.Caller(0)
										// runtime doesn't use os.PathSeparator
										splitPath := strings.Split(expectedFile, "/")
										expectedCaller := splitPath[len(splitPath)-1]
										require.Contains(t, lrAttrs["caller"], expectedCaller)
										delete(lrAttrs, "caller")

										// argOrder doesn't like this for some reason
										// nolint:gocritic
										if strings.Contains("error fatal dpanic panic", level) {
											require.Contains(t, lrAttrs, "stacktrace")
											delete(lrAttrs, "stacktrace")
										}

										require.Equal(t, map[string]any{
											"discovery.status": string(status),
											"name":             `a.receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`,
											"one":              "one.value",
											"two":              "two.value",
										}, lrAttrs)

										require.Equal(t, "desired body content", lr.Body().AsString())
										require.Equal(t, level, lr.SeverityText())
									}
								})
							}
						})
					}
				})
			}
		})
	}
}

func TestLogRecordDefaultAndArbitrarySeverityText(t *testing.T) {
	observerID := config.NewComponentIDWithName("an.observer", "observer.name")
	cfg := &Config{
		Receivers: map[config.ComponentID]ReceiverEntry{
			config.NewComponentIDWithName("a.receiver", "receiver.name"): {
				Rule:   "a.rule",
				Status: &Status{Statements: map[discovery.StatusType][]Match{discovery.Successful: {Match{Strict: "match.me"}}}},
			},
		},
		WatchObservers: []config.ComponentID{observerID},
	}
	require.NoError(t, cfg.Validate())

	plogs := make(chan plog.Logs)

	logger := zaptest.NewLogger(t)
	cStore := newCorrelationStore(logger, time.Hour)
	cStore.UpdateEndpoint(
		observer.Endpoint{ID: "endpoint.id"},
		addedState, observerID,
	)

	se, err := newStatementEvaluator(logger, cfg, plogs, cStore)
	require.NoError(t, err)
	require.NotNil(t, se)

	s := &statussources.Statement{
		Message:    "match.me",
		Level:      "",
		Time:       time.Now(),
		LoggerName: "logger.name",
		Fields: map[string]any{
			"name": `a.receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`,
		},
	}

	logs := se.evaluateStatement(s)
	require.Equal(t, 1, logs.LogRecordCount())
	lr := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	require.Equal(t, "INFO", lr.SeverityText())

	s.Level = "arbitrary level used as severity text"
	logs = se.evaluateStatement(s)
	require.Equal(t, 1, logs.LogRecordCount())
	lr = logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	require.Equal(t, "arbitrary level used as severity text", lr.SeverityText())
}
