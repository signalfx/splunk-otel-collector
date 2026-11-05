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
	"net"
	"slices"
	"sync"

	tpb "github.com/openconfig/grpctunnel/proto/tunnel"
	"github.com/openconfig/grpctunnel/tunnel"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const grpcPassthroughScheme = "passthrough:///"

type tunnelServer struct {
	consumer consumer.Metrics
	ctx      context.Context
	host     component.Host
	listener net.Listener
	ts       *tunnel.Server
	logger   *zap.Logger
	cfg      *TunnelConfig
	server   *grpc.Server
	schema   *yangSchema
	wg       *sync.WaitGroup
	cancels  map[tunnel.Target]context.CancelFunc
	settings component.TelemetrySettings
	mu       sync.Mutex
}

func newTunnelServer(
	startCtx context.Context,
	cfg *TunnelConfig,
	schema *yangSchema,
	host component.Host,
	settings component.TelemetrySettings,
	nextConsumer consumer.Metrics,
) (*tunnelServer, error) {
	s := &tunnelServer{
		cfg:      cfg,
		schema:   schema,
		host:     host,
		settings: settings,
		consumer: nextConsumer,
		logger:   settings.Logger,
		cancels:  make(map[tunnel.Target]context.CancelFunc),
	}

	ts, err := tunnel.NewServer(tunnel.ServerConfig{
		AddTargetHandler:    s.addTarget,
		DeleteTargetHandler: s.deleteTarget,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating tunnel server: %w", err)
	}
	s.ts = ts

	server, err := cfg.ToServer(startCtx, host.GetExtensions(), settings)
	if err != nil {
		return nil, fmt.Errorf("create gRPC server for tunnel: %w", err)
	}
	tpb.RegisterTunnelServer(server, s.ts)
	s.server = server

	listener, err := net.Listen(string(cfg.NetAddr.Transport), cfg.NetAddr.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("listen for tunnel clients on %q: %w", cfg.NetAddr.Endpoint, err)
	}
	s.listener = listener

	return s, nil
}

func (s *tunnelServer) addTarget(t tunnel.Target) error {
	if !slices.Contains(s.cfg.TargetTypes, t.Type) {
		s.logger.Info("ignoring gNMI tunnel target with unaccepted type",
			zap.String("target", t.ID), zap.String("target_type", t.Type))
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.cancels[t]; ok {
		cancel()
		delete(s.cancels, t)
	}

	if s.ctx.Err() != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.cancels[t] = cancel

	client := s.newClient(t)
	s.wg.Go(func() {
		defer cancel()
		client.connectAndRun(ctx)
	})

	s.logger.Info("gNMI tunnel target added",
		zap.String("target", t.ID), zap.String("target_type", t.Type),
		zap.Int("subscriptions", len(s.cfg.Subscriptions)))

	return nil
}

func (s *tunnelServer) deleteTarget(t tunnel.Target) error {
	s.mu.Lock()
	cancel, ok := s.cancels[t]
	delete(s.cancels, t)
	s.mu.Unlock()

	if ok {
		cancel()
		s.logger.Info("gNMI tunnel target deleted",
			zap.String("target", t.ID), zap.String("target_type", t.Type))
	}

	return nil
}

func (s *tunnelServer) newClient(t tunnel.Target) *gnmiClient {
	target := s.cfg.targetConfig(t.Type)
	target.ClientConfig.Endpoint = grpcPassthroughScheme + t.ID
	parser := newMetricParser(t.ID, target.Subscriptions, s.schema, s.logger)

	dialTunnelOption := configgrpc.WithGrpcDialOption(grpc.WithContextDialer(
		func(dialCtx context.Context, _ string) (net.Conn, error) {
			return tunnel.ServerConn(dialCtx, s.ts, &tunnel.Target{ID: t.ID, Type: t.Type})
		}))

	return newGNMIClient(t.ID, &target, s.host, s.settings, s.consumer, parser, dialTunnelOption)
}

func (s *tunnelServer) drainErrors() {
	errCh := s.ts.ErrorChan()
	for {
		select {
		case <-s.ctx.Done():
			return
		case err := <-errCh:
			s.logger.Debug("gNMI tunnel target error", zap.Error(err))
		}
	}
}

func (s *tunnelServer) start(ctx context.Context, wg *sync.WaitGroup) {
	s.ctx = ctx
	s.wg = wg
	wg.Go(s.drainErrors)
}
