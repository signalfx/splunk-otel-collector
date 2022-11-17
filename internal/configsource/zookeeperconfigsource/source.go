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

package zookeeperconfigsource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// zkConfigSource implements the configprovider.Session interface.
type zkConfigSource struct {
	logger  *zap.Logger
	connect connectFunc
}

func newConfigSource(params configprovider.CreateParams, cfg *Config) (configprovider.ConfigSource, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, &errMissingEndpoint{errors.New("cannot connect to zk without any endpoints")}
	}

	return newZkConfigSource(params, newConnectFunc(cfg.Endpoints, cfg.Timeout)), nil
}

func newZkConfigSource(params configprovider.CreateParams, connect connectFunc) *zkConfigSource {
	return &zkConfigSource{
		logger:  params.Logger,
		connect: connect,
	}
}

func (s *zkConfigSource) Retrieve(ctx context.Context, selector string, _ *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	conn, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	value, _, watchCh, err := conn.GetW(selector)
	if err != nil {
		return nil, err
	}

	closeCh := make(chan struct{})
	startWatcher(watchCh, closeCh, watcher)
	return confmap.NewRetrieved(string(value), confmap.WithRetrievedClose(func(ctx context.Context) error {
		close(closeCh)
		conn.Close()
		return nil
	}))
}

func (s *zkConfigSource) Shutdown(context.Context) error {
	return nil
}

func startWatcher(watchCh <-chan zk.Event, closeCh <-chan struct{}, watcher confmap.WatcherFunc) {
	go func() {
		select {
		case <-closeCh:
			return
		case e, ok := <-watchCh:
			if !ok {
				// Channel close without any event, connection must have been closed.
				return
			}
			if e.Err != nil {
				watcher(&confmap.ChangeEvent{Error: e.Err})
			}

			switch e.Type {
			case zk.EventNodeCreated, zk.EventNodeDataChanged, zk.EventNodeChildrenChanged:
				// EventNodeCreated should never happen but we cover it for completeness.
				watcher(&confmap.ChangeEvent{})
				return
			}
			watcher(&confmap.ChangeEvent{Error: fmt.Errorf("zookeeper watcher stopped")})
		}
	}()
}

// newConnectFunc returns a new function that can be used to establish and return a connection
// to a zookeeper cluster. Every function returned by newConnectFunc will return the same
// underlying connection until it is lost.
func newConnectFunc(endpoints []string, timeout time.Duration) connectFunc {
	var conn *zk.Conn
	return func(ctx context.Context) (zkConnection, error) {
		if conn != nil && conn.State() != zk.StateDisconnected {
			return conn, nil
		}

		conn, _, err := zk.Connect(endpoints, timeout, zk.WithLogInfo(false))
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}
