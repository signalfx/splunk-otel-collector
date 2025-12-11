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
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap/zaptest"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestNewCorrelationStore(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cStore := newCorrelationStore(logger, time.Hour)
	require.NotNil(t, cStore)
	require.Same(t, logger, cStore.logger)
	require.Equal(t, time.Hour, cStore.ttl)
	require.Equal(t, 30*time.Second, cStore.reapInterval)
	require.NotNil(t, time.Hour, cStore.attrs)
	require.NotNil(t, time.Hour, cStore.correlations)
	require.NotNil(t, time.Hour, cStore.endpointLocks)
	require.NotZero(t, cStore.sentinel)
}

func TestGetOrCreateUndiscoveredReceiver(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	createdCorr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, createdCorr)
	require.Equal(t, discovery.NoType, createdCorr.receiverID)
	require.Equal(t, discovery.NoType, createdCorr.observerID)
	require.Zero(t, createdCorr.endpoint)
	require.False(t, createdCorr.stale)
	require.Zero(t, createdCorr.lastUpdated)

	createdCorr.observerID = component.MustNewID("an_observer")
	gotCorr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, gotCorr)
	// all returned correlations are copies whose mutations don't persist in storage
	require.NotEqual(t, createdCorr, gotCorr)
	require.Equal(t, discovery.NoType, gotCorr.observerID)
}

func TestGetOrCreateDiscoveredReceiver(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := component.MustNewIDWithName("observer", "name")
	now := time.Now()
	receiverID := component.MustNewIDWithName("receiver", "name")
	cs.UpdateEndpoint(endpoint, receiverID, observerID)

	corr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, corr)
	require.Equal(t, receiverID, corr.receiverID)
	require.Equal(t, observerID, corr.observerID)
	require.Equal(t, endpoint, corr.endpoint)
	require.False(t, corr.stale)
	require.GreaterOrEqual(t, corr.lastUpdated, now)

	typedReceiverCorr := cs.GetOrCreate(endpointID, receiverID)
	require.NotNil(t, typedReceiverCorr)
	require.Equal(t, receiverID, typedReceiverCorr.receiverID)
	require.Equal(t, observerID, typedReceiverCorr.observerID)
	require.Equal(t, endpoint, typedReceiverCorr.endpoint)
	require.False(t, typedReceiverCorr.stale)
	require.GreaterOrEqual(t, typedReceiverCorr.lastUpdated, now)
}

func TestGetOrCreateLaterDiscoveredReceiverWithUpdatedEndpoint(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	createdCorr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, createdCorr)
	require.Equal(t, discovery.NoType, createdCorr.receiverID)
	require.Equal(t, discovery.NoType, createdCorr.observerID)
	require.Zero(t, createdCorr.endpoint)
	require.False(t, createdCorr.stale)
	require.Zero(t, createdCorr.lastUpdated)

	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := component.MustNewIDWithName("observer", "name")
	now := time.Now()
	receiverID := component.MustNewIDWithName("receiver", "name")
	cs.UpdateEndpoint(endpoint, receiverID, observerID)

	gotCorr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, createdCorr)
	require.Equal(t, receiverID, gotCorr.receiverID)
	require.Equal(t, observerID, gotCorr.observerID)
	require.Equal(t, endpoint, gotCorr.endpoint)
	require.False(t, gotCorr.stale)
	require.GreaterOrEqual(t, gotCorr.lastUpdated, now)

	// confirm state is if receiverID isn't provided
	noTypedReceiverCorr := cs.GetOrCreate(endpointID, discovery.NoType)
	require.NotNil(t, createdCorr)
	require.Equal(t, receiverID, noTypedReceiverCorr.receiverID)
	require.Equal(t, observerID, noTypedReceiverCorr.observerID)
	require.Equal(t, endpoint, noTypedReceiverCorr.endpoint)
	require.False(t, noTypedReceiverCorr.stale)
	require.GreaterOrEqual(t, noTypedReceiverCorr.lastUpdated, now)
}

func TestAttrs(t *testing.T) {
	cs := newCorrelationStore(zaptest.NewLogger(t), time.Hour)
	endpointID := observer.EndpointID("an.endpoint")
	attrs := cs.Attrs(endpointID)
	require.Empty(t, attrs)
	// all returned attrs are copies and don't persist changes in storage
	attrs["key"] = "value"
	require.Empty(t, cs.Attrs(endpointID))

	attrs["another.key"] = "another.value"
	cs.UpdateAttrs(endpointID, attrs)
	updated := cs.Attrs(endpointID)
	require.Equal(t, map[string]string{"key": "value", "another.key": "another.value"}, updated)

	cs.UpdateAttrs(endpointID, map[string]string{"key": "changed.value", "yet.another.key": "yet.another.value"})
	updated = cs.Attrs(endpointID)
	require.Equal(t, map[string]string{
		"key":             "changed.value",
		"another.key":     "another.value",
		"yet.another.key": "yet.another.value",
	}, updated)
}

func TestReaperLoop(t *testing.T) {
	cStore := newCorrelationStore(zaptest.NewLogger(t), time.Nanosecond)
	require.NotNil(t, cStore)
	// update the reapInterval so as not to wait 30 seconds for test logic
	cStore.reapInterval = time.Millisecond

	endpointID := observer.EndpointID("an.endpoint")
	endpoint := observer.Endpoint{
		ID:     endpointID,
		Target: "a.target",
	}
	observerID := component.MustNewIDWithName("observer", "name")

	cStore.Start()
	t.Cleanup(cStore.Stop)

	receiverID := component.MustNewIDWithName("receiver", "name")
	cStore.UpdateEndpoint(endpoint, receiverID, observerID)

	corr := cStore.GetOrCreate(endpointID, receiverID)
	require.False(t, corr.stale)
	rMap, ok := cStore.correlations.Load(endpointID)
	require.True(t, ok)
	fetchedCorr, isCorr := rMap.(*correlation)
	require.True(t, isCorr)
	require.False(t, fetchedCorr.stale)

	cStore.MarkStale(endpointID)

	// repeat check once to ensure loop-driven removal
	_, hasEndpoint := cStore.correlations.Load(endpointID)
	require.True(t, hasEndpoint)
	require.True(t, fetchedCorr.stale)

	// confirm reaping occurs within interval
	require.Eventually(t, func() bool {
		_, hasCorrelations := cStore.correlations.Load(endpointID)
		return !hasCorrelations
	}, 100*time.Millisecond, time.Millisecond) // windows test seems to require more time.
}
