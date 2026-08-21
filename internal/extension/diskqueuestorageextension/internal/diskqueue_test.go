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
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDiskQueuePutGetDelete(t *testing.T) {
	dq := New("test", t.TempDir(), 1024, 1, time.Second, zap.NewNop())
	defer dq.Close()

	require.Equal(t, int64(0), dq.Depth())

	msg := []byte("hello world")
	require.NoError(t, dq.Put(1, msg))
	require.Equal(t, int64(1), dq.Depth())

	got, err := dq.Get(1)
	require.NoError(t, err)
	require.Equal(t, msg, got)

	// Get is a peek: it must not remove the record.
	got, err = dq.Get(1)
	require.NoError(t, err)
	require.Equal(t, msg, got)
	require.Equal(t, int64(1), dq.Depth())

	require.NoError(t, dq.Delete(1))
	require.Equal(t, int64(0), dq.Depth())

	got, err = dq.Get(1)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestDiskQueueGetMissingKey(t *testing.T) {
	dq := New("test", t.TempDir(), 1024, 1, time.Second, zap.NewNop())
	defer dq.Close()

	got, err := dq.Get(42)
	require.NoError(t, err)
	require.Nil(t, got)

	require.NoError(t, dq.Delete(42))
}

// TestDiskQueueOutOfOrderDelete is the regression test for the head-only
// FIFO bug: deleting keys out of write order must only remove the deleted
// key's data, never a different, still-live key's data. This is the scenario
// a >1 num_consumers exporter queue, or an unordered dispatched-item replay
// on restart, will trigger routinely.
func TestDiskQueueOutOfOrderDelete(t *testing.T) {
	dq := New("test", t.TempDir(), 1024, 1, time.Second, zap.NewNop())
	defer dq.Close()

	msgs := map[uint64][]byte{
		1: []byte("first"),
		2: []byte("second"),
		3: []byte("third"),
	}
	for key, msg := range msgs {
		require.NoError(t, dq.Put(key, msg))
	}
	require.Equal(t, int64(3), dq.Depth())

	// Delete the middle key first.
	require.NoError(t, dq.Delete(2))
	require.Equal(t, int64(2), dq.Depth())

	got, err := dq.Get(1)
	require.NoError(t, err)
	require.Equal(t, msgs[1], got)

	got, err = dq.Get(3)
	require.NoError(t, err)
	require.Equal(t, msgs[3], got)

	got, err = dq.Get(2)
	require.NoError(t, err)
	require.Nil(t, got)

	require.NoError(t, dq.Delete(1))
	require.NoError(t, dq.Delete(3))
	require.Equal(t, int64(0), dq.Depth())
}

func TestDiskQueueRoll(t *testing.T) {
	dir := t.TempDir()
	msg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	dq := New("test", dir, 10*int64(len(msg)), 1, time.Second, zap.NewNop())
	defer dq.Close()

	for i := range uint64(11) {
		require.NoError(t, dq.Put(i, msg))
	}
	require.Equal(t, int64(11), dq.Depth())

	for i := range uint64(11) {
		got, err := dq.Get(i)
		require.NoError(t, err)
		require.Equal(t, msg, got)
	}

	require.Equal(t, int64(1), dq.(*diskQueue).writeFileNum)
}

// TestDiskQueueRollFileCleanup verifies that once every record in a rolled
// segment file has been deleted, the file is removed, but only once it is no
// longer the active write file.
func TestDiskQueueRollFileCleanup(t *testing.T) {
	dir := t.TempDir()
	msg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	dqName := "test"
	dq := New(dqName, dir, 2*int64(len(msg)), 1, time.Second, zap.NewNop())
	defer dq.Close()

	// Two messages per file: keys 0,1 -> file 0; keys 2,3 -> file 1.
	for i := range uint64(4) {
		require.NoError(t, dq.Put(i, msg))
	}

	require.FileExists(t, filepath.Join(dir, dqName+".diskqueue.000000.dat"))
	require.FileExists(t, filepath.Join(dir, dqName+".diskqueue.000001.dat"))

	require.NoError(t, dq.Delete(0))
	require.FileExists(t, filepath.Join(dir, dqName+".diskqueue.000000.dat"))

	require.NoError(t, dq.Delete(1))
	require.NoFileExists(t, filepath.Join(dir, dqName+".diskqueue.000000.dat"))

	// File 1 is still the active write file, so it must survive even
	// once all its records are deleted.
	require.NoError(t, dq.Delete(2))
	require.NoError(t, dq.Delete(3))
	require.FileExists(t, filepath.Join(dir, dqName+".diskqueue.000001.dat"))
}

func TestDiskQueueRestart(t *testing.T) {
	dir := t.TempDir()
	dqName := "test"
	msg := []byte("persisted")

	dq := New(dqName, dir, 1024, 1, time.Second, zap.NewNop())
	require.NoError(t, dq.Put(1, msg))
	require.NoError(t, dq.Put(2, msg))
	require.NoError(t, dq.Delete(1))
	require.NoError(t, dq.Close())

	dq = New(dqName, dir, 1024, 1, time.Second, zap.NewNop())
	defer dq.Close()

	require.Equal(t, int64(1), dq.Depth())

	got, err := dq.Get(2)
	require.NoError(t, err)
	require.Equal(t, msg, got)

	got, err = dq.Get(1)
	require.NoError(t, err)
	require.Nil(t, got)

	// A new key must land after the previously written data, not overwrite it.
	require.NoError(t, dq.Put(3, msg))
	got, err = dq.Get(2)
	require.NoError(t, err)
	require.Equal(t, msg, got)
}

func TestDiskQueueConcurrentPutGetDelete(t *testing.T) {
	dq := New("test", t.TempDir(), 4096, 10, 50*time.Millisecond, zap.NewNop())
	defer dq.Close()

	const n = 200
	var wg sync.WaitGroup
	for i := range uint64(n) {
		wg.Add(1)
		go func(key uint64) {
			defer wg.Done()
			msg := []byte(fmt.Sprintf("msg-%d", key))
			require.NoError(t, dq.Put(key, msg))
			got, err := dq.Get(key)
			require.NoError(t, err)
			require.Equal(t, msg, got)
			require.NoError(t, dq.Delete(key))
		}(i)
	}
	wg.Wait()

	require.Equal(t, int64(0), dq.Depth())
}

func BenchmarkDiskQueuePut(b *testing.B) {
	dq := New("bench", b.TempDir(), 1024*1024, 100, time.Second, zap.NewNop())
	defer dq.Close()
	data := make([]byte, 256)

	for i := 0; b.Loop(); i++ {
		if err := dq.Put(uint64(i), data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiskQueueGet(b *testing.B) {
	dq := New("bench", b.TempDir(), 1024*1024, 100, time.Second, zap.NewNop())
	defer dq.Close()
	data := make([]byte, 256)

	const n = 1000
	for i := range uint64(n) {
		if err := dq.Put(i, data); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; b.Loop(); i++ {
		key := uint64(i % n)
		if _, err := dq.Get(key); err != nil {
			b.Fatal(err)
		}
	}
}
