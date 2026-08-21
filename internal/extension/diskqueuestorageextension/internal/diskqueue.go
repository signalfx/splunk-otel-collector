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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Interface is a key-addressable, disk-backed record store. Keys are assigned
// by the caller (they map 1:1 to the persistent queue's item indices) so that
// Get and Delete can operate on a specific record regardless of write or
// delete order.
type Interface interface {
	Put(key uint64, data []byte) error
	Get(key uint64) ([]byte, error)
	Delete(key uint64) error
	Depth() int64
	Close() error
}

// location records where a record physically lives on disk.
type location struct {
	fileNum int64
	offset  int64
	length  int64
}

// diskQueue implements a filesystem backed, key-addressable record store.
// Records are appended to sequentially numbered segment files; an in-memory
// index (persisted alongside the data) maps each key to its segment file and
// byte range, so a record can be read or deleted independently of any other
// record, in any order.
type diskQueue struct {
	index           map[uint64]location
	liveCount       map[int64]int64
	writeFile       *os.File
	doneChan        chan struct{}
	exitChan        chan struct{}
	logger          *zap.Logger
	dataPath        string
	name            string
	writeFileNum    int64
	pendingSync     int64
	maxBytesPerFile int64
	syncEvery       int64
	syncTimeout     time.Duration
	writePos        int64
	mu              sync.Mutex
	closed          bool
}

// New instantiates a diskQueue, restoring its index from the filesystem if
// present, and starts a background goroutine that periodically flushes
// pending writes according to syncEvery/syncTimeout.
func New(name, dataPath string, maxBytesPerFile int64,
	syncEvery int64, syncTimeout time.Duration, logger *zap.Logger,
) Interface {
	d := &diskQueue{
		name:            name,
		dataPath:        dataPath,
		maxBytesPerFile: maxBytesPerFile,
		syncEvery:       syncEvery,
		syncTimeout:     syncTimeout,
		logger:          logger,
		index:           make(map[uint64]location),
		liveCount:       make(map[int64]int64),
		exitChan:        make(chan struct{}),
		doneChan:        make(chan struct{}),
	}

	if err := d.retrieveIndex(); err != nil && !os.IsNotExist(err) {
		d.logger.Error("failed to retrieve index", zap.String("name", d.name), zap.Error(err))
	}

	go d.syncLoop()
	return d
}

// Depth returns the number of live (not yet deleted) records.
func (d *diskQueue) Depth() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return int64(len(d.index))
}

// Put appends data to the current segment file and records its location
// under key, making it independently retrievable and deletable.
func (d *diskQueue) Put(key uint64, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return errors.New("exiting")
	}

	loc, err := d.writeOne(data)
	if err != nil {
		return err
	}

	d.index[key] = loc
	d.liveCount[loc.fileNum]++
	d.pendingSync++

	if d.pendingSync >= d.syncEvery {
		if syncErr := d.syncLocked(); syncErr != nil {
			d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(syncErr))
		}
	}

	return nil
}

// Get returns the data stored under key, or (nil, nil) if key is not present.
func (d *diskQueue) Get(key uint64) ([]byte, error) {
	d.mu.Lock()
	loc, ok := d.index[key]
	d.mu.Unlock()

	if !ok {
		return nil, nil
	}

	f, err := os.OpenFile(d.fileName(loc.fileNum), os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	buf := make([]byte, loc.length)
	if _, err := f.ReadAt(buf, loc.offset); err != nil {
		return nil, err
	}

	return buf, nil
}

// Delete removes key from the index. If it was the last live record in its
// segment file, and that file is no longer being written to, the file is
// removed from disk.
func (d *diskQueue) Delete(key uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	loc, ok := d.index[key]
	if !ok {
		return nil
	}

	delete(d.index, key)
	d.liveCount[loc.fileNum]--
	d.pendingSync++

	var err error
	if d.liveCount[loc.fileNum] <= 0 && loc.fileNum != d.writeFileNum {
		delete(d.liveCount, loc.fileNum)
		if removeErr := os.Remove(d.fileName(loc.fileNum)); removeErr != nil && !os.IsNotExist(removeErr) {
			err = removeErr
		}
	}

	if d.pendingSync >= d.syncEvery {
		if syncErr := d.syncLocked(); syncErr != nil {
			d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(syncErr))
		}
	}

	return err
}

