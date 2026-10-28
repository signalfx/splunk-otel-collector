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
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/aristanetworks/goarista/gnmireverse"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type gnmiReceiver struct {
	consumer consumer.Metrics
	cfg      *Config
	cancel   context.CancelFunc
	servers  []*grpc.Server
	settings receiver.Settings
	wg       sync.WaitGroup
}

type grpcListener struct {
	listener net.Listener
	server   *grpc.Server
	name     string
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
	schema, err := loadYangSchema(r.cfg.YangModules)
	if err != nil {
		return fmt.Errorf("failed to load yang_modules: %w", err)
	}

	clients := make([]*gnmiClient, 0, len(r.cfg.Targets))
	for i := range r.cfg.Targets {
		parser := newMetricParser(
			r.cfg.Targets[i].ClientConfig.Endpoint,
			r.cfg.Targets[i].Subscriptions,
			schema,
			r.settings.TelemetrySettings.Logger,
		)
		client := newGNMIClient(
			&r.cfg.Targets[i],
			host,
			r.settings.TelemetrySettings,
			r.consumer,
			parser,
		)

		if err = client.connect(startCtx); err != nil {
			for _, started := range clients {
				if started.conn != nil {
					_ = started.conn.Close()
				}
			}
			return fmt.Errorf("target %q: %w", client.target.ClientConfig.Endpoint, err)
		}
		clients = append(clients, client)
	}

	var listeners []grpcListener
	if r.cfg.GNMIReverse != nil && r.cfg.GNMIReverse.NetAddr.Endpoint != "" {
		endpoint := r.cfg.GNMIReverse.NetAddr.Endpoint
		server, serverErr := r.cfg.GNMIReverse.ToServer(startCtx, host.GetExtensions(), r.settings.TelemetrySettings)
		if serverErr != nil {
			closeGNMIClients(clients)
			return fmt.Errorf("create gNMI server: %w", serverErr)
		}
		listener, err := net.Listen(string(r.cfg.GNMIReverse.NetAddr.Transport), endpoint)
		if err != nil {
			closeGNMIClients(clients)
			return fmt.Errorf("listen for gNMI clients on %q: %w", endpoint, err)
		}
		gnmireverse.RegisterGNMIReverseServer(server, &gnmiServer{
			consumer: r.consumer,
			parser: newMetricParser(
				endpoint, r.cfg.GNMIReverse.Subscriptions, schema, r.settings.TelemetrySettings.Logger,
			),
			logger: r.settings.TelemetrySettings.Logger,
		})
		listeners = append(listeners, grpcListener{name: endpoint, server: server, listener: listener})
	}

	ctx, cancel := context.WithCancel(context.WithoutCancel(startCtx))
	r.cancel = cancel

	for _, client := range clients {
		r.wg.Go(func() {
			client.run(ctx)
		})
	}
	for _, l := range listeners {
		r.servers = append(r.servers, l.server)
		r.wg.Go(func() {
			if serveErr := l.server.Serve(l.listener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
				r.settings.TelemetrySettings.Logger.Error("gNMI server stopped",
					zap.String("endpoint", l.name), zap.Error(serveErr))
			}
		})
	}
	return nil
}

func (r *gnmiReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, srv := range r.servers {
			srv.GracefulStop()
		}
		r.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		for _, srv := range r.servers {
			srv.Stop()
		}
		<-done
		return ctx.Err()
	}
}

func closeGNMIClients(clients []*gnmiClient) {
	for _, client := range clients {
		if client.conn != nil {
			_ = client.conn.Close()
		}
	}
}
