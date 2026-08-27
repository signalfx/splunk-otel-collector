package record

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableAppendRecord(t *testing.T) {
	tests := []struct {
		name       string
		records    [][]byte
		expectHead []segment
		expectData []byte
	}{
		{
			name:       "empty record",
			records:    [][]byte{nil},
			expectHead: []segment{{offset: 0, size: 0}},
		},
		{
			name:       "multiple records",
			records:    [][]byte{[]byte("first"), []byte("second")},
			expectHead: []segment{{offset: 0, size: 5}, {offset: 5, size: 6}},
			expectData: []byte("firstsecond"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tab Table
			for _, record := range tt.records {
				require.NoError(t, tab.AppendRecord(record))
			}

			assert.Equal(t, tt.expectHead, tab.header)
			assert.Equal(t, tt.expectData, tab.data)
		})
	}
}

func TestTableAppendRecordCopiesData(t *testing.T) {
	record := []byte("record")
	var tab Table
	require.NoError(t, tab.AppendRecord(record))

	record[0] = 'R'

	assert.Equal(t, []byte("record"), tab.data)
}

func TestTableLen(t *testing.T) {
	tests := []struct {
		name    string
		records [][]byte
		expect  int
	}{
		{name: "empty", expect: 0},
		{name: "records", records: [][]byte{[]byte("first"), []byte("second")}, expect: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tab Table
			for _, record := range tt.records {
				require.NoError(t, tab.AppendRecord(record))
			}

			assert.Equal(t, tt.expect, tab.Len())
		})
	}
}

func TestTableAt(t *testing.T) {
	var tab Table
	require.NoError(t, tab.AppendRecord([]byte("first")))
	require.NoError(t, tab.AppendRecord([]byte("second")))

	tests := []struct {
		name      string
		index     int
		expect    []byte
		expectErr bool
	}{
		{name: "first record", index: 0, expect: []byte("first")},
		{name: "second record", index: 1, expect: []byte("second")},
		{name: "negative index", index: -1, expectErr: true},
		{name: "length index", index: 2, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tab.At(tt.index)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestTableAtReturnsDataView(t *testing.T) {
	var tab Table
	require.NoError(t, tab.AppendRecord([]byte("record")))

	record, err := tab.At(0)
	require.NoError(t, err)
	record[0] = 'R'

	assert.Equal(t, []byte("Record"), tab.data)
}

func TestTableReset(t *testing.T) {
	var tab Table
	require.NoError(t, tab.AppendRecord([]byte("record")))

	tab.Reset()

	assert.Equal(t, []segment{}, tab.header)
	assert.Equal(t, []byte{}, tab.data)
}

func TestTableMarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		tab  Table
	}{
		{
			name: "empty table",
		},
		{
			name: "multiple records",
			tab: Table{
				header: []segment{{offset: 0, size: 5}, {offset: 5, size: 6}},
				data:   []byte("firstsecond"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tab.MarshalBinary()
			require.NoError(t, err)

			assert.Equal(t, marshalTable(tt.tab), got)
			assert.Equal(t, tt.tab.Size(), len(got))
		})
	}
}

func TestTableAppendBinary(t *testing.T) {
	tab := Table{
		header: []segment{{offset: 0, size: 5}},
		data:   []byte("first"),
	}
	prefix := []byte("prefix")
	originalPrefix := append([]byte(nil), prefix...)

	got, err := tab.AppendBinary(prefix)
	require.NoError(t, err)

	want := append(originalPrefix, marshalTable(tab)...)
	assert.Equal(t, want, got)
	assert.Equal(t, originalPrefix, prefix)
}

func TestTableUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		expectHead []segment
		expectData []byte
	}{
		{
			name:       "empty table",
			data:       marshalTable(Table{}),
			expectHead: []segment{},
			expectData: []byte{},
		},
		{
			name: "multiple records",
			data: marshalTable(Table{
				header: []segment{{offset: 0, size: 5}, {offset: 5, size: 6}},
				data:   []byte("firstsecond"),
			}),
			expectHead: []segment{{offset: 0, size: 5}, {offset: 5, size: 6}},
			expectData: []byte("firstsecond"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tab Table
			require.NoError(t, tab.UnmarshalBinary(tt.data))

			assert.Equal(t, tt.expectHead, tab.header)
			assert.Equal(t, tt.expectData, tab.data)
		})
	}
}

func TestTableUnmarshalBinaryReplacesExistingContents(t *testing.T) {
	tab := Table{
		header: []segment{{offset: 0, size: 3}},
		data:   []byte("old"),
	}
	encoded := marshalTable(Table{
		header: []segment{{offset: 0, size: 5}},
		data:   []byte("fresh"),
	})

	require.NoError(t, tab.UnmarshalBinary(encoded))

	assert.Equal(t, []segment{{offset: 0, size: 5}}, tab.header)
	assert.Equal(t, []byte("fresh"), tab.data)
}

