package record

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
)

type segment struct {
	offset uint32
	size   uint32
}

// Table stores records contiguously with a private index describing each
// record's offset and size.
//
// The binary representation consists of an eight-byte record count, followed
// by one eight-byte offset-and-size entry per record, followed by the record
// data.
type Table struct {
	header []segment
	data   []byte
}

var (
	_ encoding.BinaryMarshaler   = (*Table)(nil)
	_ encoding.BinaryAppender    = (*Table)(nil)
	_ encoding.BinaryUnmarshaler = (*Table)(nil)
	_ io.WriterTo                = (*Table)(nil)
	_ io.ReaderFrom              = (*Table)(nil)
)

func NewTable() *Table {
	return &Table{
		header: make([]segment, 0),
		data:   make([]byte, 0),
	}
}

// Len returns the number of records in the table.
func (tab *Table) Len() int {
	return len(tab.header)
}

// At returns the record at index i.
//
// The returned slice aliases the table's data and should not be retained after
// the table is modified or reset.
func (tab *Table) At(i int) ([]byte, error) {
	if i < 0 || i >= len(tab.header) {
		return nil, fmt.Errorf("record index out of bounds: %d", i)
	}
	seg := tab.header[i]
	return tab.data[seg.offset : seg.offset+seg.size], nil
}

// Size returns the number of bytes used by the table's binary representation.
func (tab *Table) Size() int {
	return 8 + 8*len(tab.header) + len(tab.data)
}

// Reset clears the table while retaining its allocated storage for reuse.
func (tab *Table) Reset() {
	clear(tab.header)
	clear(tab.data)
	tab.header = tab.header[:0]
	tab.data = tab.data[:0]
}

// AppendRecord appends data to the table and records its offset and size.
//
// The data is copied into the table, so the caller may safely reuse or modify
// data after AppendRecord returns. AppendRecord returns an error if the table
// would exceed the uint32 offset and size limits.
//
// Table is intended to be used by a single owner and is not concurrency-safe.
func (tab *Table) AppendRecord(data []byte) error {
	var (
		offset = uint64(len(tab.data))
		size   = uint64(len(data))
	)

	if offset > math.MaxUint32 || size > math.MaxUint32-offset {
		return errors.New("record exceeds table size limit")
	}

	tab.header = append(tab.header, segment{
		offset: uint32(len(tab.data)),
		size:   uint32(len(data)),
	})
	tab.data = append(tab.data, data...)
	return nil
}

// AppendBinary appends the table's binary representation to dst without
// modifying the bytes already present in dst.
func (tab *Table) AppendBinary(dst []byte) ([]byte, error) {
	dst = slices.Grow(dst, tab.Size())
	dst = binary.NativeEndian.AppendUint64(dst, uint64(len(tab.header)))
	for _, seg := range tab.header {
		dst = binary.NativeEndian.AppendUint32(dst, seg.offset)
		dst = binary.NativeEndian.AppendUint32(dst, seg.size)
	}
	return append(dst, tab.data...), nil
}

// MarshalBinary returns a new slice containing the table's binary
// representation.
func (tab *Table) MarshalBinary() ([]byte, error) {
	return tab.AppendBinary(make([]byte, 0, tab.Size()))
}

// UnmarshalBinary replaces the table with the binary representation in p.
//
// The input is copied, and the table is left unchanged if decoding fails.
func (tab *Table) UnmarshalBinary(p []byte) error {
	const (
		countSize   = 8
		segmentSize = 8
	)

	if len(p) < countSize {
		return fmt.Errorf("missing record count")
	}

	count := binary.NativeEndian.Uint64(p[:countSize])
	availableSegments := uint64((len(p) - countSize) / segmentSize)
	if count > availableSegments {
		return fmt.Errorf("record count %d exceeds available segment data: %w", count, io.ErrUnexpectedEOF)
	}

	header := make([]segment, int(count))
	position := countSize
	for i := range header {
		header[i] = segment{
			offset: binary.NativeEndian.Uint32(p[position:]),
			size:   binary.NativeEndian.Uint32(p[position+4:]),
		}
		position += segmentSize
	}

	data := p[position:]
	for i, seg := range header {
		end := uint64(seg.offset) + uint64(seg.size)
		if end > uint64(len(data)) {
			return fmt.Errorf(
				"segment %d exceeds data bounds: offset=%d size=%d data=%d",
				i, seg.offset, seg.size, len(data),
			)
		}
	}

	tab.header = header
	tab.data = slices.Clone(data)
	return nil
}

// WriteTo writes the table's binary representation to dst and returns the
// number of bytes written.
func (tab *Table) WriteTo(dst io.Writer) (int64, error) {
	buf, err := tab.MarshalBinary()
	if err != nil {
		return 0, err
	}
	r := bytes.NewReader(buf)
	return io.Copy(dst, r)
}

// ReadFrom reads a complete binary table from src and replaces the table's
// current contents.
func (tab *Table) ReadFrom(src io.Reader) (int64, error) {
	buf := bytes.NewBuffer(nil)
	n, err := buf.ReadFrom(src)
	if err != nil {
		return n, err
	}
	err = tab.UnmarshalBinary(buf.Bytes())
	return n, err
}
