package disk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension/internal/record"
)

func TestRecordQueueBatch(t *testing.T) {
	tests := []struct {
		name            string
		max             int
		batch           [][]byte
		expectAccepted  int
		expectPartition int
		expectPending   int
		expectActiveLen int
		expectFiles     [][][]byte
	}{
		{
			name:            "keeps active table below soft limit",
			max:             25,
			batch:           [][]byte{[]byte("first")},
			expectAccepted:  1,
			expectActiveLen: 1,
		},
		{
			name:            "persists table after crossing soft limit",
			max:             25,
			batch:           [][]byte{[]byte("first"), []byte("second")},
			expectAccepted:  2,
			expectPartition: 1,
			expectFiles:     [][][]byte{{[]byte("first"), []byte("second")}},
		},
		{
			name:            "persists every rotated table",
			max:             16,
			batch:           [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			expectAccepted:  3,
			expectPartition: 3,
			expectFiles:     [][][]byte{{[]byte("a")}, {[]byte("b")}, {[]byte("c")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			rq := NewRecordQueue(dir, "test", tt.max, zap.NewNop())
			require.NoError(t, rq.Start(context.Background()))
			t.Cleanup(func() {
				stopPersistenceWorker(rq)
				_ = rq.root.Close()
			})

			accepted, err := rq.Batch(tt.batch...)
			require.NoError(t, err)
			if tt.expectPartition > 0 {
				waitForPersisted(t, rq, tt.expectPartition)
			}
			assert.Equal(t, tt.expectAccepted, accepted)
			assert.Equal(t, tt.expectPartition, rq.partition)
			assert.Equal(t, tt.expectPending, len(rq.pending))
			assert.Equal(t, tt.expectActiveLen, rq.active.Len())

			for partition, expectedRecords := range tt.expectFiles {
				got := readPersistedTable(t, dir, "test", partition)
				assert.Equal(t, len(expectedRecords), got.Len())
				for index, expected := range expectedRecords {
					recordData, err := got.At(index)
					require.NoError(t, err)
					assert.Equal(t, expected, recordData)
				}
			}
		})
	}
}

func TestRecordQueueBatchReportsAcceptedRecordsWhenPersistenceFails(t *testing.T) {
	dir := t.TempDir()
	rq := NewRecordQueue(dir, "test", 0, zap.NewNop())
	require.NoError(t, rq.Start(context.Background()))
	require.NoError(t, rq.root.Close())

	accepted, err := rq.Batch([]byte("record"))

	assert.Equal(t, 1, accepted)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		rq.write.Lock()
		defer rq.write.Unlock()
		return len(rq.pending) == 1
	}, time.Second, time.Millisecond)
	assert.Equal(t, 1, len(rq.pending))
	stopPersistenceWorker(rq)
}

func BenchmarkRecordQueueBatch(b *testing.B) {
	for _, payloadSize := range []int{1 << 10, 64 << 10, 1 << 20, 100 << 20} {
		b.Run(fmt.Sprintf("%dB", payloadSize), func(b *testing.B) {
			dir := b.TempDir()
			rq := NewRecordQueue(dir, "benchmark", payloadSize, zap.NewNop())
			if err := rq.Start(context.Background()); err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() {
				stopPersistenceWorker(rq)
				_ = rq.root.Close()
			})

			payload := make([]byte, payloadSize)
			b.SetBytes(int64(payloadSize))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := rq.Batch(payload); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func waitForPersisted(t *testing.T, rq *RecordQueue, partition int) {
	t.Helper()

	require.Eventually(t, func() bool {
		rq.write.Lock()
		defer rq.write.Unlock()
		return rq.partition == partition && len(rq.pending) == 0
	}, time.Second, time.Millisecond)
}

func stopPersistenceWorker(rq *RecordQueue) {
	rq.running.Store(false)
	rq.write.Lock()
	rq.fsync.Broadcast()
	rq.write.Unlock()
	_ = rq.wg.Wait()
}

func readPersistedTable(t *testing.T, dir, name string, partition int) *record.Table {
	t.Helper()

	fileName := fmt.Sprintf("%s.diskqueue.%08d.dat", name, partition)
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	require.NoError(t, err)

	tab := record.NewTable()
	require.NoError(t, tab.UnmarshalBinary(data))
	return tab
}
