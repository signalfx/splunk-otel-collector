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
	"io"
	"sync"

	"github.com/pierrec/lz4/v4"
)

// ResetWriteCloser is an io.WriteCloser that can be reset to write to another destination.
type ResetWriteCloser interface {
	io.Closer
	io.Writer

	// Reset prepares the writer to write to w.
	Reset(w io.Writer)
}

// WriterPool reuses ResetWriteClosers to copy data from readers into writers.
type WriterPool interface {
	// Copy resets a pooled writer with dst and copies src through it.
	Copy(dst io.Writer, src io.Reader) (int64, error)
	// Write copies data from the byte slice into dst using a pooled writer.
	Write(dst io.Writer, data []byte) (int64, error)
}

type writerPool[RWC ResetWriteCloser] struct {
	_    struct{}
	pool sync.Pool
}

// NewGZIPWriterPool returns a WriterPool that compresses data with gzip.BestSpeed.
func NewGZIPWriterPool() WriterPool {
	return NewWriterPool(func() *gzip.Writer {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	})
}

// NewLZ4WriterPool returns a WriterPool that compresses data with LZ4.
func NewLZ4WriterPool() WriterPool {
	return NewWriterPool(func() *lz4.Writer { return lz4.NewWriter(io.Discard) })
}

// NewWriterPool returns a WriterPool that creates pooled writers with factory.
func NewWriterPool[RWC ResetWriteCloser](factory func() RWC) WriterPool {
	return &writerPool[RWC]{
		pool: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

func (wp *writerPool[RWC]) Copy(dst io.Writer, src io.Reader) (int64, error) {
	compressor := wp.pool.Get().(RWC)
	defer wp.pool.Put(compressor)

	compressor.Reset(dst)

	n, err := io.Copy(compressor, src)
	errs := errors.Join(err, compressor.Close())

	return n, errs
}

func (wp *writerPool[RWC]) Write(dst io.Writer, data []byte) (int64, error) {
	return wp.Copy(dst, bytes.NewReader(data))
}
