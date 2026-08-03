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
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Nil(t *testing.T, object interface{}) {
	if !isNil(object) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\t   <nil> (expected)\n\n\t!= %#v (actual)\033[39m\n\n",
			filepath.Base(file), line, object)
		t.FailNow()
	}
}

func NotNil(t *testing.T, object interface{}) {
	if isNil(object) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\tExpected value not to be <nil>\033[39m\n\n",
			filepath.Base(file), line)
		t.FailNow()
	}
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}

func TestDiskQueue(t *testing.T) {
	dqName := "test_disk_queue" + strconv.Itoa(int(time.Now().Unix()))
	tmpDir := t.TempDir()
	dq := New(dqName, tmpDir, 1024, 4, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	msg := []byte("test")
	err := dq.Put(msg)
	Nil(t, err)
	require.Equal(t, int64(1), dq.Depth())

	msgOut := <-dq.ReadChan()
	require.Equal(t, msg, msgOut)
}

func TestDiskQueueRoll(t *testing.T) {
	dqName := "test_disk_queue_roll" + strconv.Itoa(int(time.Now().Unix()))
	msg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	ml := int64(len(msg))
	dq := New(dqName, t.TempDir(), 10*(ml+8), ml, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	for i := 0; i < 11; i++ {
		err := dq.Put(msg)
		Nil(t, err)
		require.Equal(t, int64(i+1), dq.Depth())
	}

	require.Equal(t, int64(1), dq.(*diskQueue).writeFileNum)
	require.Equal(t, ml+8, dq.(*diskQueue).writePos)

	for i := 11; i > 0; i-- {
		require.Equal(t, msg, <-dq.ReadChan())
		require.Equal(t, int64(i-1), dq.Depth())
	}
}

func TestDiskQueuePeek(t *testing.T) {
	dqName := "test_disk_queue_peek" + strconv.Itoa(int(time.Now().Unix()))
	msg := bytes.Repeat([]byte{0}, 10)
	ml := int64(len(msg))
	dq := New(dqName, t.TempDir(), 10*(ml+8), ml, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	t.Run("roll", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := dq.Put(msg)
			Nil(t, err)
			require.Equal(t, int64(i+1), dq.Depth())
		}

		for i := 10; i > 0; i-- {
			require.Equal(t, msg, <-dq.PeekChan())
			require.Equal(t, int64(i), dq.Depth())

			require.Equal(t, msg, <-dq.ReadChan())
			require.Equal(t, int64(i-1), dq.Depth())
		}

		Nil(t, dq.Empty())
	})

	t.Run("peek-read", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := dq.Put(msg)
			Nil(t, err)
			require.Equal(t, int64(i+1), dq.Depth())
		}

		for i := 10; i > 0; i-- {
			require.Equal(t, msg, <-dq.PeekChan())
			require.Equal(t, int64(i), dq.Depth())

			require.Equal(t, msg, <-dq.PeekChan())
			require.Equal(t, int64(i), dq.Depth())

			require.Equal(t, msg, <-dq.ReadChan())
			require.Equal(t, int64(i-1), dq.Depth())
		}

		Nil(t, dq.Empty())
	})

	t.Run("read-peek", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := dq.Put(msg)
			Nil(t, err)
			require.Equal(t, int64(i+1), dq.Depth())
		}

		for i := 10; i > 1; i-- {
			require.Equal(t, msg, <-dq.PeekChan())
			require.Equal(t, int64(i), dq.Depth())

			require.Equal(t, msg, <-dq.ReadChan())
			require.Equal(t, int64(i-1), dq.Depth())

			require.Equal(t, msg, <-dq.PeekChan())
			require.Equal(t, int64(i-1), dq.Depth())
		}

		Nil(t, dq.Empty())
	})
}

func assertFileNotExist(t *testing.T, fn string) {
	f, err := os.OpenFile(fn, os.O_RDONLY, 0o600)
	require.Equal(t, (*os.File)(nil), f)
	require.True(t, os.IsNotExist(err))
}