// Close stops the background sync loop, persists any pending state, and
// closes the current write file.
func (d *diskQueue) Close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	d.mu.Unlock()

	close(d.exitChan)
	<-d.doneChan

	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.syncLocked()
	if d.writeFile != nil {
		if cerr := d.writeFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
		d.writeFile = nil
	}
	return err
}

// writeOne appends data to the current write file, rolling to a new segment
// file if it would exceed maxBytesPerFile. Callers must hold d.mu.
func (d *diskQueue) writeOne(data []byte) (location, error) {
	dataLen := int64(len(data))

	if d.writePos > 0 && d.writePos+dataLen > d.maxBytesPerFile {
		if d.writeFile != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
		}
		d.writeFileNum++
		d.writePos = 0
	}

	if d.writeFile == nil {
		fileName := d.fileName(d.writeFileNum)
		f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0o600)
		if err != nil {
			return location{}, err
		}

		d.logger.Info("opened", zap.String("name", d.name), zap.String("filename", fileName))

		if d.writePos > 0 {
			if _, err := f.Seek(d.writePos, 0); err != nil {
				_ = f.Close()
				return location{}, err
			}
		}
		d.writeFile = f
	}

	n, err := d.writeFile.Write(data)
	if err != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
		return location{}, err
	}

	loc := location{fileNum: d.writeFileNum, offset: d.writePos, length: int64(n)}
	d.writePos += int64(n)

	return loc, nil
}

// syncLocked fsyncs the current write file and persists the index. Callers
// must hold d.mu.
func (d *diskQueue) syncLocked() error {
	var errs []error

	if d.writeFile != nil {
		if err := d.writeFile.Sync(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := d.persistIndexLocked(); err != nil {
		errs = append(errs, err)
	}

	d.pendingSync = 0

	return errors.Join(errs...)
}

// syncLoop periodically flushes pending writes so that data is not held
// unsynced indefinitely when syncEvery has not been reached.
func (d *diskQueue) syncLoop() {
	ticker := time.NewTicker(d.syncTimeout)
	defer ticker.Stop()
	defer close(d.doneChan)

	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			if d.pendingSync > 0 {
				if err := d.syncLocked(); err != nil {
					d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(err))
				}
			}
			d.mu.Unlock()
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) fileName(fileNum int64) string {
	return filepath.Join(d.dataPath, fmt.Sprintf("%s.diskqueue.%06d.dat", d.name, fileNum))
}

func (d *diskQueue) indexFilePath() string {
	return filepath.Join(d.dataPath, d.name+".diskqueue.index.dat")
}

// persistIndexLocked atomically writes the write position and the full index
// to disk. Callers must hold d.mu.
func (d *diskQueue) persistIndexLocked() error {
	fileName := d.indexFilePath()
	f, err := os.CreateTemp(d.dataPath, filepath.Base(fileName)+"-*")
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	if _, err := fmt.Fprintf(w, "%d,%d\n%d\n", d.writeFileNum, d.writePos, len(d.index)); err != nil {
		_ = f.Close()
		return err
	}
	for key, loc := range d.index {
		if _, err := fmt.Fprintf(w, "%d,%d,%d,%d\n", key, loc.fileNum, loc.offset, loc.length); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(f.Name(), fileName)
}

// retrieveIndex restores the write position and index from the filesystem.
func (d *diskQueue) retrieveIndex() error {
	f, err := os.Open(d.indexFilePath())
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	r := bufio.NewReader(f)

	if _, err := fmt.Fscanf(r, "%d,%d\n", &d.writeFileNum, &d.writePos); err != nil {
		return err
	}

	var count int64
	if _, err := fmt.Fscanf(r, "%d\n", &count); err != nil {
		return err
	}

	index := make(map[uint64]location, count)
	liveCount := make(map[int64]int64)
	for i := int64(0); i < count; i++ {
		var key uint64
		var loc location
		if _, err := fmt.Fscanf(r, "%d,%d,%d,%d\n", &key, &loc.fileNum, &loc.offset, &loc.length); err != nil {
			return err
		}
		index[key] = loc
		liveCount[loc.fileNum]++
	}

	d.index = index
	d.liveCount = liveCount

	return nil
}
