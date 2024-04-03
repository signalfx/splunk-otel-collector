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
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestStatementEvaluation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		match Match
	}{
		{name: "strict", match: Match{Strict: "desired.statement"}},
		{name: "regexp", match: Match{Regexp: `"message":"d[esired]{6}.statement"`}},
		{name: "expr", match: Match{Expr: "message == 'desired.statement' && ExprEnv['field.one'] == 'field.one.value' && field_two contains 'two.value'"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, appendPattern := range []bool{true, false} {
				t.Run(fmt.Sprintf("append_%v", appendPattern), func(t *testing.T) {
					match := tc.match
					match.Record = &LogRecord{
						Body:          "desired body content",
						AppendPattern: appendPattern,
						Attributes: map[string]string{
							"attr.one": "attr.one.value", "attr.two": "attr.two.value",
						},
					}
					for _, status := range discovery.StatusTypes {
						match.Status = status
						t.Run(string(status), func(t *testing.T) {
							for _, firstOnly := range []bool{true, false} {
								match.FirstOnly = firstOnly
								t.Run(fmt.Sprintf("FirstOnly:%v", firstOnly), func(t *testing.T) {
									observerID := component.MustNewIDWithName("an_observer", "observer.name")
									cfg := &Config{
										Receivers: map[component.ID]ReceiverEntry{
											component.MustNewIDWithName("a_receiver", "receiver.name"): {
												Rule:   "a.rule",
												Status: &Status{Statements: []Match{match}},
											},
										},
										WatchObservers: []component.ID{observerID},
									}
									require.NoError(t, cfg.Validate())

									plogs := make(chan plog.Logs)

									// If debugging tests, replace the Nop Logger with a test instance to see
									// all statements. Not in regular use to avoid spamming output.
									// logger := zaptest.NewLogger(t)
									logger := zap.NewNop()
									cStore := newCorrelationStore(logger, time.Hour)
									cStore.UpdateEndpoint(
										observer.Endpoint{ID: "endpoint.id"},
										addedState, observerID,
									)

									se, err := newStatementEvaluator(logger, component.MustNewID("some_type"), cfg, plogs, cStore)
									require.NoError(t, err)

									evaluatedLogger := se.evaluatedLogger.With(
										zap.String("name", `a_receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`),
									)

									numExpected := 1
									if !firstOnly {
										numExpected = 3
									}

									emitted := plog.NewLogs()
									wg := sync.WaitGroup{}
									wg.Add(numExpected)

									go func() {
										for i := 0; i < numExpected; i++ {
											logs := <-plogs
											if emitted.LogRecordCount() == 0 {
												emitted = logs
											} else {
												logs.ResourceLogs().MoveAndAppendTo(emitted.ResourceLogs())
											}
											wg.Done()
										}
									}()

									for _, statement := range []string{
										"undesired.statement",
										"another.undesired.statement",
										"desired.statement",
										"desired.statement",
										"desired.statement",
									} {
										evaluatedLogger.Info(
											statement,
											zap.String("field.one", "field.one.value"),
											zap.String("field_two", "field.two.value"),
										)
									}

									require.Eventually(t, func() bool {
										wg.Wait()
										return true
									}, 1*time.Second, time.Millisecond)
									close(plogs)

									for i := 0; i < numExpected; i++ {
										rl := emitted.ResourceLogs().At(i)
										rAttrs := rl.Resource().Attributes()
										require.Equal(t, map[string]any{
											"discovery.endpoint.id":   "endpoint.id",
											"discovery.event.type":    "statement.match",
											"discovery.observer.id":   "an_observer/observer.name",
											"discovery.receiver.name": "receiver.name",
											"discovery.receiver.rule": "a.rule",
											"discovery.receiver.type": "a_receiver",
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

										require.Equal(t, map[string]any{
											"discovery.status": string(status),
											"name":             `a_receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`,
											"attr.one":         "attr.one.value",
											"attr.two":         "attr.two.value",
											"field.one":        "field.one.value",
											"field_two":        "field.two.value",
										}, lrAttrs)

										expected := "desired body content"
										if match.Record.AppendPattern {
											if match.Strict != "" {
												expected = fmt.Sprintf("%s (evaluated \"desired.statement\")", expected)
											} else {
												expected = fmt.Sprintf("%s (evaluated \"{\\\"field.one\\\":\\\"field.one.value\\\",\\\"field_two\\\":\\\"field.two.value\\\",\\\"message\\\":\\\"desired.statement\\\"}\")", expected)
											}
										}
										require.Equal(t, expected, lr.Body().AsString())
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