func TestDiskQueueEmpty(t *testing.T) {
	dqName := "test_disk_queue_empty" + strconv.Itoa(int(time.Now().Unix()))
	msg := bytes.Repeat([]byte{0}, 10)
	dq := New(dqName, t.TempDir(), 100, 0, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	for i := 0; i < 100; i++ {
		err := dq.Put(msg)
		Nil(t, err)
		require.Equal(t, int64(i+1), dq.Depth())
	}

	for i := 0; i < 3; i++ {
		<-dq.ReadChan()
	}

	for dq.Depth() != 97 {
		time.Sleep(50 * time.Millisecond)
	}
	require.Equal(t, int64(97), dq.Depth())

	numFiles := dq.(*diskQueue).writeFileNum
	dq.Empty()

	assertFileNotExist(t, dq.(*diskQueue).metaDataFileName())
	for i := int64(0); i <= numFiles; i++ {
		assertFileNotExist(t, dq.(*diskQueue).fileName(i))
	}
	require.Equal(t, int64(0), dq.Depth())
	require.Equal(t, dq.(*diskQueue).writeFileNum, dq.(*diskQueue).readFileNum)
	require.Equal(t, dq.(*diskQueue).writePos, dq.(*diskQueue).readPos)
	require.Equal(t, dq.(*diskQueue).readPos, dq.(*diskQueue).nextReadPos)
	require.Equal(t, dq.(*diskQueue).readFileNum, dq.(*diskQueue).nextReadFileNum)

	for i := 0; i < 100; i++ {
		err := dq.Put(msg)
		Nil(t, err)
		require.Equal(t, int64(i+1), dq.Depth())
	}

	for i := 0; i < 100; i++ {
		<-dq.ReadChan()
	}

	for dq.Depth() != 0 {
		time.Sleep(50 * time.Millisecond)
	}

	require.Equal(t, int64(0), dq.Depth())
	require.Equal(t, dq.(*diskQueue).writeFileNum, dq.(*diskQueue).readFileNum)
	require.Equal(t, dq.(*diskQueue).writePos, dq.(*diskQueue).readPos)
	require.Equal(t, dq.(*diskQueue).readPos, dq.(*diskQueue).nextReadPos)
}

func TestDiskQueueCorruption(t *testing.T) {
	dqName := "test_disk_queue_corruption" + strconv.Itoa(int(time.Now().Unix()))
	// require a non-zero message length for the corrupt (len 0) test below
	dq := New(dqName, t.TempDir(), 1040, 10, 1<<10, 5, 2*time.Second, zap.NewNop())
	defer dq.Close()

	msg := make([]byte, 120) // 128 bytes per message, 8 messages (1024 bytes) per file
	msg[0] = 91
	msg[62] = 4
	msg[119] = 211

	for i := 0; i < 25; i++ {
		dq.Put(msg)
	}

	require.Equal(t, int64(25), dq.Depth())

	// corrupt the 2nd file
	dqFn := dq.(*diskQueue).fileName(1)
	os.Truncate(dqFn, 400) // 3 valid messages, 5 corrupted

	// corrupt the 4th (current) file
	dqFn = dq.(*diskQueue).fileName(3)
	os.Truncate(dqFn, 100)

	for i := 0; i < 18; i++ {
		require.Equal(t, msg, <-dq.ReadChan())
	}
	require.Equal(t, int64(7), dq.Depth())

	require.Equal(t, msg, <-dq.ReadChan())
	require.Equal(t, int64(0), dq.Depth())

	dq.Put(msg) // in 5th file

	require.Equal(t, msg, <-dq.ReadChan())
	require.Equal(t, int64(0), dq.Depth())

	// write a corrupt (len 0) message at the 5th (current) file
	dq.(*diskQueue).writeFile.Write([]byte{0, 0, 0, 0})

	// force a new 6th file - put into 5th, then readOne errors, then put into 6th
	dq.Put(msg)
	dq.Put(msg)

	require.Equal(t, msg, <-dq.ReadChan())

	dq.Put(msg)
	dq.Put(msg)

	// corrupt the last file
	dqFn = dq.(*diskQueue).fileName(5)
	os.Truncate(dqFn, 100)

	require.Equal(t, int64(2), dq.Depth())

	// return one message and try reading again from corrupted file
	require.Equal(t, msg, <-dq.ReadChan())

	// give diskqueue time to handle read error
	time.Sleep(50 * time.Millisecond)

	// the last log file is now considered corrupted leaving no more log messages
	require.Equal(t, int64(0), dq.Depth())
}

type md struct {
	depth        int64
	readFileNum  int64
	writeFileNum int64
	readPos      int64
	writePos     int64
}

func readMetaDataFile(fileName string, retried int) md {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0o600)
	if err != nil {
		// provide a simple retry that results in up to
		// another 500ms for the file to be written.
		if retried < 9 {
			retried++
			time.Sleep(50 * time.Millisecond)
			return readMetaDataFile(fileName, retried)
		}
		panic(err)
	}
	defer f.Close()

	var ret md
	_, err = fmt.Fscanf(f, "%d\n%d,%d\n%d,%d\n",
		&ret.depth,
		&ret.readFileNum, &ret.readPos,
		&ret.writeFileNum, &ret.writePos)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestDiskQueueSyncAfterRead(t *testing.T) {
	dqName := "test_disk_queue_read_after_sync" + strconv.Itoa(int(time.Now().Unix()))
	dq := New(dqName, t.TempDir(), 1<<11, 0, 1<<10, 2500, 50*time.Millisecond, zap.NewNop())
	defer dq.Close()

	msg := make([]byte, 1000)
	dq.Put(msg)

	for i := 0; i < 10; i++ {
		d := readMetaDataFile(dq.(*diskQueue).metaDataFileName(), 0)
		if d.depth == 1 &&
			d.readFileNum == 0 &&
			d.writeFileNum == 0 &&
			d.readPos == 0 &&
			d.writePos == 1008 {
			// success
			goto next
		}
		time.Sleep(100 * time.Millisecond)
	}
	panic("fail")

next:
	dq.Put(msg)
	<-dq.ReadChan()

	for i := 0; i < 10; i++ {
		d := readMetaDataFile(dq.(*diskQueue).metaDataFileName(), 0)
		if d.depth == 1 &&
			d.readFileNum == 0 &&
			d.writeFileNum == 0 &&
			d.readPos == 1008 &&
			d.writePos == 2016 {
			// success
			goto done
		}
		time.Sleep(100 * time.Millisecond)
	}
	panic("fail")

done:
}

func TestDiskQueueTorture(t *testing.T) {
	var wg sync.WaitGroup

	dqName := "test_disk_queue_torture" + strconv.Itoa(int(time.Now().Unix()))
	dq := New(dqName, t.TempDir(), 262144, 0, 1<<10, 2500, 2*time.Second, zap.NewNop())
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	msg := []byte("aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff")

	numWriters := 4
	numReaders := 4
	readExitChan := make(chan int)
	writeExitChan := make(chan int)

	var depth int64
	for range numWriters {
		wg.Go(func() {
			for {
				time.Sleep(100000 * time.Nanosecond)
				select {
				case <-writeExitChan:
					return
				default:
					err := dq.Put(msg)
					if err == nil {
						atomic.AddInt64(&depth, 1)
					}
				}
			}
		})
	}

	time.Sleep(1 * time.Second)

	dq.Close()

	t.Logf("closing writeExitChan")
	close(writeExitChan)
	wg.Wait()

	t.Logf("restarting diskqueue")

	dq = New(dqName, t.TempDir(), 262144, 0, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, depth, dq.Depth())

	var read int64
	for range numReaders {
		wg.Go(func() {
			for {
				time.Sleep(100000 * time.Nanosecond)
				select {
				case m := <-dq.ReadChan():
					if bytes.Equal(m, msg) {
						atomic.AddInt64(&read, 1)
					}
				case <-readExitChan:
					return
				}
			}
		})
	}

	t.Logf("waiting for depth 0")
	for dq.Depth() != 0 {
		time.Sleep(50 * time.Millisecond)
	}

	t.Logf("closing readExitChan")
	close(readExitChan)
	wg.Wait()
	require.Equal(t, depth, read)
}

func TestDiskQueueResize(t *testing.T) {
	dqName := "test_disk_queue_resize" + strconv.Itoa(int(time.Now().Unix()))
	msg := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	ml := int64(len(msg))
	dq := New(dqName, t.TempDir(), 8*(ml+8), ml, 1<<10, 2500, time.Second, zap.NewNop())
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	for i := 0; i < 9; i++ {
		msg[0] = byte(i)
		err := dq.Put(msg)
		Nil(t, err)
	}
	require.Equal(t, int64(1), dq.(*diskQueue).writeFileNum)
	require.Equal(t, ml+8, dq.(*diskQueue).writePos)
	require.Equal(t, int64(9), dq.Depth())

	dq.Close()
	dq = New(dqName, t.TempDir(), 10*(ml+8), ml, 1<<10, 2500, time.Second, zap.NewNop())

	for i := 0; i < 10; i++ {
		msg[0] = byte(20 + i)
		err := dq.Put(msg)
		Nil(t, err)
	}
	require.Equal(t, int64(2), dq.(*diskQueue).writeFileNum)
	require.Equal(t, ml+8, dq.(*diskQueue).writePos)
	require.Equal(t, int64(19), dq.Depth())

	for i := 0; i < 9; i++ {
		msg[0] = byte(i)
		require.Equal(t, msg, <-dq.ReadChan())
	}
	for i := 0; i < 10; i++ {
		msg[0] = byte(20 + i)
		require.Equal(t, msg, <-dq.ReadChan())
	}
	require.Equal(t, int64(0), dq.Depth())
	dq.Close()

	// make sure there aren't "bad" files due to read logic errors
	files, err := filepath.Glob(filepath.Join(t.TempDir(), dqName+"*.bad"))
	Nil(t, err)
	// empty files slice is actually nil, length check is less confusing
	if len(files) > 0 {
		require.Equal(t, []string{}, files)
	}
}

func BenchmarkDiskQueuePut16(b *testing.B) {
	benchmarkDiskQueuePut(b, 16)
}

func BenchmarkDiskQueuePut64(b *testing.B) {
	benchmarkDiskQueuePut(b, 64)
}

func BenchmarkDiskQueuePut256(b *testing.B) {
	benchmarkDiskQueuePut(b, 256)
}

func BenchmarkDiskQueuePut1024(b *testing.B) {
	benchmarkDiskQueuePut(b, 1024)
}

func BenchmarkDiskQueuePut4096(b *testing.B) {
	benchmarkDiskQueuePut(b, 4096)
}

func BenchmarkDiskQueuePut16384(b *testing.B) {
	benchmarkDiskQueuePut(b, 16384)
}

func BenchmarkDiskQueuePut65536(b *testing.B) {
	benchmarkDiskQueuePut(b, 65536)
}

func BenchmarkDiskQueuePut262144(b *testing.B) {
	benchmarkDiskQueuePut(b, 262144)
}

func BenchmarkDiskQueuePut1048576(b *testing.B) {
	benchmarkDiskQueuePut(b, 1048576)
}

func benchmarkDiskQueuePut(b *testing.B, size int64) {
	b.StopTimer()
	dqName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	dq := New(dqName, b.TempDir(), 1024768*100, 0, 1<<20, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	b.SetBytes(size)
	data := make([]byte, size)
	b.StartTimer()

	for b.Loop() {
		err := dq.Put(data)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkDiskWrite16(b *testing.B) {
	benchmarkDiskWrite(b, 16)
}

func BenchmarkDiskWrite64(b *testing.B) {
	benchmarkDiskWrite(b, 64)
}

func BenchmarkDiskWrite256(b *testing.B) {
	benchmarkDiskWrite(b, 256)
}

func BenchmarkDiskWrite1024(b *testing.B) {
	benchmarkDiskWrite(b, 1024)
}

func BenchmarkDiskWrite4096(b *testing.B) {
	benchmarkDiskWrite(b, 4096)
}

func BenchmarkDiskWrite16384(b *testing.B) {
	benchmarkDiskWrite(b, 16384)
}

func BenchmarkDiskWrite65536(b *testing.B) {
	benchmarkDiskWrite(b, 65536)
}

func BenchmarkDiskWrite262144(b *testing.B) {
	benchmarkDiskWrite(b, 262144)
}

func BenchmarkDiskWrite1048576(b *testing.B) {
	benchmarkDiskWrite(b, 1048576)
}

func benchmarkDiskWrite(b *testing.B, size int64) {
	b.StopTimer()
	fileName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	f, _ := os.OpenFile(path.Join(b.TempDir(), fileName), os.O_RDWR|os.O_CREATE, 0o600)
	b.SetBytes(size)
	data := make([]byte, size)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		f.Write(data)
	}
	f.Sync()
}

func BenchmarkDiskWriteBuffered16(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 16)
}

func BenchmarkDiskWriteBuffered64(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 64)
}

func BenchmarkDiskWriteBuffered256(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 256)
}

