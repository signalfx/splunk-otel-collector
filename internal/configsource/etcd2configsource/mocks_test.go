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

	"go.etcd.io/etcd/client/v2"
	"go.uber.org/atomic"
)

type MockWatcher struct {
	closed *atomic.Bool
	values chan string
	errors chan error
}

func newMockWatcher() *MockWatcher {
	return &MockWatcher{
		closed: atomic.NewBool(false),
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

func (k *MockKeysAPI) Watcher(_ string, _ *client.WatcherOptions) client.Watcher {
	return k.activeWatcher
}

func (k *MockKeysAPI) Set(_ context.Context, _, _ string, _ *client.SetOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Delete(_ context.Context, _ string, _ *client.DeleteOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Create(_ context.Context, _, _ string) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) CreateInOrder(_ context.Context, _, _ string, _ *client.CreateInOrderOptions) (*client.Response, error) {
	return nil, nil
}

func (k *MockKeysAPI) Update(_ context.Context, _, _ string) (*client.Response, error) {
	return nil, nil
}
