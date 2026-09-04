// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

func newTestObserverHandler(t *testing.T, splunkHome string, factory *mockSubReceiverFactory) *observerHandler {
	t.Helper()
	options := newFactoryOptions(WithSubReceiver(factory))
	settings := receiver.Settings{
		ID:                component.MustNewID("splunk_inputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
	return newObserverHandler(splunkHome, options, settings, testNopConsumer{})
}

func TestObserverHandlerOnAdd(t *testing.T) {
	t.Run("adds_ta_to_active", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		require.NoError(t, h.OnAdd(context.Background(), []string{taDir}))

		assert.Contains(t, h.active, taDir)
	})

	t.Run("is_idempotent", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		require.NoError(t, h.OnAdd(context.Background(), []string{taDir}))
		require.NoError(t, h.OnAdd(context.Background(), []string{taDir}))

		// second call should be a no-op — only one receiver created
		assert.Len(t, factory.receivers, 1)
	})

	t.Run("adds_multiple_tas", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		ta1 := makeTA(t, splunkHome, "splunk_ta_syslog")
		ta2 := makeTA(t, splunkHome, "splunk_ta_auth")

		require.NoError(t, h.OnAdd(context.Background(), []string{ta1, ta2}))

		assert.Contains(t, h.active, ta1)
		assert.Contains(t, h.active, ta2)
	})
}

func TestObserverHandlerOnRemove(t *testing.T) {
	t.Run("shuts_down_receivers_and_removes_from_active", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		// seed active directly with a mock receiver
		mock := &mockReceiver{}
		h.active[taDir] = []receiver.Logs{mock}

		h.OnRemove(context.Background(), []string{taDir})

		assert.NotContains(t, h.active, taDir)
		assert.Equal(t, 1, mock.shutdownCount)
	})

	t.Run("removing_unknown_ta_is_a_noop", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		// should not panic or error
		h.OnRemove(context.Background(), []string{"/nonexistent/ta"})

		assert.Empty(t, h.active)
	})
}

func TestObserverHandlerOnChange(t *testing.T) {
	t.Run("shuts_down_old_and_starts_new_receiver", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		old := &mockReceiver{}
		h.active[taDir] = []receiver.Logs{old}

		require.NoError(t, h.OnChange(context.Background(), []string{taDir}))

		assert.Equal(t, 1, old.shutdownCount, "old receiver should be shut down")
		assert.Contains(t, h.active, taDir, "TA should be re-added")
	})
}

func TestObserverHandlerShutdown(t *testing.T) {
	t.Run("shuts_down_all_active_receivers", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		mock1, mock2 := &mockReceiver{}, &mockReceiver{}
		h.active["ta1"] = []receiver.Logs{mock1}
		h.active["ta2"] = []receiver.Logs{mock2}

		h.shutdown(context.Background())

		assert.Empty(t, h.active)
		assert.Equal(t, 1, mock1.shutdownCount)
		assert.Equal(t, 1, mock2.shutdownCount)
	})

	t.Run("shutdown_on_empty_handler_is_a_noop", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		h := newTestObserverHandler(t, splunkHome, factory)

		// should not panic
		h.shutdown(context.Background())

		assert.Empty(t, h.active)
	})
}
