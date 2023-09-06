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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func setup() (*evaluator, component.ID, observer.EndpointID) {
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

	receiverID := component.NewIDWithName("type", "name")
	endpointID := observer.EndpointID("endpoint")
	return eval, receiverID, endpointID
}

func TestEvaluateMatch(t *testing.T) {
	eval, receiverID, endpointID := setup()
	anotherReceiverID := component.NewIDWithName("type", "another.name")

	for _, tc := range []struct {
		typ string
		m   Match
	}{
		{typ: "strict", m: Match{Strict: "must.match"}},
		{typ: "regexp", m: Match{Regexp: "must.*"}},
		{typ: "expr", m: Match{Expr: "item == 'must.match'"}},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			tc.m.FirstOnly = true
			shouldLog, err := eval.evaluateMatch(tc.m, "must.match", "some.status", receiverID, endpointID)
			require.NoError(t, err)
			require.True(t, shouldLog)

			shouldLog, err = eval.evaluateMatch(tc.m, "must.match", "some.status", receiverID, endpointID)
			require.NoError(t, err)
			require.False(t, shouldLog)

			shouldLog, err = eval.evaluateMatch(tc.m, "must.match", "some.status", anotherReceiverID, endpointID)
			require.NoError(t, err)
			require.True(t, shouldLog)

			tc.m.FirstOnly = false
			shouldLog, err = eval.evaluateMatch(tc.m, "must.match", "some.status", receiverID, endpointID)
			require.NoError(t, err)
			require.True(t, shouldLog)

			shouldLog, err = eval.evaluateMatch(tc.m, "doesn't.match", "another.status", receiverID, endpointID)
			require.NoError(t, err)
			require.False(t, shouldLog)
		})
	}
}

func TestEvaluateInvalidMatch(t *testing.T) {
	eval, receiverID, endpointID := setup()

	for _, tc := range []struct {
		typ           string
		expectedError string
		m             Match
	}{
		{typ: "regexp", m: Match{Regexp: "*"}, expectedError: "invalid match regexp statement: error parsing regexp: missing argument to repetition operator: `*`"},
		{typ: "expr", m: Match{Expr: "not_a_thing"}, expectedError: "invalid match expr statement: unknown name not_a_thing (1:1)\n | not_a_thing\n | ^"},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			tc.m.FirstOnly = true
			shouldLog, err := eval.evaluateMatch(tc.m, "a.pattern", "some.status", receiverID, endpointID)
			require.EqualError(t, err, tc.expectedError)
			require.False(t, shouldLog)
		})
	}
}

func TestCorrelateResourceAttrs(t *testing.T) {
	for _, embed := range []bool{false, true} {
		t.Run(fmt.Sprintf("embed-%v", embed), func(t *testing.T) {
			eval, _, endpointID := setup()
			eval.config.EmbedReceiverConfig = embed

			endpoint := observer.Endpoint{ID: endpointID}
			observerID := component.NewIDWithName("type", "name")
			eval.correlations.UpdateEndpoint(endpoint, addedState, observerID)

			corr := eval.correlations.GetOrCreate(discovery.NoType, endpointID)

			from := pcommon.NewMap()
			from.FromRaw(
				map[string]interface{}{
					"one": "one.val",
					"two": 2,
				})

			to := pcommon.NewMap()

			require.Empty(t, eval.correlations.Attrs(discovery.NoType))
			eval.correlateResourceAttributes(from, to, corr)

			expectedResourceAttrs := map[string]any{
				"one":                   "one.val",
				"two":                   int64(2),
				"discovery.observer.id": "type/name",
			}

			encodedWatchObserver := base64.StdEncoding.EncodeToString([]byte("watch_observers:\n- type/name\n"))
			if embed {
				expectedResourceAttrs["discovery.receiver.config"] = encodedWatchObserver
			}

			require.Equal(t, expectedResourceAttrs, to.AsRaw())

			attrs := eval.correlations.Attrs(discovery.NoType)

			expectedAttrs := map[string]string{}
			if embed {
				expectedAttrs["discovery.receiver.updated.config.type/name"] = encodedWatchObserver
			}

			require.Equal(t, expectedAttrs, attrs)
		})
	}
}

func TestCorrelateResourceAttrsWithExistingConfig(t *testing.T) {
	for _, embed := range []bool{false, true} {
		t.Run(fmt.Sprintf("embed-%v", embed), func(t *testing.T) {
			eval, _, endpointID := setup()
			eval.config.EmbedReceiverConfig = embed

			endpoint := observer.Endpoint{ID: endpointID}
			observerID := component.NewIDWithName("type", "name")
			eval.correlations.UpdateEndpoint(endpoint, addedState, observerID)

			corr := eval.correlations.GetOrCreate(discovery.NoType, endpointID)

			encodedConfig := base64.StdEncoding.EncodeToString([]byte("config: some config\nrule: some rule\n"))

			from := pcommon.NewMap()
			from.FromRaw(
				map[string]interface{}{
					"discovery.receiver.config": encodedConfig,
					"one":                       "one.val",
					"two":                       2,
				})

			to := pcommon.NewMap()

			require.Empty(t, eval.correlations.Attrs(discovery.NoType))
			eval.correlateResourceAttributes(from, to, corr)

			var receiverConfig string
			if embed {
				receiverConfig = base64.StdEncoding.EncodeToString([]byte("config: some config\nrule: some rule\nwatch_observers:\n- type/name\n"))
			} else {
				receiverConfig = encodedConfig
			}

			expectedResourceAttrs := map[string]any{
				"one":                       "one.val",
				"two":                       int64(2),
				"discovery.observer.id":     "type/name",
				"discovery.receiver.config": receiverConfig,
			}

			require.Equal(t, expectedResourceAttrs, to.AsRaw())

			attrs := eval.correlations.Attrs(discovery.NoType)
			expectedAttrs := map[string]string{}

			if embed {
				expectedAttrs["discovery.receiver.updated.config.type/name"] = receiverConfig
			}

			require.Equal(t, expectedAttrs, attrs)
		})
	}
}
