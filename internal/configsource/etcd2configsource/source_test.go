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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
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

	source := &etcd2ConfigSource{logger: logger, kapi: kapi}
	testsCases := []struct {
		params any
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
			retrieved, err := source.Retrieve(context.Background(), c.key, nil, nil)
			if c.expect != nil {
				assert.NoError(t, err)
				assert.NoError(t, retrieved.Close(context.Background()))
				return
			}
			assert.Error(t, err)
			assert.Nil(t, retrieved)
		})
	}
	assert.NoError(t, source.Shutdown(context.Background()))
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
		{name: "source-closed", close: true, result: "", err: nil},
		{name: "client-error", close: false, result: "", err: errors.New("client error")},
	}

	for _, c := range testsCases {
		t.Run(c.name, func(t *testing.T) {
			watcher := newMockWatcher()
			kapi.activeWatcher = watcher

			watchChannel := make(chan *confmap.ChangeEvent, 1)
			source := &etcd2ConfigSource{logger: logger, kapi: kapi}
			retrieved, err := source.Retrieve(context.Background(), "k1", nil, func(ce *confmap.ChangeEvent) {
				watchChannel <- ce
			})
			require.NoError(t, err)

			val, err := retrieved.AsRaw()
			assert.NoError(t, err)
			assert.NotNil(t, val)

			assert.False(t, watcher.closed.Load())
			switch {
			case c.close:
				assert.NoError(t, retrieved.Close(context.Background()))
				assert.Eventually(t, func() bool {
					return watcher.closed.Load()
				}, 2*time.Second, 10*time.Millisecond)
			case c.err != nil:
				watcher.errors <- c.err
				ce := <-watchChannel
				assert.ErrorIs(t, ce.Error, c.err)
				assert.NoError(t, retrieved.Close(context.Background()))
			case c.result != "":
				watcher.values <- c.result
				ce := <-watchChannel
				assert.NoError(t, ce.Error)
				assert.NoError(t, retrieved.Close(context.Background()))
			}
			assert.NoError(t, source.Shutdown(context.Background()))
		})
	}
}
