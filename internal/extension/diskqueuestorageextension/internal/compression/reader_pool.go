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
)

// ResetReadCloser is an io.ReadCloser that can be reset to read from another source.
type ResetReadCloser interface {
	io.ReadCloser

	// Reset prepares the reader to read from r.
	Reset(r io.Reader) error
}

// ReaderPool reuses ResetReadClosers to copy data from readers into writers.
type ReaderPool interface {
	// Copy resets a pooled reader with src and copies its output to dst.
	Copy(dst io.Writer, src io.Reader) (int64, error)
	// Read copies data from the byte slice into dst using a pooled reader.
	Read(dst io.Writer, data []byte) (int64, error)
}

type readerPool[RRC ResetReadCloser] struct {
	_    struct{}
	pool sync.Pool
}

// NewGZIPReaderPool returns a ReaderPool that decompresses gzip data.
func NewGZIPReaderPool() ReaderPool {
	return NewReaderPool(func() *gzip.Reader { return new(gzip.Reader) })
}

// NewReaderPool returns a ReaderPool that creates pooled readers with factory.
func NewReaderPool[RRC ResetReadCloser](factory func() RRC) ReaderPool {
	return &readerPool[RRC]{
		pool: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

func (rp *readerPool[RRC]) Read(dst io.Writer, data []byte) (int64, error) {
	return rp.Copy(dst, bytes.NewReader(data))
}

func (rp *readerPool[RRC]) Copy(dst io.Writer, src io.Reader) (int64, error) {
	compressor := rp.pool.Get().(RRC)
	defer rp.pool.Put(compressor)

	if err := compressor.Reset(src); err != nil {
		return 0, err
	}

	n, err := io.Copy(dst, compressor)
	errs := errors.Join(err, compressor.Close())

	return n, errs
}
