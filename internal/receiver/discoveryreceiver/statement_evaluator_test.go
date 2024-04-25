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
							observerID := component.MustNewIDWithName("an_observer", "observer.name")
							cfg := &Config{
								Receivers: map[component.ID]ReceiverEntry{
									component.MustNewIDWithName("a_receiver", "receiver.name"): {
										Rule:   mustNewRule(`type == "container"`),
										Status: &Status{Statements: []Match{match}},
									},
								},
								WatchObservers: []component.ID{observerID},
							}
							require.NoError(t, cfg.Validate())

							// If debugging tests, replace the Nop Logger with a test instance to see
							// all statements. Not in regular use to avoid spamming output.
							// logger := zaptest.NewLogger(t)
							logger := zap.NewNop()
							cStore := newCorrelationStore(logger, time.Hour)

							emitCh := cStore.EmitCh()
							emitWG := sync.WaitGroup{}
							emitWG.Add(1)
							var corr correlation
							go func() {
								corr = <-emitCh
								emitWG.Done()
							}()

							receiverID := component.MustNewIDWithName("a_receiver", "receiver.name")
							cStore.UpdateEndpoint(observer.Endpoint{ID: "endpoint.id"}, receiverID, observerID)

							se, err := newStatementEvaluator(logger, component.MustNewID("some_type"), cfg, cStore)
							require.NoError(t, err)

							evaluatedLogger := se.evaluatedLogger.With(
								zap.String("name", `a_receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`),
							)

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

							// wait for the emit channel to be processed
							emitWG.Wait()

							entityEvents, numFailed, err := entityStateEvents(corr.observerID,
								[]observer.Endpoint{corr.endpoint}, cStore, time.Now())
							require.NoError(t, err)
							require.Equal(t, 0, numFailed)
							emitted := entityEvents.ConvertAndMoveToLogs()

							require.Equal(t, 1, emitted.ResourceLogs().Len())
							rl := emitted.ResourceLogs().At(0)
							require.Equal(t, 0, rl.Resource().Attributes().Len())

							sLogs := rl.ScopeLogs()
							require.Equal(t, 1, sLogs.Len())
							sl := sLogs.At(0)
							lrs := sl.LogRecords()
							require.Equal(t, 1, lrs.Len())
							lr := sl.LogRecords().At(0)

							oea, ok := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
							require.True(t, ok)
							entityAttrs := oea.Map()

							// Validate "caller" attribute
							callerAttr, ok := entityAttrs.Get("caller")
							require.True(t, ok)
							_, expectedFile, _, _ := runtime.Caller(0)
							// runtime doesn't use os.PathSeparator
							splitPath := strings.Split(expectedFile, "/")
							expectedCaller := splitPath[len(splitPath)-1]
							require.Contains(t, callerAttr.Str(), expectedCaller)
							entityAttrs.Remove("caller")

							// Validate the rest of the attributes
							expectedMsg := "desired body content"
							if match.Record.AppendPattern {
								if match.Strict != "" {
									expectedMsg = fmt.Sprintf("%s (evaluated \"desired.statement\")", expectedMsg)
								} else {
									expectedMsg = fmt.Sprintf("%s (evaluated \"{\\\"field.one\\\":\\\"field.one.value\\\",\\\"field_two\\\":\\\"field.two.value\\\",\\\"message\\\":\\\"desired.statement\\\"}\")", expectedMsg)
								}
							}
							require.Equal(t, map[string]any{
								discovery.OtelEntityIDAttr: map[string]any{
									"discovery.endpoint.id": "endpoint.id",
								},
								discovery.OtelEntityEventTypeAttr: discovery.OtelEntityEventTypeState,
								discovery.OtelEntityAttributesAttr: map[string]any{
									"discovery.event.type":    "statement.match",
									"discovery.observer.id":   "an_observer/observer.name",
									"discovery.receiver.name": "receiver.name",
									"discovery.receiver.rule": `type == "container"`,
									"discovery.receiver.type": "a_receiver",
									"discovery.status":        string(status),
									"discovery.message":       expectedMsg,
									"name":                    `a_receiver/receiver.name/receiver_creator/rc.name/{endpoint=""}/endpoint.id`,
									"attr.one":                "attr.one.value",
									"attr.two":                "attr.two.value",
									"field.one":               "field.one.value",
									"field_two":               "field.two.value",
									"discovery.observer.name": "observer.name",
									"discovery.observer.type": "an_observer",
									"endpoint":                "",
								},
							}, lr.Attributes().AsRaw())
						})
					}
				})
			}
		})
	}
}
