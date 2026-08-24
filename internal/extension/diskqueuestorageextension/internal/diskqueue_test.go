// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newQueue(t *testing.T) Queue {
	return New("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
}

func TestEmptyQueue(t *testing.T) {
	empty := newQueue(t)
	assert.Equal(t, int64(0), empty.Depth())
	select {
	case <-empty.PeekChan():
		assert.Fail(t, "should not peek")
	default:
	}
	require.NoError(t, empty.Close())
}

func TestPutPeekConsume(t *testing.T) {
	q := newQueue(t)
	require.NoError(t, q.Put([]byte("hello world")))
	assert.Equal(t, int64(1), q.Depth())
	msg := <-q.PeekChan()
	assert.Equal(t, int64(1), q.Depth())
	msg.ConsumeCallback()
	assert.Equal(t, int64(0), q.Depth())
	require.NoError(t, q.Close())
}

func TestTwoPutsPeekConsume(t *testing.T) {
	q := newQueue(t)
	require.NoError(t, q.Put([]byte("hello world")))
	require.NoError(t, q.Put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.Depth())
	msg := <-q.PeekChan()
	require.NoError(t, q.Close())
	require.Equal(t, "hello world", string(msg.Payload()))
}

func TestThreePutsThreeConsumes(t *testing.T) {
	q := newQueue(t)
	require.NoError(t, q.Put([]byte("hello world")))
	assert.Equal(t, int64(1), q.Depth())
	require.NoError(t, q.Put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.Depth())
	require.NoError(t, q.Put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.Depth())
	msg := <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	require.Equal(t, "hello world", string(msg.Payload()))
	// do it again
	msg = <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	require.Equal(t, "hello world2", string(msg.Payload()))
	assert.Equal(t, int64(3), q.Depth())
	msg = <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	assert.Equal(t, "hello world3", string(msg.Payload()))
	msg.ConsumeCallback()
	assert.Equal(t, int64(2), q.Depth())
	require.NoError(t, q.Close())
}

func TestThreePutsThreeConsumesOutOfOrder(t *testing.T) {
	q := newQueue(t)
	require.NoError(t, q.Put([]byte("hello world")))
	assert.Equal(t, int64(1), q.Depth())
	require.NoError(t, q.Put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.Depth())
	require.NoError(t, q.Put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.Depth())
	msg1 := <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	msg1.ConsumeCallback()
	assert.Equal(t, int64(2), q.Depth())
	msg2 := <-q.PeekChan()
	assert.Equal(t, int64(2), q.Depth())
	msg3 := <-q.PeekChan()
	assert.Equal(t, int64(2), q.Depth())
	msg3.ConsumeCallback()
	assert.Equal(t, int64(1), q.Depth())
	msg2.ConsumeCallback()
	assert.Equal(t, int64(0), q.Depth())
	require.NoError(t, q.Close())
}

func TestMultipleWorkers(t *testing.T) {
	q := New("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
	require.NoError(t, q.Put([]byte("hello world")))
	assert.Equal(t, int64(1), q.Depth())
	require.NoError(t, q.Put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.Depth())
	require.NoError(t, q.Put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.Depth())
	msg := <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	msg.ConsumeCallback()
	require.Equal(t, "hello world", string(msg.Payload()))
	assert.Equal(t, int64(2), q.Depth())
	msg = <-q.PeekChan()
	assert.Equal(t, int64(2), q.Depth())
	require.Equal(t, "hello world2", string(msg.Payload()))
	msg.ConsumeCallback()
	assert.Equal(t, int64(1), q.Depth())
	msg = <-q.PeekChan()
	assert.Equal(t, int64(1), q.Depth())
	msg.ConsumeCallback()
	require.Equal(t, "hello world3", string(msg.Payload()))
	assert.Equal(t, int64(0), q.Depth())
	require.NoError(t, q.Close())
}

func TestMultipleWorkersOutOfOrder(t *testing.T) {
	q := New("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
	require.NoError(t, q.Put([]byte("hello world")))
	assert.Equal(t, int64(1), q.Depth())
	require.NoError(t, q.Put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.Depth())
	require.NoError(t, q.Put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.Depth())
	msg1 := <-q.PeekChan()
	assert.Equal(t, int64(3), q.Depth())
	msg1.ConsumeCallback()
	assert.Equal(t, int64(2), q.Depth())
	msg2 := <-q.PeekChan()
	assert.Equal(t, int64(2), q.Depth())
	fmt.Println("shoo")
	msg3 := <-q.PeekChan()
	assert.Equal(t, int64(2), q.Depth())
	msg3.ConsumeCallback()
	assert.Equal(t, int64(1), q.Depth())
	msg2.ConsumeCallback()
	assert.Equal(t, int64(0), q.Depth())
	require.NoError(t, q.Close())

	assert.Equal(t, "hello world", string(msg1.Payload()))
	assert.Equal(t, "hello world2", string(msg2.Payload()))
	assert.Equal(t, "hello world3", string(msg3.Payload()))
}
