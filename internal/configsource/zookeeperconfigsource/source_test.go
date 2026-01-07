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
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func sPtr(s string) *string {
	return &s
}

func TestSessionRetrieve(t *testing.T) {
	conn := newMockConnection(map[string]string{
		"k1":       "v1",
		"d1/d2/k1": "v5",
	})
	source := newZkConfigSource(newMockConnectFunc(conn))

	testsCases := []struct {
		expect *string
		params any
		name   string
		key    string
	}{
		{name: "present", key: "k1", expect: sPtr("v1"), params: nil},
		{name: "present/params", key: "d1/d2/k1", expect: sPtr("v5"), params: "param string"},
		{name: "absent", key: "k2", expect: nil, params: nil},
	}

	for _, c := range testsCases {
		t.Run(c.name, func(t *testing.T) {
			retrieved, err := source.Retrieve(context.Background(), c.key, nil, nil)
			if c.expect != nil {
				require.NoError(t, err)
				require.NoError(t, retrieved.Close(context.Background()))
				return
			}
			require.Error(t, err)
			assert.Nil(t, retrieved)
		})
	}
}

func TestWatcher(t *testing.T) {
	testsCases := []struct {
		name   string
		result string
		close  bool
		err    bool
	}{
		{name: "updated", close: false, result: "v", err: false},
		{name: "source-closed", close: true, result: "", err: false},
		{name: "client-error", close: false, result: "", err: true},
	}

	for _, c := range testsCases {
		t.Run(c.name, func(t *testing.T) {
			conn := newMockConnection(map[string]string{"k1": "v1"})
			connect := newMockConnectFunc(conn)
			source := newZkConfigSource(connect)

			assert.Nil(t, conn.watcherCh)
			watchChannel := make(chan *confmap.ChangeEvent, 1)
			retrieved, err := source.Retrieve(context.Background(), "k1", nil, func(ce *confmap.ChangeEvent) {
				watchChannel <- ce
			})
			require.NoError(t, err)

			val, err := retrieved.AsRaw()
			require.NoError(t, err)
			assert.NotNil(t, val)

			require.NotNil(t, conn.watcherCh)
			switch {
			case c.close:
				require.NoError(t, retrieved.Close(context.Background()))
				assert.Nil(t, conn.watcherCh)
			case c.result != "":
				conn.watcherCh <- zk.Event{
					Type: zk.EventNodeDataChanged,
				}
				ce := <-watchChannel
				assert.NoError(t, ce.Error)
				require.NoError(t, retrieved.Close(context.Background()))
				assert.Nil(t, conn.watcherCh)
			case c.err:
				conn.watcherCh <- zk.Event{
					Err: errors.New("zookeeper error"),
				}
				ce := <-watchChannel
				require.EqualError(t, ce.Error, "zookeeper error")
				require.NoError(t, retrieved.Close(context.Background()))
				assert.Nil(t, conn.watcherCh)
			}
		})
	}
}
