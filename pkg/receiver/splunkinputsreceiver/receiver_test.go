// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type testNopConsumer struct{}

func (testNopConsumer) Capabilities() consumer.Capabilities          { return consumer.Capabilities{} }
func (testNopConsumer) ConsumeLogs(context.Context, plog.Logs) error { return nil }

// mockReceiver is a mock receiver.Logs that tracks Start and Shutdown calls.
type mockReceiver struct {
	startCount    int
	shutdownCount int
}

func (r *mockReceiver) Start(context.Context, component.Host) error {
	r.startCount++
	return nil
}

func (r *mockReceiver) Shutdown(context.Context) error {
	r.shutdownCount++
	return nil
}

// mockSubReceiverFactory is a mock SubReceiverFactory that returns a mockReceiver per request.
type mockSubReceiverFactory struct {
	receivers map[string]*mockReceiver // keyed by BaseDir (taDir)
}

func newMockFactory() *mockSubReceiverFactory {
	return &mockSubReceiverFactory{receivers: map[string]*mockReceiver{}}
}

func (f *mockSubReceiverFactory) Scheme() string { return "monitor" }

func (f *mockSubReceiverFactory) CreateLogs(_ context.Context, _ receiver.Settings, req ReceiverRequest, _ consumer.Logs) (receiver.Logs, error) {
	r := &mockReceiver{}
	f.receivers[req.BaseDir] = r
	return r, nil
}

// makeTA creates a minimal TA directory with an inputs.conf under splunkHome and returns the TA dir path.
func makeTA(t *testing.T, splunkHome, taName string) string {
	t.Helper()
	taDefaultDir := filepath.Join(splunkHome, "etc", "apps", taName, "default")
	require.NoError(t, os.MkdirAll(taDefaultDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(taDefaultDir, "inputs.conf"),
		[]byte("[monitor:///var/log/syslog]\nsourcetype = syslog\n"),
		0o600,
	))
	return filepath.Join(splunkHome, "etc", "apps", taName)
}

