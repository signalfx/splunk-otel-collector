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
	"errors"
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
			match := tc.match
			match.Record = &LogRecord{
				Body: "desired body content",
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
					go func() {
						<-emitCh
						emitWG.Done()
					}()

					receiverID := component.MustNewIDWithName("a_receiver", "receiver.name")
					endpointID := observer.EndpointID("endpoint.id")
					cStore.UpdateEndpoint(observer.Endpoint{ID: endpointID}, receiverID, observerID)

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
							zap.Error(errors.New("some error")),
						)
					}

					// wait for the emit channel to be processed
					emitWG.Wait()

					// Validate the attributes
					require.Equal(t, map[string]string{
						"discovery.observer.id":   "an_observer/observer.name",
						"discovery.receiver.name": "receiver.name",
						"discovery.receiver.rule": `type == "container"`,
						"discovery.receiver.type": "a_receiver",
						"discovery.status":        string(status),
						"discovery.message":       "desired body content",
						"discovery.matched_log":   "desired.statement (error: some error)",
						"attr.one":                "attr.one.value",
						"attr.two":                "attr.two.value",
					}, cStore.Attrs(endpointID))
				})
			}
		})
	}
}