func BenchmarkDiskWriteBuffered1024(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 1024)
}

func BenchmarkDiskWriteBuffered4096(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 4096)
}

func BenchmarkDiskWriteBuffered16384(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 16384)
}

func BenchmarkDiskWriteBuffered65536(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 65536)
}

func BenchmarkDiskWriteBuffered262144(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 262144)
}

func BenchmarkDiskWriteBuffered1048576(b *testing.B) {
	benchmarkDiskWriteBuffered(b, 1048576)
}

func benchmarkDiskWriteBuffered(b *testing.B, size int64) {
	b.StopTimer()
	fileName := "bench_disk_queue_put" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	f, _ := os.OpenFile(path.Join(b.TempDir(), fileName), os.O_RDWR|os.O_CREATE, 0o600)
	b.SetBytes(size)
	data := make([]byte, size)
	w := bufio.NewWriterSize(f, 1024*4)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w.Write(data)
		if i%1024 == 0 {
			w.Flush()
		}
	}
	w.Flush()
	f.Sync()
}

// you might want to run this like
// $ go test -bench=DiskQueueGet -benchtime 0.1s
// too avoid doing too many iterations.
func BenchmarkDiskQueueGet16(b *testing.B) {
	benchmarkDiskQueueGet(b, 16)
}

func BenchmarkDiskQueueGet64(b *testing.B) {
	benchmarkDiskQueueGet(b, 64)
}

