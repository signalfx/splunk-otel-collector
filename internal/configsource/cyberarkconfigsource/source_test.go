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

package cyberarkconfigsource

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

// fakeRetriever is a test double for the retriever interface. It returns whatever the
// current results function yields and counts invocations.
type fakeRetriever struct {
	mu     sync.Mutex
	calls  int
	result func(call int) (map[string]any, error)
}

func (f *fakeRetriever) retrieve(context.Context) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.result(f.calls)
}

func (f *fakeRetriever) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func newTestSource(r retriever, autoRefresh bool, pollInterval time.Duration) *cyberarkConfigSource {
	return &cyberarkConfigSource{
		logger:       zap.NewNop(),
		retriever:    r,
		autoRefresh:  autoRefresh,
		pollInterval: pollInterval,
	}
}

func TestRetrieve_CacheReuseAcrossSelectors(t *testing.T) {
	fields := map[string]any{"Password": "s3cr3t", "UserName": "svc"}
	r := &fakeRetriever{result: func(int) (map[string]any, error) { return fields, nil }}
	source := newTestSource(r, false, defaultPollInterval)

	// Empty selector -> Password.
	got, err := source.Retrieve(context.Background(), "", nil, nil)
	require.NoError(t, err)
	val, err := got.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", val)

	// Explicit UserName selector.
	got, err = source.Retrieve(context.Background(), "UserName", nil, nil)
	require.NoError(t, err)
	val, err = got.AsRaw()
	require.NoError(t, err)
	assert.Equal(t, "svc", val)

	// The object is fetched only once regardless of the number of selectors.
	assert.Equal(t, 1, r.callCount())
}

func TestRetrieve_BadSelector(t *testing.T) {
	r := &fakeRetriever{result: func(int) (map[string]any, error) {
		return map[string]any{"Password": "s3cr3t"}, nil
	}}
	source := newTestSource(r, false, defaultPollInterval)

	got, err := source.Retrieve(context.Background(), "Nonexistent", nil, nil)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.IsType(t, &errBadSelector{}, err)
}

func TestRetrieve_RetrieverErrorLeavesCacheNil(t *testing.T) {
	wantErr := errors.New("boom")
	r := &fakeRetriever{result: func(int) (map[string]any, error) { return nil, wantErr }}
	source := newTestSource(r, false, defaultPollInterval)

	got, err := source.Retrieve(context.Background(), "", nil, nil)
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
	assert.Nil(t, source.fields)
}

func TestRetrieve_StaticYieldsNilCloseFunc(t *testing.T) {
	r := &fakeRetriever{result: func(int) (map[string]any, error) {
		return map[string]any{"Password": "s3cr3t"}, nil
	}}
	source := newTestSource(r, false, defaultPollInterval)

	got, err := source.Retrieve(context.Background(), "", nil, func(*confmap.ChangeEvent) {
		t.Fatal("watcher must not be triggered for static config source")
	})
	require.NoError(t, err)
	// A nil close func is a valid no-op close.
	require.NoError(t, got.Close(context.Background()))
}

func TestRetrieve_AutoRefreshEmitsChangeEvent(t *testing.T) {
	// First call returns the original values, later calls return changed values.
	r := &fakeRetriever{result: func(call int) (map[string]any, error) {
		if call == 1 {
			return map[string]any{"Password": "old"}, nil
		}
		return map[string]any{"Password": "new"}, nil
	}}
	source := newTestSource(r, true, 10*time.Millisecond)

	watchCh := make(chan *confmap.ChangeEvent, 1)
	got, err := source.Retrieve(context.Background(), "", nil, func(ce *confmap.ChangeEvent) {
		watchCh <- ce
	})
	require.NoError(t, err)

	select {
	case ce := <-watchCh:
		require.NoError(t, ce.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("expected a change event from the polling watcher")
	}

	require.NoError(t, got.Close(context.Background()))
}

func TestRetrieve_AutoRefreshErrorEmitsErrorEvent(t *testing.T) {
	r := &fakeRetriever{result: func(call int) (map[string]any, error) {
		if call == 1 {
			return map[string]any{"Password": "old"}, nil
		}
		return nil, errors.New("refresh failed")
	}}
	source := newTestSource(r, true, 10*time.Millisecond)

	watchCh := make(chan *confmap.ChangeEvent, 1)
	got, err := source.Retrieve(context.Background(), "", nil, func(ce *confmap.ChangeEvent) {
		watchCh <- ce
	})
	require.NoError(t, err)

	select {
	case ce := <-watchCh:
		require.Error(t, ce.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("expected an error change event from the polling watcher")
	}

	require.NoError(t, got.Close(context.Background()))
}

func TestNewRetriever_UnsupportedMode(t *testing.T) {
	_, err := newRetriever(&Config{RetrievalMode: "ccp"})
	require.Error(t, err)
	assert.IsType(t, &errUnsupportedMode{}, err)
}
