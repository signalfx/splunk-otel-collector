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

	"github.com/go-zookeeper/zk"
)

// zkConnection defines an interface that satisfies all functionality
// a session needs from zk.Conn. This allows us to easily mock
// the connection in tests.
type zkConnection interface {
	GetW(string) ([]byte, *zk.Stat, <-chan zk.Event, error)
}

type connectFunc func(context.Context) (zkConnection, error)
