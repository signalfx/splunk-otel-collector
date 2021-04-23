// Copyright 2020 Splunk, Inc.
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
	"fmt"

	"github.com/go-zookeeper/zk"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// zkSession implements the configsource.Session interface.
type zkSession struct {
	logger  *zap.Logger
	connect connectFunc
	closeCh chan struct{}
}

var _ configsource.Session = (*zkSession)(nil)

func newSession(logger *zap.Logger, connect connectFunc) *zkSession {
	return &zkSession{
		logger:  logger,
		connect: connect,
		closeCh: make(chan struct{}),
	}
}

func (s *zkSession) Retrieve(ctx context.Context, selector string, _ interface{}) (configsource.Retrieved, error) {
	conn, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	value, _, watchCh, err := conn.GetW(selector)
	if err != nil {
		return nil, err
	}

	return configprovider.NewRetrieved(value, newWatcher(ctx, watchCh, s.closeCh)), nil
}

func (s *zkSession) RetrieveEnd(context.Context) error {
	return nil
}

func (s *zkSession) Close(context.Context) error {
	close(s.closeCh)
	return nil
}

func newWatcher(ctx context.Context, watchCh <-chan zk.Event, closeCh <-chan struct{}) func() error {
	return func() error {
		select {
		case <-closeCh:
			return configsource.ErrSessionClosed
		case e := <-watchCh:
			if e.Err != nil {
				return e.Err
			}

			switch e.Type {
			case zk.EventNodeCreated, zk.EventNodeDataChanged, zk.EventNodeChildrenChanged:
				// EventNodeCreated should never happen but we cover it for completeness.
				return configsource.ErrValueUpdated
			}
			return fmt.Errorf("zookeeper watcher stopped")
		}
	}
}
