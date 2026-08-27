package disk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension/internal/record"
)

type RecordQueue struct {
	_ struct{}

	max  int
	dir  string
	name string

	pool sync.Pool
	wg   errgroup.Group

	root    *os.Root
	logger  *zap.Logger
	running atomic.Bool

	write     sync.Mutex
	fsync     *sync.Cond
	partition int
	active    *record.Table
	pending   []*record.Table
}

func NewRecordQueue(dir, name string, max int, logger *zap.Logger) *RecordQueue {
	rq := &RecordQueue{
		max:  max,
		dir:  dir,
		name: name,
		pool: sync.Pool{
			New: func() any {
				return record.NewTable()
			},
		},
		logger:  logger,
		active:  record.NewTable(),
		pending: make([]*record.Table, 0),
	}
	rq.fsync = sync.NewCond(&rq.write)
	return rq
}

func (rq *RecordQueue) Start(ctx context.Context) error {
	if rq.running.Swap(true) {
		return errors.New("queue already running")
	}
	var (
		err error
	)

	rq.root, err = os.OpenRoot(rq.dir)
	if err != nil {
		rq.running.Store(false)
		return err
	}

	rq.wg.Go(func() error {
		return rq.persist()
	})

	// TODO
	// Will need to walk the current fs to see what is the latest partition
	// that is currently written on disk to so that we can continue on from there.

	return nil
}

func (rq *RecordQueue) Shutdown(ctx context.Context) error {
	if !rq.running.Swap(false) {
		return errors.New("queue already shutdown")
	}
	// TODO
	// Need to persist any state that is currently
	// held incase it is needed when after a restart.
	return errors.Join(
		rq.root.Close(),
		rq.wg.Wait(),
	)
}

func (rq *RecordQueue) Put(buf []byte) error {
	_, err := rq.Batch(buf)
	return err
}

func (rq *RecordQueue) Batch(batch ...[]byte) (int, error) {
	if !rq.running.Load() {
		return 0, errors.New("queue not started")
	}

	rq.write.Lock()
	defer rq.write.Unlock()

	rotated := false
	for i, b := range batch {
		if err := rq.active.AppendRecord(b); err != nil {
			return i, err
		}
		if rq.active.Size() <= rq.max {
			continue
		}
		rq.pending = append(rq.pending, rq.active)
		active := rq.pool.Get().(*record.Table)
		active.Reset()
		rq.active = active
		rq.logger.Debug(
			"Rotated active table",
			zap.Int("pending", len(rq.pending)),
		)
		rotated = true
	}
	if rotated {
		rq.fsync.Signal()
	}
	return len(batch), nil
}

func (rq *RecordQueue) Get(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	default:
		// Need to read from a notifier queue
	}
	return nil, nil
}

func (rq *RecordQueue) persist() error {
	for {
		rq.write.Lock()
		for len(rq.pending) == 0 && rq.running.Load() {
			rq.fsync.Wait()
		}

		if !rq.running.Load() {
			rq.write.Unlock()
			return nil
		}

		tab := rq.pending[0]
		partition := rq.partition
		rq.write.Unlock()

		finalName := fmt.Sprintf(
			"%s.diskqueue.%08d.dat",
			rq.name,
			partition,
		)
		err := rq.writeTable(tab, finalName)

		if err != nil {
			rq.logger.Error(
				"Failed to persist pending table",
				zap.Int("partition", partition),
				zap.Error(err),
			)

			continue
		}
		rq.write.Lock()
		rq.pending = rq.pending[1:]
		rq.partition++
		rq.pool.Put(tab)
		rq.write.Unlock()
	}
}

func (rq *RecordQueue) writeTable(tab *record.Table, finalName string) error {
	tempName := finalName + ".tmp"

	f, err := rq.root.OpenFile(
		tempName,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0600,
	)
	if err != nil {
		return err
	}

	n, err := tab.WriteTo(f)
	if err == nil && n != int64(tab.Size()) {
		err = fmt.Errorf(
			"file %q: wrote %d bytes, expected %d",
			tempName,
			n,
			tab.Size(),
		)
	}

	if err := errors.Join(err, f.Sync(), f.Close()); err != nil {
		_ = rq.root.Remove(tempName)
		return err
	}

	if err := rq.root.Rename(tempName, finalName); err != nil {
		_ = rq.root.Remove(tempName)
		return err
	}

	return nil
}
