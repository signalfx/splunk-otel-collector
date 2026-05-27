// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package otlpreceiver

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

var receivers = newReceiverMap()

type receiverMap struct {
	lock      sync.Mutex
	receivers map[*Config]*delayedReceiver
}

func newReceiverMap() *receiverMap {
	return &receiverMap{
		receivers: map[*Config]*delayedReceiver{},
	}
}

func (m *receiverMap) LoadOrStore(cfg *Config, set receiver.Settings, comp component.Component) *delayedReceiver {
	m.lock.Lock()
	defer m.lock.Unlock()

	if receiver, ok := m.receivers[cfg]; ok {
		return receiver
	}

	receiver := &delayedReceiver{
		cfg:       cfg,
		component: comp,
		logger:    set.Logger,
		removeFunc: func() {
			m.lock.Lock()
			defer m.lock.Unlock()
			delete(m.receivers, cfg)
		},
	}
	m.receivers[cfg] = receiver
	return receiver
}

type delayedReceiver struct {
	cfg       *Config
	component component.Component
	logger    *zap.Logger

	delayOnce sync.Once
	delayErr  error
	stopOnce  sync.Once

	removeFunc func()
}

func (r *delayedReceiver) Start(ctx context.Context, host component.Host) error {
	if err := r.waitForStartDelay(ctx); err != nil {
		return err
	}
	return r.component.Start(ctx, host)
}

func (r *delayedReceiver) waitForStartDelay(ctx context.Context) error {
	r.delayOnce.Do(func() {
		start := time.Now()
		if r.cfg.StartDelay > 0 {
			timer := time.NewTimer(r.cfg.StartDelay)
			defer timer.Stop()

			select {
			case <-timer.C:
			case <-ctx.Done():
				r.delayErr = ctx.Err()
				return
			}
		}

		elapsed := time.Since(start)
		r.logger.Info(
			"Starting OTLP receiver after configured delay",
			zap.Duration("configured_delay", r.cfg.StartDelay),
			zap.Duration("elapsed_since_collector_start", elapsed),
		)
	})
	return r.delayErr
}

func (r *delayedReceiver) Shutdown(ctx context.Context) error {
	var err error
	r.stopOnce.Do(func() {
		err = r.component.Shutdown(ctx)
		r.removeFunc()
	})
	return err
}
