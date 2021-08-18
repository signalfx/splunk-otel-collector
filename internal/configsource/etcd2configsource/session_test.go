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

package etcd2configsource

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

func sPtr(s string) *string {
	return &s
}

func TestSessionRetrieve(t *testing.T) {
	logger := zap.NewNop()
	kapi := &MockKeysAPI{
		db: map[string]string{
			"k1":       "v1",
			"d1/d2/k1": "v5",
		},
	}

	session := &etcd2Session{logger: logger, kapi: kapi}
	testsCases := []struct {
		params interface{}
		expect *string
		name   string
		key    string
	}{
		{name: "present", key: "k1", expect: sPtr("v1"), params: nil},
		{name: "present/params", key: "d1/d2/k1", expect: sPtr("v5"), params: "param string"},
		{name: "absent", key: "k2", expect: nil, params: nil},
	}

	for _, c := range testsCases {
		t.Run(c.name, func(t *testing.T) {
			retrieved, err := session.Retrieve(context.Background(), c.key, nil)
			if c.expect != nil {
				assert.NoError(t, err)
				_, okWatcher := retrieved.(configsource.Watchable)
				assert.True(t, okWatcher)
				return
			}
			assert.Error(t, err)
			assert.Nil(t, retrieved)
			assert.NoError(t, session.RetrieveEnd(context.Background()))
			assert.NoError(t, session.Close(context.Background()))
		})
	}
}

func TestWatcher(t *testing.T) {
	logger := zap.NewNop()
	kapi := &MockKeysAPI{db: map[string]string{"k1": "v1"}}

	testsCases := []struct {
		err    error
		name   string
		result string
		close  bool
	}{
		{name: "updated", close: false, result: "v", err: nil},
		{name: "session-closed", close: true, result: "", err: nil},
		{name: "client-error", close: false, result: "", err: errors.New("client error")},
	}

	for _, c := range testsCases {
		t.Run(c.name, func(t *testing.T) {
			watcher := newMockWatcher()
			kapi.activeWatcher = watcher

			session := &etcd2Session{logger: logger, kapi: kapi}
			retrieved, err := session.Retrieve(context.Background(), "k1", nil)
			assert.NoError(t, err)
			assert.NotNil(t, retrieved.Value)
			retrievedWatcher, okWatcher := retrieved.(configsource.Watchable)
			assert.True(t, okWatcher)
			assert.False(t, watcher.closed)

			go func() {
				switch {
				case c.close:
					session.Close(context.Background())
				case c.err != nil:
					watcher.errors <- c.err
				case c.result != "":
					watcher.values <- c.result
				}
			}()

			err = retrievedWatcher.WatchForUpdate()

			switch {
			case c.close:
				assert.ErrorIs(t, err, configsource.ErrSessionClosed)
				assert.True(t, watcher.closed)
			case c.err != nil:
				assert.ErrorIs(t, err, c.err)
			case c.result != "":
				assert.ErrorIs(t, err, configsource.ErrValueUpdated)
			}
		})
	}
}
