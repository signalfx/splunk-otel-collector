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
	"time"

	"github.com/go-zookeeper/zk"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

type zookeeperconfigsource struct {
	logger    *zap.Logger
	endpoints []string
	timeout   time.Duration
}

var _ configsource.ConfigSource = (*zookeeperconfigsource)(nil)

func (z *zookeeperconfigsource) NewSession(ctx context.Context) (configsource.Session, error) {
	return newSession(z.logger, newConnectFunc(z.endpoints, z.timeout)), nil
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
