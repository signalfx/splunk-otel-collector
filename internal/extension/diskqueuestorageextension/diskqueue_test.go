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
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newQueueForTesting(t *testing.T) *diskQueue {
	logger, _ := zap.NewDevelopment()
	d, _ := newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, false, logger)
	return d
}

func TestEmptyQueue(t *testing.T) {
	empty := newQueueForTesting(t)
	select {
	case <-empty.peek():
		assert.Fail(t, "should not peek")
	default:
	}
	require.NoError(t, empty.close())
}

func TestPutPeekConsume(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	msg := <-q.peek()
	msg.consumeCallback()
	require.NoError(t, q.close())
}

func TestCatchUpToHeadAndReadOne(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	msg := <-q.peek()
	msg.consumeCallback()
	// we caught up to tip, now do one more
	require.NoError(t, q.put([]byte("hello world")))
	msg = <-q.peek()
	msg.consumeCallback()
	require.NoError(t, q.close())
}

func TestWaitForOneMore(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	msg := <-q.peek()
	msg.consumeCallback()

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = q.put([]byte("hello world"))
	}()
	msg = <-q.peek()

	msg.consumeCallback()
	require.NoError(t, q.close())
}

func TestTwoPutsPeek(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	msg := <-q.peek()
	require.NoError(t, q.close())
	require.Equal(t, "hello world", string(msg.payload))
}

func TestThreePutsThreeConsumes(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	require.NoError(t, q.put([]byte("hello world3")))
	msg := <-q.peek()
	require.Equal(t, "hello world", string(msg.payload))
	// do it again
	msg = <-q.peek()
	require.Equal(t, "hello world2", string(msg.payload))
	msg = <-q.peek()
	assert.Equal(t, "hello world3", string(msg.payload))
	msg.consumeCallback()
	require.NoError(t, q.close())
}

func TestThreePutsThreeConsumesOutOfOrder(t *testing.T) {
	q := newQueueForTesting(t)
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	require.NoError(t, q.put([]byte("hello world3")))
	msg1 := <-q.peek()
	msg1.consumeCallback()
	msg2 := <-q.peek()
	msg3 := <-q.peek()
	msg3.consumeCallback()
	msg2.consumeCallback()
	require.NoError(t, q.close())
}

func TestMultipleWorkers(t *testing.T) {
	q, _ := newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, false, zap.NewNop())
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	require.NoError(t, q.put([]byte("hello world3")))
	msg := <-q.peek()
	require.Equal(t, "hello world", string(msg.payload))
	msg.consumeCallback()
	msg = <-q.peek()
	require.Equal(t, "hello world2", string(msg.payload))
	msg.consumeCallback()
	msg = <-q.peek()
	require.Equal(t, "hello world3", string(msg.payload))
	msg.consumeCallback()
	require.NoError(t, q.close())
}

func TestMultipleWorkersOutOfOrder(t *testing.T) {
	q, _ := newQueue("foo", t.TempDir(), 10_000_000, 1, 1*time.Second, false, zap.NewNop())
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	require.NoError(t, q.put([]byte("hello world3")))
	msg1 := <-q.peek()
	m1 := slices.Clone(msg1.payload)
	msg1.consumeCallback()
	msg2 := <-q.peek()
	m2 := slices.Clone(msg2.payload)
	msg3 := <-q.peek()
	m3 := slices.Clone(msg3.payload)
	msg3.consumeCallback()
	msg2.consumeCallback()
	require.NoError(t, q.close())

	assert.Equal(t, "hello world", string(m1))
	assert.Equal(t, "hello world2", string(m2))
	assert.Equal(t, "hello world3", string(m3))
}

func TestStartStopRestart(t *testing.T) {
	dir := t.TempDir()
	logger, _ := zap.NewDevelopment()
	q, _ := newQueue("foo", dir, 10_000_000, 1, 1*time.Second, false, logger)
	require.NoError(t, q.put([]byte("hello world")))
	require.NoError(t, q.put([]byte("hello world2")))
	require.NoError(t, q.put([]byte("hello world3")))
	msg1 := <-q.peek()
	msg1.consumeCallback()
	require.NoError(t, q.close())
	q, _ = newQueue("foo", dir, 10_000_000, 1, 1*time.Second, false, logger)
	msg2 := <-q.peek()
	require.Equal(t, "hello world2", string(msg2.payload))
	require.NoError(t, q.close())
}