func BenchmarkDiskQueueGet256(b *testing.B) {
	benchmarkDiskQueueGet(b, 256)
}

func BenchmarkDiskQueueGet1024(b *testing.B) {
	benchmarkDiskQueueGet(b, 1024)
}

func BenchmarkDiskQueueGet4096(b *testing.B) {
	benchmarkDiskQueueGet(b, 4096)
}

func BenchmarkDiskQueueGet16384(b *testing.B) {
	benchmarkDiskQueueGet(b, 16384)
}

func BenchmarkDiskQueueGet65536(b *testing.B) {
	benchmarkDiskQueueGet(b, 65536)
}

func BenchmarkDiskQueueGet262144(b *testing.B) {
	benchmarkDiskQueueGet(b, 262144)
}

func BenchmarkDiskQueueGet1048576(b *testing.B) {
	benchmarkDiskQueueGet(b, 1048576)
}

func benchmarkDiskQueueGet(b *testing.B, size int64) {
	b.StopTimer()
	dqName := "bench_disk_queue_get" + strconv.Itoa(b.N) + strconv.Itoa(int(time.Now().Unix()))
	dq := New(dqName, b.TempDir(), 1024768, 0, 1<<30, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	b.SetBytes(size)
	data := make([]byte, size)
	for i := 0; i < b.N; i++ {
		dq.Put(data)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		<-dq.ReadChan()
	}
}

func TestDiskQueueRollAsync(t *testing.T) {
	dqName := "test_disk_queue_roll" + strconv.Itoa(int(time.Now().Unix()))
	msg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	ml := int64(len(msg))
	dq := New(dqName, t.TempDir(), 10*(ml+8), ml, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	for range 11 {
		err := dq.Put(msg)
		Nil(t, err)
		require.Equal(t, int64(1), dq.Depth())

		require.Equal(t, msg, <-dq.ReadChan())
		require.Equal(t, int64(0), dq.Depth())
	}

	require.Equal(t, int64(1), dq.(*diskQueue).writeFileNum)
	require.Equal(t, ml+8, dq.(*diskQueue).writePos)

	filepath.Walk(t.TempDir(), func(path string, _ fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".bad") {
			t.FailNow()
		}

		return err
	})
}

func TestWriteRollReadEOF(t *testing.T) {
	dqName := "test_disk_queue_roll_readEOF" + strconv.Itoa(int(time.Now().Unix()))
	dq := New(dqName, t.TempDir(), 1024, 4, 1<<10, 2500, 2*time.Second, zap.NewNop())
	defer dq.Close()
	NotNil(t, dq)
	require.Equal(t, int64(0), dq.Depth())

	for i := 0; i < 205; i++ { // 204 messages fit, but message 205 will be too big
		msg := []byte(fmt.Sprintf("%05d", i)) // 5 bytes
		err := dq.Put(msg)
		require.NoError(t, err)

		msgOut := <-dq.ReadChan()
		require.Equal(t, msg, msgOut)
	}

	filepath.Walk(t.TempDir(), func(path string, _ fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".bad") {
			t.FailNow()
		}

		return err
	})
}

