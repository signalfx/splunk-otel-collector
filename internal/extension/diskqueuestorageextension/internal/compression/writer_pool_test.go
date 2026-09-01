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

package compression

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGZIPWriterPool(t *testing.T) {
	testWriterPool(t, NewGZIPWriterPool, gunzipData)
}

func testWriterPool(t *testing.T, newPool func() WriterPool, decompress func(testing.TB, []byte) []byte) {
	t.Helper()
	t.Parallel()

	tests := []struct {
		name  string
		data  string
		write bool
	}{
		{name: "empty", write: true},
		{name: "small", data: "hello, world", write: true},
		{name: "copy", data: "copy data to compress"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var dst bytes.Buffer
			pool := newPool()
			var (
				n   int64
				err error
			)
			if tt.write {
				n, err = pool.Write(&dst, []byte(tt.data))
			} else {
				n, err = pool.Copy(&dst, bytes.NewReader([]byte(tt.data)))
			}

			require.NoError(t, err)
			require.Equal(t, int64(len(tt.data)), n)
			require.Equal(t, tt.data, string(decompress(t, dst.Bytes())))
		})
	}
}

func TestWriterPoolErrors(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")

	tests := []struct {
		src         io.Reader
		writer      *testResetWriteCloser
		name        string
		expectedErr []error
		expectedN   int64
	}{
		{
			name:        "read and close failures",
			writer:      &testResetWriteCloser{closeErr: closeErr},
			src:         errorReader{err: readErr},
			expectedErr: []error{readErr, closeErr},
		},
		{
			name:        "write and close failures",
			writer:      &testResetWriteCloser{writeErr: writeErr, closeErr: closeErr},
			src:         bytes.NewReader([]byte("data")),
			expectedErr: []error{writeErr, closeErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool := NewWriterPool(func() *testResetWriteCloser { return tt.writer })
			n, err := pool.Copy(io.Discard, tt.src)

			require.Equal(t, tt.expectedN, n)
			for _, expectedErr := range tt.expectedErr {
				require.ErrorIs(t, err, expectedErr)
			}
			require.Equal(t, 1, tt.writer.closeCalls)
		})
	}
}

func BenchmarkGZIPWriterPool(b *testing.B) {
	benchmarkWriterPool(b, NewGZIPWriterPool)
}

func benchmarkWriterPool(b *testing.B, newPool func() WriterPool) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			data := bytes.Repeat([]byte("a"), size)
			pool := newPool()
			var dst bytes.Buffer

			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				dst.Reset()
				n, err := pool.Write(&dst, data)
				if err != nil {
					b.Fatal(err)
				}
				if n != int64(len(data)) {
					b.Fatalf("wrote %d bytes, want %d", n, len(data))
				}
			}
		})
	}
}

func BenchmarkGZIPWriterPoolParallel(b *testing.B) {
	benchmarkWriterPoolParallel(b, NewGZIPWriterPool)
}

func benchmarkWriterPoolParallel(b *testing.B, newPool func() WriterPool) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			data := bytes.Repeat([]byte("a"), size)
			pool := newPool()

			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				var dst bytes.Buffer
				for pb.Next() {
					dst.Reset()
					n, err := pool.Write(&dst, data)
					if err != nil {
						b.Fatal(err)
					}
					if n != int64(len(data)) {
						b.Fatalf("wrote %d bytes, want %d", n, len(data))
					}
				}
			})
		})
	}
}

func gunzipData(tb testing.TB, data []byte) []byte {
	tb.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(tb, err)
	decompressed, err := io.ReadAll(reader)
	require.NoError(tb, err)
	require.NoError(tb, reader.Close())
	return decompressed
}

type testResetWriteCloser struct {
	dst        io.Writer
	writeErr   error
	closeErr   error
	closeCalls int
}

func (w *testResetWriteCloser) Write(data []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.dst.Write(data)
}

func (w *testResetWriteCloser) Close() error {
	w.closeCalls++
	return w.closeErr
}

func (w *testResetWriteCloser) Reset(dst io.Writer) {
	w.dst = dst
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}
