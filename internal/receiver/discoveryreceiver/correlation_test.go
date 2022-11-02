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
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap/zaptest"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestNewCorrelationStore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cs := newCorrelationStore(logger, time.Hour)
	cStore, ok := cs.(*store)
	require.True(t, ok)
	require.NotNil(t, cStore)
	require.Same(t, logger, cStore.logger)
	require.Equal(t, time.Hour, cStore.ttl)
	require.Equal(t, 30*time.Second, cStore.reapInterval)
	require.NotNil(t, time.Hour, cStore.receiverAttrs)
	require.NotNil(t, time.Hour, cStore.correlations)
	require.NotNil(t, time.Hour, cStore.endpointLocks)
	require.NotNil(t, time.Hour, cStore.receiverLocks)
	require.NotZero(t, cStore.sentinel)
}

func TestGetOrCreateUndiscoveredReceiver(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	createdCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), createdCorr.receiverID)
	require.Equal(t, config.NewComponentID(""), createdCorr.observerID)
	require.Zero(t, createdCorr.endpoint)
	require.Empty(t, createdCorr.lastState)
	require.Zero(t, createdCorr.lastUpdated)

	createdCorr.observerID = config.NewComponentID("an.observer")
	gotCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, gotCorr)
	// all returned correlations are copies whose mutations don't persist in storage
	require.NotSame(t, createdCorr, gotCorr)
	require.Equal(t, config.NewComponentID(""), gotCorr.observerID)
}

func TestGetOrCreateDiscoveredReceiver(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := config.NewComponentIDWithName("observer", "name")
	now := time.Now()
	cs.UpdateEndpoint(endpoint, "a.state", observerID)

	corr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, corr)
	require.Equal(t, config.NewComponentID(""), corr.receiverID)
	require.Equal(t, observerID, corr.observerID)
	require.Equal(t, endpoint, corr.endpoint)
	require.Equal(t, endpointState("a.state"), corr.lastState)
	require.GreaterOrEqual(t, corr.lastUpdated, now)

	receiverID := config.NewComponentIDWithName("receiver", "name")
	typedReceiverCorr := cs.GetOrCreate(receiverID, endpointID)
	require.NotNil(t, typedReceiverCorr)
	require.Equal(t, receiverID, typedReceiverCorr.receiverID)
	require.Equal(t, observerID, typedReceiverCorr.observerID)
	require.Equal(t, endpoint, typedReceiverCorr.endpoint)
	require.Equal(t, endpointState("a.state"), typedReceiverCorr.lastState)
	require.GreaterOrEqual(t, typedReceiverCorr.lastUpdated, now)
}

func TestGetOrCreateLaterDiscoveredReceiver(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	createdCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), createdCorr.receiverID)
	require.Equal(t, config.NewComponentID(""), createdCorr.observerID)
	require.Zero(t, createdCorr.endpoint)
	require.Empty(t, createdCorr.lastState)
	require.Zero(t, createdCorr.lastUpdated)

	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := config.NewComponentIDWithName("observer", "name")
	now := time.Now()
	cs.UpdateEndpoint(endpoint, "a.state", observerID)

	gotCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), gotCorr.receiverID)
	require.Equal(t, observerID, gotCorr.observerID)
	require.Equal(t, endpoint, gotCorr.endpoint)
	require.Equal(t, endpointState("a.state"), gotCorr.lastState)
	require.GreaterOrEqual(t, gotCorr.lastUpdated, now)

	receiverID := config.NewComponentIDWithName("receiver", "name")
	typedReceiverCorr := cs.GetOrCreate(receiverID, endpointID)
	require.NotNil(t, typedReceiverCorr)
	require.Equal(t, receiverID, typedReceiverCorr.receiverID)
	require.Equal(t, observerID, typedReceiverCorr.observerID)
	require.Equal(t, endpoint, typedReceiverCorr.endpoint)
	require.Equal(t, endpointState("a.state"), typedReceiverCorr.lastState)
	require.GreaterOrEqual(t, typedReceiverCorr.lastUpdated, now)
}

