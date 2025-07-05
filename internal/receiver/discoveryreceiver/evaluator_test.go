// Copyright Splunk, Inc.
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
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func setup(_ *testing.T) (*evaluator, component.ID, observer.EndpointID) {
	// If debugging tests, replace the Nop Logger with a test instance to see
	// all statements. Not in regular use to avoid spamming output.
	// logger := zaptest.NewLogger(t)
	logger := zap.NewNop()
	alreadyLogged := &sync.Map{}
	eval := &evaluator{
		logger:        logger,
		config:        &Config{},
		correlations:  newCorrelationStore(logger, time.Hour),
		alreadyLogged: alreadyLogged,
		exprEnv: func(pattern string) map[string]any {
			return map[string]any{"item": pattern}
		},
	}

	receiverID := component.MustNewIDWithName("type", "name")
	endpointID := observer.EndpointID("endpoint")
	return eval, receiverID, endpointID
}

func TestEvaluateMatch(t *testing.T) {
	eval, receiverID, endpointID := setup(t)
	anotherReceiverID := component.MustNewIDWithName("type", "another.name")

	for _, tc := range []struct {
		typ string
		m   Match
	}{
		{typ: "strict", m: Match{Strict: "must.match"}},
		{typ: "regexp", m: Match{Regexp: "must.*"}},
		{typ: "expr", m: Match{Expr: "item == 'must.match'"}},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			shouldLog, err := eval.evaluateMatch(tc.m, "must.match", "some.status", receiverID, endpointID)
			require.NoError(t, err)
			require.True(t, shouldLog)

			shouldLog, err = eval.evaluateMatch(tc.m, "must.match", "some.status", receiverID, endpointID)
			require.NoError(t, err)
			require.False(t, shouldLog)

			shouldLog, err = eval.evaluateMatch(tc.m, "must.match", "some.status", anotherReceiverID, endpointID)
			require.NoError(t, err)
			require.True(t, shouldLog)
		})
	}
}

func TestEvaluateInvalidMatch(t *testing.T) {
	eval, receiverID, endpointID := setup(t)

	for _, tc := range []struct {
		typ           string
		expectedError string
		m             Match
	}{
		{typ: "regexp", m: Match{Regexp: "*"}, expectedError: "invalid match regexp statement: error parsing regexp: missing argument to repetition operator: `*`"},
		{typ: "expr", m: Match{Expr: "not_a_thing"}, expectedError: "invalid match expr statement: unknown name not_a_thing (1:1)\n | not_a_thing\n | ^"},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			shouldLog, err := eval.evaluateMatch(tc.m, "a.pattern", "some.status", receiverID, endpointID)
			require.EqualError(t, err, tc.expectedError)
			require.False(t, shouldLog)
		})
	}
}

func TestCorrelateResourceAttrs(t *testing.T) {
	for _, embed := range []bool{false, true} {
		t.Run(fmt.Sprintf("embed-%v", embed), func(t *testing.T) {
			eval, _, endpointID := setup(t)
			eval.config.EmbedReceiverConfig = embed

			endpoint := observer.Endpoint{ID: endpointID}
			observerID := component.MustNewIDWithName("type", "name")
			receiverID := component.MustNewIDWithName("receiver", "name")
			eval.correlations.UpdateEndpoint(endpoint, receiverID, observerID)

			corr := eval.correlations.GetOrCreate(endpointID, receiverID)

			cfg := &Config{
				Receivers: map[component.ID]ReceiverEntry{
					receiverID: {
						ServiceType: "a_service",
						Rule:        mustNewRule(`type == "container"`),
						Config: map[string]any{
							"config_option": "val",
						},
						ResourceAttributes: map[string]string{
							"one": "one.val",
							"two": "2",
						},
					},
				},
			}

			to := map[string]string{}
			require.Empty(t, eval.correlations.Attrs(endpointID))
			eval.correlateResourceAttributes(cfg, to, corr)

			expectedResourceAttrs := map[string]string{
				"service.type":          "a_service",
				"discovery.observer.id": "type/name",
			}

			if embed {
				expectedResourceAttrs["discovery.receiver.config"] = base64.StdEncoding.EncodeToString([]byte(`receivers:
  receiver/name:
    config:
      config_option: val
    resource_attributes:
      one: one.val
      two: "2"
    rule: type == "container"
watch_observers:
- type/name
`))
			}

			require.Equal(t, expectedResourceAttrs, to)
		})
	}
}
