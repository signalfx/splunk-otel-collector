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
	"sync/atomic"

	"go.etcd.io/etcd/client/v2"
)

type MockWatcher struct {
	values chan string
	errors chan error
	closed atomic.Bool
}

func newMockWatcher() *MockWatcher {
	return &MockWatcher{
		values: make(chan string, 1),
		errors: make(chan error, 1),
	}
}

func (w *MockWatcher) Next(ctx context.Context) (*client.Response, error) {
	select {
	case <-ctx.Done():
		w.closed.Store(true)
		return nil, context.Canceled
	case err := <-w.errors:
		return nil, err
	case val := <-w.values:
		return &client.Response{
			Node: &client.Node{
				Value: val,
			},
		}, nil
	}
}

type MockKeysAPI struct {
	db            map[string]string
	activeWatcher *MockWatcher
}

func (k *MockKeysAPI) Get(_ context.Context, key string, _ *client.GetOptions) (*client.Response, error) {
	if v, ok := k.db[key]; ok {
		return &client.Response{
			Node: &client.Node{
				Value: v,
			},
		}, nil
	}
	return nil, errors.New("not found")
}

func (k *MockKeysAPI) Watcher(string, *client.WatcherOptions) client.Watcher {
	return k.activeWatcher
}

func (k *MockKeysAPI) Set(context.Context, string, string, *client.SetOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Delete(context.Context, string, *client.DeleteOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Create(context.Context, string, string) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) CreateInOrder(context.Context, string, string, *client.CreateInOrderOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Update(context.Context, string, string) (*client.Response, error) {
	return nil, nil
}
