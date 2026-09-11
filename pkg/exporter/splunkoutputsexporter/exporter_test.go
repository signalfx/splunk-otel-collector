// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// fakeWatcher is an in-memory fileWatcher for unit tests.
// Close signals the watch loop to exit via a separate done channel so that
// in-flight events are not lost when the channels are closed.
type fakeWatcher struct {
	added  map[string]struct{}
	events chan fsnotify.Event
	errors chan error
	done   chan struct{}
}

func newFakeWatcher() *fakeWatcher {
	return &fakeWatcher{
		added:  map[string]struct{}{},
		events: make(chan fsnotify.Event, 1),
		errors: make(chan error, 1),
		done:   make(chan struct{}),
	}
}

func (f *fakeWatcher) Add(name string) error { f.added[name] = struct{}{}; return nil }
func (f *fakeWatcher) Close() error          { close(f.done); return nil }

func (f *fakeWatcher) Events() <-chan fsnotify.Event {
	// Return a channel that is closed when done is closed, so watchLoop exits.
	ch := make(chan fsnotify.Event, 1)
	go func() {
		for {
			select {
			case <-f.done:
				close(ch)
				return
			case ev, ok := <-f.events:
				if !ok {
					return
				}
				ch <- ev
			}
		}
	}()
	return ch
}

func (f *fakeWatcher) Errors() <-chan error { return f.errors }

// mockExporter tracks Start and Shutdown calls.
type mockExporter struct {
	startCount    int
	shutdownCount int
	consumeCount  int
}

func (m *mockExporter) Start(context.Context, component.Host) error {
	m.startCount++
	return nil
}

func (m *mockExporter) Shutdown(context.Context) error {
	m.shutdownCount++
	return nil
}

func (m *mockExporter) Capabilities() consumer.Capabilities { return consumer.Capabilities{} }

func (m *mockExporter) ConsumeLogs(_ context.Context, _ plog.Logs) error {
	m.consumeCount++
	return nil
}

// mockSubExporterFactory returns a new mockExporter for every CreateLogs call
// and records all created instances.
type mockSubExporterFactory struct {
	scheme  string
	created []*mockExporter
}

func (f *mockSubExporterFactory) Scheme() string { return f.scheme }

func (f *mockSubExporterFactory) CreateLogs(_ context.Context, _ exporter.Settings, _ ExporterRequest) (exporter.Logs, error) {
	m := &mockExporter{}
	f.created = append(f.created, m)
	return m, nil
}

func (f *mockSubExporterFactory) last() *mockExporter {
	return f.created[len(f.created)-1]
}

func newTestSettings() exporter.Settings {
	return exporter.Settings{
		ID:                component.MustNewID("splunk_outputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
}

func makeSystemOutputsConf(t *testing.T, content string) string {
	t.Helper()
	splunkHome := t.TempDir()
	dir := filepath.Join(splunkHome, "etc", "system", "default")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "outputs.conf"), []byte(content), 0o600))
	return splunkHome
}

// newTestExporter builds a splunkOutputsExporter pre-wired with a fake watcher
// so tests can exercise reconcile without a real OS watcher.
func newTestExporter(t *testing.T, splunkHome string, factory SubExporterFactory) (*splunkOutputsExporter, *fakeWatcher) {
	t.Helper()
	opts := newFactoryOptions(WithSubExporter(factory))
	e := newSplunkOutputsExporter(context.Background(), splunkHome, opts, newTestSettings())
	fake := newFakeWatcher()
	e.watcher = fake
	return e, fake
}

func TestShutdownWithoutStart(t *testing.T) {
	splunkHome := t.TempDir()
	opts := newFactoryOptions()
	e := newSplunkOutputsExporter(context.Background(), splunkHome, opts, newTestSettings())

	// watcher is nil — Shutdown must return without blocking.
	require.NoError(t, e.Shutdown(context.Background()))
}

func TestStartBuildsInitialExporter(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, _ := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	assert.Equal(t, 1, factory.created[0].startCount)
}

func TestStartNoOutputsConf(t *testing.T) {
	splunkHome := t.TempDir() // no outputs.conf anywhere
	opts := newFactoryOptions()
	e := newSplunkOutputsExporter(context.Background(), splunkHome, opts, newTestSettings())
	e.watcher = newFakeWatcher()

	// Should succeed and use nopInstance.
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	e.mu.RLock()
	active := e.active
	e.mu.RUnlock()
	assert.Equal(t, nopInstance, active)
}

func TestConsumeLogsDelegatesToActive(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, _ := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	require.NoError(t, e.ConsumeLogs(context.Background(), plog.NewLogs()))
	assert.Equal(t, 1, factory.created[0].consumeCount)
}

func TestReconcileSwapsExporter(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, _ := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	first := factory.created[0]
	e.reconcile(context.Background())

	// Old exporter should have been shut down.
	assert.Equal(t, 1, first.shutdownCount)
	// A new exporter instance should now be active.
	assert.Len(t, factory.created, 2)
	assert.NotSame(t, first, factory.last())
}

func TestWatchLoopTriggersReconcileAfterDebounce(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, fake := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	first := factory.created[0]

	// Send a filesystem event into the fake watcher.
	fake.events <- fsnotify.Event{Name: filepath.Join(splunkHome, "etc", "system", "default", "outputs.conf")}

	// Wait for debounce + reconcile. debounceDuration is 500ms; give it 2s total.
	assert.Eventually(t, func() bool {
		return first.shutdownCount >= 1
	}, 2*time.Second, 50*time.Millisecond, "expected exporter to reconcile after debounce")
}

func TestShutdownDrainsWatchLoop(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, _ := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = e.Shutdown(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown did not return in time")
	}
}

func TestWatchedDirsRegistered(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, fake := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	assert.Contains(t, fake.added, filepath.Join(splunkHome, "etc", "system"))
	assert.Contains(t, fake.added, filepath.Join(splunkHome, "etc", "system", "default"))
	assert.Contains(t, fake.added, filepath.Join(splunkHome, "etc", "system", "local"))
}

func TestReconcileRegistersLateLocalDir(t *testing.T) {
	factory := &mockSubExporterFactory{scheme: "httpout"}
	splunkHome := makeSystemOutputsConf(t, "[httpout]\nuri = https://hec.example.com\nhttpEventCollectorToken = tok\n")

	e, fake := newTestExporter(t, splunkHome, factory)
	require.NoError(t, e.Start(context.Background(), nil))
	defer e.Shutdown(context.Background()) //nolint:errcheck

	// local/ did not exist at Start time — simulate it being created later.
	localDir := filepath.Join(splunkHome, "etc", "system", "local")
	require.NoError(t, os.MkdirAll(localDir, 0o755))

	e.reconcile(context.Background())

	assert.Contains(t, fake.added, localDir)
}
