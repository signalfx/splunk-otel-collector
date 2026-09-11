// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/splunk/tarunner/pkg/splunkta/tabuilder"
)

const debounceDuration = 500 * time.Millisecond

// fileWatcher abstracts the subset of *fsnotify.Watcher used by the observer.
// It lets tests drive reload logic with a fake instead of a real OS watcher.
type fileWatcher interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// fsnotifyWatcher adapts *fsnotify.Watcher to fileWatcher.
type fsnotifyWatcher struct {
	w *fsnotify.Watcher
}

func newFSNotifyWatcher() (*fsnotifyWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsnotifyWatcher{w: w}, nil
}

func (f *fsnotifyWatcher) Add(name string) error         { return f.w.Add(name) }
func (f *fsnotifyWatcher) Close() error                  { return f.w.Close() }
func (f *fsnotifyWatcher) Events() <-chan fsnotify.Event { return f.w.Events }
func (f *fsnotifyWatcher) Errors() <-chan error          { return f.w.Errors }

// splunkOutputsExporter watches etc/system/default and etc/system/local for
// outputs.conf changes and hot-reloads the underlying exporter on any change.
type splunkOutputsExporter struct {
	splunkHome string
	options    factoryOptions
	settings   exporter.Settings
	host       component.Host
	createCtx  context.Context // context from factory creation, used for initial build

	mu     sync.RWMutex
	active exporter.Logs // currently running exporter; nopInstance when no outputs configured

	watcher fileWatcher
	doneCh  chan struct{}
}

func newSplunkOutputsExporter(ctx context.Context, splunkHome string, options factoryOptions, settings exporter.Settings) *splunkOutputsExporter {
	return &splunkOutputsExporter{
		splunkHome: splunkHome,
		options:    options,
		settings:   settings,
		createCtx:  ctx,
		doneCh:     make(chan struct{}),
	}
}

func (e *splunkOutputsExporter) Start(ctx context.Context, host component.Host) error {
	e.host = host

	initial, err := e.buildExporter(e.createCtx)
	if err != nil {
		return err
	}
	e.active = initial

	// Allow tests to inject a fake watcher before Start is called.
	if e.watcher == nil {
		watcher, err := newFSNotifyWatcher()
		if err != nil {
			return err
		}
		e.watcher = watcher
	}

	for _, dir := range tabuilder.SystemDirs(e.splunkHome) {
		_ = e.watcher.Add(dir) // best-effort; dirs may not exist yet
	}

	go e.watchLoop(ctx)
	return nil
}

func (e *splunkOutputsExporter) Shutdown(ctx context.Context) error {
	if e.watcher != nil {
		_ = e.watcher.Close()
		<-e.doneCh
	}
	e.mu.Lock()
	active := e.active
	e.mu.Unlock()
	if active != nil {
		return active.Shutdown(ctx)
	}
	return nil
}

func (e *splunkOutputsExporter) Capabilities() consumer.Capabilities {
	e.mu.RLock()
	active := e.active
	e.mu.RUnlock()
	if active == nil {
		return consumer.Capabilities{}
	}
	return active.Capabilities()
}

func (e *splunkOutputsExporter) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	e.mu.RLock()
	active := e.active
	e.mu.RUnlock()
	return active.ConsumeLogs(ctx, logs)
}

func (e *splunkOutputsExporter) watchLoop(ctx context.Context) {
	defer close(e.doneCh)

	logger := e.settings.Logger
	var debounce <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-e.watcher.Events():
			if !ok {
				return
			}
			debounce = time.After(debounceDuration)
		case err, ok := <-e.watcher.Errors():
			if !ok {
				return
			}
			logger.Warn("splunk_outputs: watcher error", zap.Error(err))
		case <-debounce:
			debounce = nil
			e.reload(ctx)
		}
	}
}

func (e *splunkOutputsExporter) reload(ctx context.Context) {
	logger := e.settings.Logger

	newExp, err := e.buildExporter(ctx)
	if err != nil {
		logger.Error("splunk_outputs: failed to build exporter on reload", zap.Error(err))
		return
	}

	e.mu.Lock()
	old := e.active
	e.active = newExp
	e.mu.Unlock()

	if old != nil {
		if err := old.Shutdown(ctx); err != nil {
			logger.Warn("splunk_outputs: error shutting down old exporter", zap.Error(err))
		}
	}
	logger.Info("splunk_outputs: reloaded exporter")
}

// buildExporter reads outputs.conf from etc/system, creates sub-exporters for
// each stanza, starts them, and returns the packed result.
func (e *splunkOutputsExporter) buildExporter(ctx context.Context) (exporter.Logs, error) {
	outputs, err := tabuilder.ReadOutputGroups(e.splunkHome)
	if err != nil {
		// No output stanzas configured — return a no-op exporter.
		e.settings.Logger.Info("splunk_outputs: no output stanzas found, using no-op exporter", zap.Error(err))
		return nopInstance, nil
	}

	exporters, err := e.options.createExporters(ctx, e.splunkHome, outputs, e.settings)
	if err != nil {
		return nil, err
	}

	packed := packExporters(exporters)
	if err := packed.Start(ctx, e.host); err != nil {
		return nil, err
	}
	return packed, nil
}
