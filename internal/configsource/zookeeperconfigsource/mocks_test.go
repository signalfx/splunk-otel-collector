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
	"fmt"

	"github.com/go-zookeeper/zk"
)

func newMockConnectFunc(conn zkConnection) connectFunc {
	return func(ctx context.Context) (zkConnection, error) {
		return conn, nil
	}
}

type mockConnection struct {
	db        map[string]string
	watcherCh chan zk.Event
}

func newMockConnection(db map[string]string) *mockConnection {
	return &mockConnection{
		db: db,
	}
}

func (m *mockConnection) GetW(key string) ([]byte, *zk.Stat, <-chan zk.Event, error) {
	if value, ok := m.db[key]; ok {
		m.watcherCh = make(chan zk.Event)
		return []byte(value), &zk.Stat{}, m.watcherCh, nil
	}
	return nil, nil, nil, fmt.Errorf("value not found")
}

func (m *mockConnection) Close() {
	close(m.watcherCh)
	m.watcherCh = nil
}