func TestTableUnmarshalBinaryCopiesInput(t *testing.T) {
	encoded := marshalTable(Table{
		header: []segment{{offset: 0, size: 6}},
		data:   []byte("record"),
	})
	var tab Table
	require.NoError(t, tab.UnmarshalBinary(encoded))

	encoded[len(encoded)-1] = 'X'

	assert.Equal(t, []byte("record"), tab.data)
}

func TestTableUnmarshalBinaryErrors(t *testing.T) {
	truncatedCount := []byte{0, 0, 0, 0, 0, 0, 0}
	missingSegment := binary.NativeEndian.AppendUint64(nil, 1)
	invalidSegment := binary.NativeEndian.AppendUint64(nil, 1)
	invalidSegment = binary.NativeEndian.AppendUint32(invalidSegment, 2)
	invalidSegment = binary.NativeEndian.AppendUint32(invalidSegment, 2)
	invalidSegment = append(invalidSegment, 0, 0, 0)
	tooManySegments := binary.NativeEndian.AppendUint64(nil, ^uint64(0))

	tests := []struct {
		name string
		data []byte
	}{
		{name: "truncated count", data: truncatedCount},
		{name: "missing segment", data: missingSegment},
		{name: "invalid segment bounds", data: invalidSegment},
		{name: "too many segments", data: tooManySegments},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tab Table
			assert.Error(t, tab.UnmarshalBinary(tt.data))
		})
	}
}

func TestTableUnmarshalBinaryIsTransactional(t *testing.T) {
	tab := Table{
		header: []segment{{offset: 0, size: 3}},
		data:   []byte("old"),
	}
	invalid := binary.NativeEndian.AppendUint64(nil, 1)

	assert.Error(t, tab.UnmarshalBinary(invalid))

	assert.Equal(t, []segment{{offset: 0, size: 3}}, tab.header)
	assert.Equal(t, []byte("old"), tab.data)
}

func TestTableWriteTo(t *testing.T) {
	tab := Table{
		header: []segment{{offset: 0, size: 5}},
		data:   []byte("first"),
	}
	var dst bytes.Buffer

	n, err := tab.WriteTo(&dst)
	require.NoError(t, err)

	want := marshalTable(tab)
	assert.Equal(t, int64(len(want)), n)
	assert.Equal(t, want, dst.Bytes())
}

func TestTableReadFrom(t *testing.T) {
	want := Table{
		header: []segment{{offset: 0, size: 5}},
		data:   []byte("first"),
	}
	var tab Table

	n, err := tab.ReadFrom(bytes.NewReader(marshalTable(want)))
	require.NoError(t, err)

	assert.Equal(t, int64(want.Size()), n)
	assert.Equal(t, want.header, tab.header)
	assert.Equal(t, want.data, tab.data)
}

func marshalTable(tab Table) []byte {
	buf := make([]byte, 0, tab.Size())
	buf = binary.NativeEndian.AppendUint64(buf, uint64(len(tab.header)))
	for _, seg := range tab.header {
		buf = binary.NativeEndian.AppendUint32(buf, seg.offset)
		buf = binary.NativeEndian.AppendUint32(buf, seg.size)
	}
	return append(buf, tab.data...)
}

func benchmarkTable(tb testing.TB, size int) Table {
	tb.Helper()
	var tab Table
	require.NoError(tb, tab.AppendRecord(make([]byte, size)))
	return tab
}

func benchmarkSizeName(size int) string {
	switch size {
	case 1 << 10:
		return "1KiB"
	case 64 << 10:
		return "64KiB"
	case 1 << 20:
		return "1MiB"
	case 100 << 20:
		return "100MiB"
	default:
		return "unknown"
	}
}

var (
	benchmarkBytes []byte
	benchmarkCount int64
)

func BenchmarkTableMarshalBinary(b *testing.B) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20, 100 << 20} {
		b.Run(benchmarkSizeName(size), func(b *testing.B) {
			tab := benchmarkTable(b, size)
			b.SetBytes(int64(tab.Size()))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				var err error
				benchmarkBytes, err = tab.MarshalBinary()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTableAppendBinary(b *testing.B) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20, 100 << 20} {
		b.Run(benchmarkSizeName(size), func(b *testing.B) {
			tab := benchmarkTable(b, size)
			buf := make([]byte, 0, tab.Size())
			b.SetBytes(int64(tab.Size()))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				var err error
				buf, err = tab.AppendBinary(buf[:0])
				if err != nil {
					b.Fatal(err)
				}
			}
			benchmarkBytes = buf
		})
	}
}

func BenchmarkTableWriteTo(b *testing.B) {
	for _, size := range []int{1 << 10, 64 << 10, 1 << 20, 100 << 20} {
		b.Run(benchmarkSizeName(size), func(b *testing.B) {
			tab := benchmarkTable(b, size)
			var dst bytes.Buffer
			dst.Grow(tab.Size())
			b.SetBytes(int64(tab.Size()))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				dst.Reset()
				var err error
				benchmarkCount, err = tab.WriteTo(&dst)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
