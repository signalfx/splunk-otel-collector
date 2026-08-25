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

package diskqueuestorageextension

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newQueueForTesting(t *testing.T) *diskQueue {
	return newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
}

func TestEmptyQueue(t *testing.T) {
	empty := newQueueForTesting(t)
	assert.Equal(t, int64(0), empty.depth())
	select {
	case <-empty.peekChan:
		assert.Fail(t, "should not peek")
	default:
	}
	require.NoError(t, empty.close())
}

func TestPutPeekConsume(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	assert.Equal(t, int64(1), q.depth())
	msg := <-q.peekChan
	assert.Equal(t, int64(1), q.depth())
	msg.consumeCallback()
	assert.Equal(t, int64(0), q.depth())
	require.NoError(t, q.close())
}

func TestTwoPutsPeek(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.depth())
	msg := <-q.peekChan
	require.NoError(t, q.close())
	require.Equal(t, "hello world", string(msg.payload()))
}

func TestThreePutsThreeConsumes(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	assert.Equal(t, int64(1), q.depth())
	require.NoError(t, q.put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.depth())
	require.NoError(t, q.put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.depth())
	msg := <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	require.Equal(t, "hello world", string(msg.payload()))
	// do it again
	msg = <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	require.Equal(t, "hello world2", string(msg.payload()))
	assert.Equal(t, int64(3), q.depth())
	msg = <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	assert.Equal(t, "hello world3", string(msg.payload()))
	msg.consumeCallback()
	assert.Equal(t, int64(2), q.depth())
	require.NoError(t, q.close())
}

func TestThreePutsThreeConsumesOutOfOrder(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	assert.Equal(t, int64(1), q.depth())
	require.NoError(t, q.put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.depth())
	require.NoError(t, q.put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.depth())
	msg1 := <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	msg1.consumeCallback()
	assert.Equal(t, int64(2), q.depth())
	msg2 := <-q.peekChan
	assert.Equal(t, int64(2), q.depth())
	msg3 := <-q.peekChan
	assert.Equal(t, int64(2), q.depth())
	msg3.consumeCallback()
	assert.Equal(t, int64(1), q.depth())
	msg2.consumeCallback()
	assert.Equal(t, int64(0), q.depth())
	require.NoError(t, q.close())
}

func TestMultipleWorkers(t *testing.T) {
	q := newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
	require.NoError(t, q.put([]byte("hello world")))
	assert.Equal(t, int64(1), q.depth())
	require.NoError(t, q.put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.depth())
	require.NoError(t, q.put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.depth())
	msg := <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	msg.consumeCallback()
	require.Equal(t, "hello world", string(msg.payload()))
	assert.Equal(t, int64(2), q.depth())
	msg = <-q.peekChan
	assert.Equal(t, int64(2), q.depth())
	require.Equal(t, "hello world2", string(msg.payload()))
	msg.consumeCallback()
	assert.Equal(t, int64(1), q.depth())
	msg = <-q.peekChan
	assert.Equal(t, int64(1), q.depth())
	msg.consumeCallback()
	require.Equal(t, "hello world3", string(msg.payload()))
	assert.Equal(t, int64(0), q.depth())
	require.NoError(t, q.close())
}

func TestMultipleWorkersOutOfOrder(t *testing.T) {
	q := newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, zap.NewNop())
	require.NoError(t, q.put([]byte("hello world")))
	assert.Equal(t, int64(1), q.depth())
	require.NoError(t, q.put([]byte("hello world2")))
	assert.Equal(t, int64(2), q.depth())
	require.NoError(t, q.put([]byte("hello world3")))
	assert.Equal(t, int64(3), q.depth())
	msg1 := <-q.peekChan
	assert.Equal(t, int64(3), q.depth())
	msg1.consumeCallback()
	assert.Equal(t, int64(2), q.depth())
	msg2 := <-q.peekChan
	assert.Equal(t, int64(2), q.depth())
	fmt.Println("shoo")
	msg3 := <-q.peekChan
	assert.Equal(t, int64(2), q.depth())
	msg3.consumeCallback()
	assert.Equal(t, int64(1), q.depth())
	msg2.consumeCallback()
	assert.Equal(t, int64(0), q.depth())
	require.NoError(t, q.close())

	assert.Equal(t, "hello world", string(msg1.payload()))
	assert.Equal(t, "hello world2", string(msg2.payload()))
	assert.Equal(t, "hello world3", string(msg3.payload()))
}