func TestGetOrCreateLaterDiscoveredReceiverWithUpdatedEndpoint(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	createdCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), createdCorr.receiverID)
	require.Equal(t, config.NewComponentID(""), createdCorr.observerID)
	require.Zero(t, createdCorr.endpoint)
	require.Empty(t, createdCorr.lastState)
	require.Zero(t, createdCorr.lastUpdated)

	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := config.NewComponentIDWithName("observer", "name")
	now := time.Now()
	cs.UpdateEndpoint(endpoint, "a.state", observerID)

	gotCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), gotCorr.receiverID)
	require.Equal(t, observerID, gotCorr.observerID)
	require.Equal(t, endpoint, gotCorr.endpoint)
	require.Equal(t, endpointState("a.state"), gotCorr.lastState)
	require.GreaterOrEqual(t, gotCorr.lastUpdated, now)

	now = time.Now()
	cs.UpdateEndpoint(endpoint, "another.state", observerID)

	receiverID := config.NewComponentIDWithName("receiver", "name")
	typedReceiverCorr := cs.GetOrCreate(receiverID, endpointID)
	require.NotNil(t, typedReceiverCorr)
	require.Equal(t, receiverID, typedReceiverCorr.receiverID)
	require.Equal(t, observerID, typedReceiverCorr.observerID)
	require.Equal(t, endpoint, typedReceiverCorr.endpoint)
	require.Equal(t, endpointState("another.state"), typedReceiverCorr.lastState)
	require.GreaterOrEqual(t, typedReceiverCorr.lastUpdated, now)

	// confirm state change propagates to other receivers
	noTypedReceiverCorr := cs.GetOrCreate(discovery.NoType, endpointID)
	require.NotNil(t, createdCorr)
	require.Equal(t, config.NewComponentID(""), noTypedReceiverCorr.receiverID)
	require.Equal(t, observerID, noTypedReceiverCorr.observerID)
	require.Equal(t, endpoint, noTypedReceiverCorr.endpoint)
	require.Equal(t, endpointState("another.state"), noTypedReceiverCorr.lastState)
	require.GreaterOrEqual(t, noTypedReceiverCorr.lastUpdated, now)
}

func TestAttrs(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	receiverID := config.NewComponentIDWithName("receiver", "name")
	attrs := cs.Attrs(receiverID)
	require.Empty(t, attrs)
	// all returned attrs are copies and don't persist changes in storage
	attrs["key"] = "value"
	require.Empty(t, cs.Attrs(receiverID))

	attrs["another.key"] = "another.value"
	cs.UpdateAttrs(receiverID, attrs)
	updated := cs.Attrs(receiverID)
	require.Equal(t, map[string]string{"key": "value", "another.key": "another.value"}, updated)

	cs.UpdateAttrs(receiverID, map[string]string{"key": "changed.value", "yet.another.key": "yet.another.value"})
	updated = cs.Attrs(receiverID)
	require.Equal(t, map[string]string{
		"key":             "changed.value",
		"another.key":     "another.value",
		"yet.another.key": "yet.another.value",
	}, updated)
}

func TestReaperLoop(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Nanosecond)
	cStore, ok := cs.(*store)
	require.True(t, ok)
	require.NotNil(t, cStore)
	// update the reapInterval so as not to wait 30 seconds for test logic
	cStore.reapInterval = time.Millisecond

	endpointID := observer.EndpointID("an.endpoint")
	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := config.NewComponentIDWithName("observer", "name")

	cs.Start()
	t.Cleanup(cs.Stop)

	cs.UpdateEndpoint(endpoint, addedState, observerID)

	receiverID := config.NewComponentIDWithName("receiver", "name")
	corr := cs.GetOrCreate(receiverID, endpointID)
	require.Equal(t, addedState, corr.lastState)
	rMap, ok := cStore.correlations.Load(endpointID)
	require.True(t, ok)
	receiverMap, isMap := rMap.(*sync.Map)
	require.True(t, isMap)

	noTypeCorr, containsNoType := receiverMap.Load(discovery.NoType)
	require.True(t, containsNoType)
	noTypedReceiverCorr, isCorr := noTypeCorr.(*correlation)
	require.True(t, isCorr)
	require.Equal(t, addedState, noTypedReceiverCorr.lastState)

	cs.UpdateEndpoint(endpoint, removedState, observerID)

	// repeat check once to ensure loop-driven removal
	_, hasEndpoint := cStore.correlations.Load(endpointID)
	require.True(t, hasEndpoint)
	require.Equal(t, removedState, noTypedReceiverCorr.lastState)

	// confirm reaping occurs within interval
	require.Eventually(t, func() bool {
		_, hasCorrelations := cStore.correlations.Load(endpointID)
		return !hasCorrelations
	}, 100*time.Millisecond, time.Millisecond) // windows test seems to require more time.
}