// newTestSplunkInputsReceiver builds a splunkInputsReceiver wired with the given mock factory
// and a real (but idle) fsnotify watcher so reconcile can call watcher.Add without panicking.
func newTestSplunkInputsReceiver(t *testing.T, splunkHome string, factory *mockSubReceiverFactory) *splunkInputsReceiver {
	t.Helper()
	options := newFactoryOptions(WithSubReceiver(factory))
	settings := receiver.Settings{
		ID:                component.MustNewID("splunk_inputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	t.Cleanup(func() { _ = watcher.Close() })
	r := newSplunkInputsReceiver(splunkHome, options, settings, testNopConsumer{})
	r.watcher = watcher
	return r
}

func TestShutdownWithoutStart(t *testing.T) {
	splunkHome := t.TempDir()
	factory := newMockFactory()
	options := newFactoryOptions(WithSubReceiver(factory))
	settings := receiver.Settings{
		ID:                component.MustNewID("splunk_inputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
	r := newSplunkInputsReceiver(splunkHome, options, settings, testNopConsumer{})

	// Start was never called — watcher is nil, doneCh was never started.
	// Shutdown must return without blocking.
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestReconcile(t *testing.T) {
	t.Run("adds_new_ta", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		makeTA(t, splunkHome, "splunk_ta_syslog")

		r.reconcile(context.Background(), map[string]struct{}{})

		assert.Len(t, r.handler.active, 1)
	})

	t.Run("removes_deleted_ta", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		// seed active with a running receiver for the TA
		mock := &mockReceiver{}
		r.handler.active[taDir] = []receiver.Logs{mock}

		// delete the TA from disk
		require.NoError(t, os.RemoveAll(filepath.Join(splunkHome, "etc", "apps", "splunk_ta_syslog")))

		r.reconcile(context.Background(), map[string]struct{}{})

		assert.Empty(t, r.handler.active)
		assert.Equal(t, 1, mock.shutdownCount)
	})

	t.Run("reloads_changed_ta", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		mock := &mockReceiver{}
		r.handler.active[taDir] = []receiver.Logs{mock}

		r.reconcile(context.Background(), map[string]struct{}{taDir: {}})

		assert.Equal(t, 1, mock.shutdownCount, "old receiver should be shut down on change")
		assert.Contains(t, r.handler.active, taDir, "TA should be re-added after reload")
	})

	t.Run("system_change_reloads_all_tas", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		ta1 := makeTA(t, splunkHome, "splunk_ta_syslog")
		ta2 := makeTA(t, splunkHome, "splunk_ta_auth")

		mock1, mock2 := &mockReceiver{}, &mockReceiver{}
		r.handler.active[ta1] = []receiver.Logs{mock1}
		r.handler.active[ta2] = []receiver.Logs{mock2}

		// "" sentinel means a system conf change — all TAs should reload
		r.reconcile(context.Background(), map[string]struct{}{"": {}})

		assert.Equal(t, 1, mock1.shutdownCount, "ta1 should be reloaded on system change")
		assert.Equal(t, 1, mock2.shutdownCount, "ta2 should be reloaded on system change")
		assert.Contains(t, r.handler.active, ta1)
		assert.Contains(t, r.handler.active, ta2)
	})

	t.Run("watches_ta_subdirs_on_add", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		r.reconcile(context.Background(), map[string]struct{}{})

		watched := r.watcher.WatchList()
		assert.Contains(t, watched, taDir, "TA dir itself should be watched")
		assert.Contains(t, watched, filepath.Join(taDir, "default"), "TA default/ should be watched")
	})

	t.Run("retries_watching_ta_subdirs_on_change", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

		mock := &mockReceiver{}
		r.handler.active[taDir] = []receiver.Logs{mock}

		// create local/ after initial setup — simulates it being added later
		localDir := filepath.Join(taDir, "local")
		require.NoError(t, os.MkdirAll(localDir, 0o755))

		r.reconcile(context.Background(), map[string]struct{}{taDir: {}})

		watched := r.watcher.WatchList()
		assert.Contains(t, watched, localDir, "local/ should be watched after it is created")
	})

	t.Run("watches_apps_dir_after_late_creation", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		// appsDir does not exist yet — watcher.Add silently failed at Start time
		appsDir := filepath.Join(splunkHome, "etc", "apps")

		// create appsDir with a TA — simulates apps/ being created after Start
		makeTA(t, splunkHome, "splunk_ta_syslog")

		r.reconcile(context.Background(), map[string]struct{}{"": {}})

		assert.Contains(t, r.watcher.WatchList(), appsDir, "apps/ should be watched after it is created")
		assert.Len(t, r.handler.active, 1, "TA should be discovered and started")
	})

	t.Run("unrelated_ta_is_not_reloaded", func(t *testing.T) {
		splunkHome := t.TempDir()
		factory := newMockFactory()
		r := newTestSplunkInputsReceiver(t, splunkHome, factory)

		ta1 := makeTA(t, splunkHome, "splunk_ta_syslog")
		ta2 := makeTA(t, splunkHome, "splunk_ta_auth")

		mock1, mock2 := &mockReceiver{}, &mockReceiver{}
		r.handler.active[ta1] = []receiver.Logs{mock1}
		r.handler.active[ta2] = []receiver.Logs{mock2}

		// only ta1 is in pending
		r.reconcile(context.Background(), map[string]struct{}{ta1: {}})

		assert.Equal(t, 1, mock1.shutdownCount, "ta1 should be reloaded")
		assert.Equal(t, 0, mock2.shutdownCount, "ta2 should not be touched")
	})
}

func TestTADirFromPath(t *testing.T) {
	appsDir := filepath.Join("/opt", "splunkforwarder", "etc", "apps")

	tests := []struct {
		name      string
		eventPath string
		want      string
	}{
		{
			name:      "file inside TA default dir",
			eventPath: filepath.Join(appsDir, "splunk_ta_syslog", "default", "inputs.conf"),
			want:      filepath.Join(appsDir, "splunk_ta_syslog"),
		},
		{
			name:      "file inside TA local dir",
			eventPath: filepath.Join(appsDir, "splunk_ta_syslog", "local", "inputs.conf"),
			want:      filepath.Join(appsDir, "splunk_ta_syslog"),
		},
		{
			name:      "appsDir itself",
			eventPath: appsDir,
			want:      "",
		},
		{
			name:      "deeply nested file",
			eventPath: filepath.Join(appsDir, "splunk_ta_syslog", "default", "subdir", "file.conf"),
			want:      filepath.Join(appsDir, "splunk_ta_syslog"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := taDirFromPath(tc.eventPath, appsDir)
			require.Equal(t, tc.want, got)
		})
	}
}
