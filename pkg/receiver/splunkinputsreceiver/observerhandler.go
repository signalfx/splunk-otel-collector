// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// observerHandler manages sub-receivers keyed by TA directory. It responds to
// TA add/remove/change notifications from the filesystem watcher and keeps
// effective configuration metadata for stanza-level reconciliation.
type observerHandler struct {
	host       component.Host
	next       consumer.Logs
	options    factoryOptions
	active     map[string][]receiver.Logs // taDir -> running receivers
	configs    map[string]map[string]receiverSpec
	names      map[string][]string // taDir -> stanza names, in active order
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
		configs:    map[string]map[string]receiverSpec{},
		names:      map[string][]string{},
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

// OnChange reconciles changed TA directories (re-reads merged conf), retaining
// receiver instances whose effective stanza configuration is unchanged.
func (h *observerHandler) OnChange(ctx context.Context, taDirs []string) error {
	h.Lock()
	defer h.Unlock()
	var errs []error
	for _, taDir := range taDirs {
		if err := h.change(ctx, taDir); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *observerHandler) add(ctx context.Context, taDirs []string) error {
	var errs []error
	for _, taDir := range taDirs {
		if _, ok := h.active[taDir]; ok {
			continue
		}
		specs, err := h.options.receiverSpecs(h.splunkHome, taDir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		started, err := h.options.startReceiverSpecs(ctx, h.host, taDir, h.next, h.settings, specs)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		rcvrs := make([]receiver.Logs, 0, len(started))
		configs := make(map[string]receiverSpec, len(started))
		names := make([]string, 0, len(started))
		for i := range started {
			startedReceiver := &started[i]
			rcvrs = append(rcvrs, startedReceiver.receiver)
			configs[startedReceiver.spec.name] = startedReceiver.spec
			names = append(names, startedReceiver.spec.name)
		}
		// Keep an empty entry for a discovered TA with no currently runnable
		// stanzas. A later edit can enable a stanza and must be reconciled. The
		// system sentinel is different: it is only active when it has work.
		if taDir != systemKey || len(rcvrs) > 0 {
			h.active[taDir] = rcvrs
			h.configs[taDir] = configs
			h.names[taDir] = names
		}
	}
	return errors.Join(errs...)
}

// change reconciles the individual receiver instances for one TA. Receivers
// whose effective configuration is unchanged are retained, while removed or
// changed stanzas are stopped and newly changed stanzas are started.
func (h *observerHandler) change(ctx context.Context, taDir string) error {
	specs, err := h.options.receiverSpecs(h.splunkHome, taDir)
	if err != nil {
		return err
	}

	oldReceivers := receiversByName(h.active[taDir], h.names[taDir])
	oldConfigs := h.configs[taDir]
	newReceivers := make([]receiver.Logs, 0, len(specs))
	newConfigs := make(map[string]receiverSpec, len(specs))
	var errs []error

	for i := range specs {
		spec := &specs[i]
		old, exists := oldReceivers[spec.name]
		oldSpec, hasSpec := oldConfigs[spec.name]
		if exists && hasSpec && oldSpec.equal(*spec) {
			newReceivers = append(newReceivers, old)
			newConfigs[spec.name] = *spec
			delete(oldReceivers, spec.name)
			continue
		}

		if exists {
			h.stopReceiver(ctx, taDir, spec.name, old)
			delete(oldReceivers, spec.name)
		}

		started, startErr := h.options.startReceiverSpecs(ctx, h.host, taDir, h.next, h.settings, []receiverSpec{*spec})
		if startErr != nil {
			errs = append(errs, startErr)
			continue
		}
		if len(started) == 1 {
			newReceivers = append(newReceivers, started[0].receiver)
			newConfigs[spec.name] = started[0].spec
		}
	}

	// Anything left was removed from the effective configuration. This also
	// stops receivers inserted by older versions that have no saved spec.
	for name, old := range oldReceivers {
		h.stopReceiver(ctx, taDir, name, old)
	}

	if len(newReceivers) == 0 {
		if taDir == systemKey {
			delete(h.active, taDir)
			delete(h.configs, taDir)
			delete(h.names, taDir)
			return errors.Join(errs...)
		}
		// Keep the TA known to the reconciler even when all of its stanzas are
		// disabled or unsupported; a future config edit may make one runnable.
		h.active[taDir] = nil
		h.configs[taDir] = newConfigs
		h.names[taDir] = nil
	} else {
		h.active[taDir] = newReceivers
		h.configs[taDir] = newConfigs
		h.names[taDir] = make([]string, 0, len(newReceivers))
		for i := range specs {
			spec := &specs[i]
			if _, ok := newConfigs[spec.name]; ok {
				h.names[taDir] = append(h.names[taDir], spec.name)
			}
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
	for i, r := range h.active[taDir] {
		name := ""
		if i < len(h.names[taDir]) {
			name = h.names[taDir][i]
		}
		h.stopReceiver(ctx, taDir, name, r)
	}
	delete(h.active, taDir)
	delete(h.configs, taDir)
	delete(h.names, taDir)
}

func (h *observerHandler) stopReceiver(ctx context.Context, taDir, name string, r receiver.Logs) {
	if err := r.Shutdown(ctx); err != nil {
		h.settings.Logger.Error("splunk_inputs: failed to stop receiver",
			zap.String("ta", taDir), zap.String("stanza", name), zap.Error(err))
	}
}

func receiversByName(rcvrs []receiver.Logs, names []string) map[string]receiver.Logs {
	result := make(map[string]receiver.Logs, len(rcvrs))
	for i, r := range rcvrs {
		name := fmt.Sprintf("\x00legacy-%d", i)
		if i < len(names) {
			name = names[i]
		}
		result[name] = r
	}
	return result
}