// TestLargeMessageBoundary verifies that file rotation works correctly when messages are large relative to maxBytesPerFile,
// ensuring no EOF errors occur at boundaries and all messages can be read back without corruption.
func TestLargeMessageBoundary(t *testing.T) {
	// Use smaller sizes to test the same behavior more efficiently
	// 5KB file limit, 4KB max message (same 10:8 ratio as 50MB:40MB in production)
	maxBytesPerFile := int64(5 * 1024)
	maxMsgSize := int64(4 * 1024)

	dq := New("test_large_msg", t.TempDir(), maxBytesPerFile, 1, maxMsgSize, 1000, 2*time.Second, zap.NewNop())
	defer dq.Close()

	// Create messages that will cause multiple rotations
	largeMsg := make([]byte, 4000) // ~4KB message
	for range 15 {                 // ~60KB total, should rotate cleanly across multiple files
		err := dq.Put(largeMsg)
		Nil(t, err)
	}

	// Read all messages back
	for range 15 {
		msg := <-dq.ReadChan()
		require.Len(t, largeMsg, len(msg))
	}

	// Verify no .bad files were created
	filepath.Walk(t.TempDir(), func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".bad") {
			t.Fatalf("Found corrupted file: %s", path)
		}
		return err
	})
}

// TestReadCurrentWriteFile verifies that when reading the current write file,
// the reader doesn't try to rotate past the write file when reaching maxBytesPerFileRead
func TestReadCurrentWriteFile(t *testing.T) {
	// Small file limit to trigger boundary easily
	maxBytesPerFile := int64(1024)
	dq := New("test_current_file", t.TempDir(), maxBytesPerFile, 4, 1024, 1000, 2*time.Second, zap.NewNop())
	defer dq.Close()

	// Write messages up to the file limit
	msg := []byte("test message")
	for i := 0; i < 60; i++ { // Enough to fill first file and start second
		err := dq.Put(msg)
		Nil(t, err)
	}

	// Read all messages back - this tests reading from current write file
	// without trying to advance past it
	for i := 0; i < 60; i++ {
		readMsg := <-dq.ReadChan()
		require.Equal(t, msg, readMsg)
	}

	// Verify no .bad files were created
	filepath.Walk(t.TempDir(), func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".bad") {
			t.Fatalf("Found corrupted file: %s", path)
		}
		return err
	})
}
