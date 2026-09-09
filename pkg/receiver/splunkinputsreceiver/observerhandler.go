// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// observerHandler manages sub-receivers keyed by TA directory. It responds to
// TA add/remove/change notifications from the filesystem watcher.
type observerHandler struct {
	host       component.Host
	next       consumer.Logs
	options    factoryOptions
	active     map[string][]receiver.Logs // taDir -> running receivers
	splunkHome string
	settings   receiver.Settings
	sync.Mutex
}

func newObserverHandler(splunkHome string, options factoryOptions, settings receiver.Settings, next consumer.Logs) *observerHandler {
	return &observerHandler{
		splunkHome: splunkHome,
		options:    options,
		settings:   settings,
		next:       next,
		active:     map[string][]receiver.Logs{},
	}
}

// OnAdd starts receivers for each newly discovered TA directory.
func (h *observerHandler) OnAdd(ctx context.Context, taDirs []string) error {
	h.Lock()
	defer h.Unlock()
	return h.add(ctx, taDirs)
}

// OnRemove stops receivers for each removed TA directory.
func (h *observerHandler) OnRemove(ctx context.Context, taDirs []string) {
	h.Lock()
	defer h.Unlock()
	h.remove(ctx, taDirs)
}

// OnChange restarts receivers for changed TA directories (re-reads merged conf).
func (h *observerHandler) OnChange(ctx context.Context, taDirs []string) error {
	h.Lock()
	defer h.Unlock()
	h.remove(ctx, taDirs)
	return h.add(ctx, taDirs)
}

func (h *observerHandler) add(ctx context.Context, taDirs []string) error {
	var errs []error
	for _, taDir := range taDirs {
		if _, ok := h.active[taDir]; ok {
			continue
		}
		rcvrs, err := h.options.startReceivers(ctx, h.host, h.splunkHome, taDir, h.next, h.settings)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if len(rcvrs) > 0 {
			h.active[taDir] = rcvrs
		}
	}
	return errors.Join(errs...)
}

func (h *observerHandler) remove(ctx context.Context, taDirs []string) {
	for _, taDir := range taDirs {
		h.stopReceivers(ctx, taDir)
	}
}

// shutdown stops all active receivers.
func (h *observerHandler) shutdown(ctx context.Context) {
	h.Lock()
	defer h.Unlock()

	for taDir := range h.active {
		h.stopReceivers(ctx, taDir)
	}
}

func (h *observerHandler) stopReceivers(ctx context.Context, taDir string) {
	for _, r := range h.active[taDir] {
		if err := r.Shutdown(ctx); err != nil {
			h.settings.Logger.Error("splunk_inputs: failed to stop receiver",
				zap.String("ta", taDir), zap.Error(err))
		}
	}
	delete(h.active, taDir)
}
