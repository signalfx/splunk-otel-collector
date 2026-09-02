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

package gnmireceiver

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

type gnmiReceiver struct {
	consumer consumer.Metrics
	cfg      *Config
	cancel   context.CancelFunc
	settings receiver.Settings
	wg       sync.WaitGroup
}

var _ receiver.Metrics = (*gnmiReceiver)(nil)

func newGNMIReceiver(cfg *Config, settings receiver.Settings, nextConsumer consumer.Metrics) *gnmiReceiver {
	return &gnmiReceiver{
		cfg:      cfg,
		settings: settings,
		consumer: nextConsumer,
	}
}

func (r *gnmiReceiver) Start(startCtx context.Context, host component.Host) error {
	clients := make([]*gnmiClient, 0, len(r.cfg.Targets))
	for i := range r.cfg.Targets {
		parser := newMetricParser(
			r.cfg.Targets[i].ClientConfig.Endpoint,
			r.cfg.Targets[i].Subscriptions,
		)
		client := newGNMIClient(
			&r.cfg.Targets[i],
			host,
			r.settings.TelemetrySettings,
			r.consumer,
			parser,
		)

		if err := client.connect(startCtx); err != nil {
			for _, started := range clients {
				if started.conn != nil {
					_ = started.conn.Close()
				}
			}
			return fmt.Errorf("target %q: %w", client.target.ClientConfig.Endpoint, err)
		}
		clients = append(clients, client)
	}

	ctx, cancel := context.WithCancel(context.WithoutCancel(startCtx))
	r.cancel = cancel

	for _, client := range clients {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			client.run(ctx)
		}()
	}
	return nil
}

func (r *gnmiReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
