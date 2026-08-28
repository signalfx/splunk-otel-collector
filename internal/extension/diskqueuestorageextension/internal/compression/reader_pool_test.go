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

func TestGZIPReaderPool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		read bool
	}{
		{name: "empty", read: true},
		{name: "small", data: "hello, world", read: true},
		{name: "copy", data: "copy compressed data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			compressed := gzipData(t, []byte(tt.data))

			var dst bytes.Buffer
			pool := NewGZIPReaderPool()
			var (
				n   int64
				err error
			)
			if tt.read {
				n, err = pool.Read(&dst, compressed)
			} else {
				n, err = pool.Copy(&dst, bytes.NewReader(compressed))
			}

			require.NoError(t, err)
			require.Equal(t, int64(len(tt.data)), n)
			require.Equal(t, tt.data, dst.String())
		})
	}
}

func TestReaderPoolErrors(t *testing.T) {
	t.Parallel()

	resetErr := errors.New("reset failed")
	readErr := errors.New("read failed")
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")

	tests := []struct {
		dst         io.Writer
		reader      *testResetReadCloser
		name        string
		expectedErr []error
		expectedN   int64
	}{
		{
			name:        "reset failure",
			reader:      &testResetReadCloser{resetErr: resetErr, closeErr: closeErr},
			dst:         io.Discard,
			expectedErr: []error{resetErr},
		},
		{
			name:        "read and close failures",
			reader:      &testResetReadCloser{readErr: readErr, closeErr: closeErr},
			dst:         io.Discard,
			expectedErr: []error{readErr, closeErr},
		},
		{
			name:        "write and close failures",
			reader:      &testResetReadCloser{data: []byte("data"), closeErr: closeErr},
			dst:         errorWriter{err: writeErr},
			expectedErr: []error{writeErr, closeErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool := NewReaderPool(func() *testResetReadCloser { return tt.reader })
			n, err := pool.Copy(tt.dst, bytes.NewReader(nil))

			require.Equal(t, tt.expectedN, n)
			for _, expectedErr := range tt.expectedErr {
				require.ErrorIs(t, err, expectedErr)
			}
			if tt.reader.resetErr != nil {
				require.Equal(t, 0, tt.reader.closeCalls)
			} else {
				require.Equal(t, 1, tt.reader.closeCalls)
			}
		})
	}
}

func BenchmarkGZIPReaderPool(b *testing.B) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			data := bytes.Repeat([]byte("a"), size)
			compressed := gzipData(b, data)
			pool := NewGZIPReaderPool()
			var dst bytes.Buffer

			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				dst.Reset()
				n, err := pool.Read(&dst, compressed)
				if err != nil {
					b.Fatal(err)
				}
				if n != int64(len(data)) {
					b.Fatalf("read %d bytes, want %d", n, len(data))
				}
			}
		})
	}
}

func BenchmarkGZIPReaderPoolParallel(b *testing.B) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			data := bytes.Repeat([]byte("a"), size)
			compressed := gzipData(b, data)
			pool := NewGZIPReaderPool()

			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				var dst bytes.Buffer
				for pb.Next() {
					dst.Reset()
					n, err := pool.Read(&dst, compressed)
					if err != nil {
						b.Fatal(err)
					}
					if n != int64(len(data)) {
						b.Fatalf("read %d bytes, want %d", n, len(data))
					}
				}
			})
		})
	}
}

func gzipData(tb testing.TB, data []byte) []byte {
	tb.Helper()

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write(data)
	require.NoError(tb, err)
	require.NoError(tb, writer.Close())
	return compressed.Bytes()
}

type testResetReadCloser struct {
	readErr    error
	resetErr   error
	closeErr   error
	reader     io.Reader
	data       []byte
	closeCalls int
}

func (r *testResetReadCloser) Read(dst []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}
	return r.reader.Read(dst)
}

func (r *testResetReadCloser) Close() error {
	r.closeCalls++
	return r.closeErr
}

func (r *testResetReadCloser) Reset(io.Reader) error {
	if r.resetErr != nil {
		return r.resetErr
	}
	r.reader = bytes.NewReader(r.data)
	return nil
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
