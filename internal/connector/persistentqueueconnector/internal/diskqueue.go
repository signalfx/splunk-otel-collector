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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Interface interface {
	Put([]byte) error
	ReadChan() <-chan []byte // this is expected to be an *unbuffered* channel
	PeekChan() <-chan []byte // this is expected to be an *unbuffered* channel
	Close() error
	Delete() error
	Depth() int64
	Empty() error
}

// diskQueue implements a filesystem backed FIFO queue
type diskQueue struct {
	writeResponseChan   chan error
	writeChan           chan []byte
	exitSyncChan        chan int
	exitChan            chan int
	emptyResponseChan   chan error
	emptyChan           chan int
	logger              *zap.Logger
	writeFile           *os.File
	readFile            *os.File
	reader              *bufio.Reader
	depthChan           chan int64
	peekChan            chan []byte
	readChan            chan []byte
	dataPath            string
	name                string
	writeBuf            bytes.Buffer
	maxBytesPerFile     int64
	readFileNum         int64
	nextReadPos         int64
	writePos            int64
	nextReadFileNum     int64
	syncTimeout         time.Duration
	syncEvery           int64
	depth               int64
	writeFileNum        int64
	maxBytesPerFileRead int64
	readPos             int64
	maxMsgSize          int64
	minMsgSize          int64
	sync.RWMutex
	exitFlag int32
	needSync bool
}

// New instantiates an instance of diskQueue, retrieving metadata
// from the filesystem and starting the read ahead goroutine
func New(name, dataPath string, maxBytesPerFile int64,
	minMsgSize, maxMsgSize int64,
	syncEvery int64, syncTimeout time.Duration, logger *zap.Logger,
) Interface {
	d := diskQueue{
		name:              name,
		dataPath:          dataPath,
		maxBytesPerFile:   maxBytesPerFile,
		minMsgSize:        minMsgSize,
		maxMsgSize:        maxMsgSize,
		readChan:          make(chan []byte),
		peekChan:          make(chan []byte),
		depthChan:         make(chan int64),
		writeChan:         make(chan []byte),
		writeResponseChan: make(chan error),
		emptyChan:         make(chan int),
		emptyResponseChan: make(chan error),
		exitChan:          make(chan int),
		exitSyncChan:      make(chan int),
		syncEvery:         syncEvery,
		syncTimeout:       syncTimeout,
		logger:            logger,
	}

	// no need to lock here, nothing else could possibly be touching this instance
	err := d.retrieveMetaData()
	if err != nil && !os.IsNotExist(err) {
		d.logger.Error(" failed to retrieveMetaData", zap.String("name", d.name), zap.Error(err))
	}

	go d.ioLoop()
	return &d
}

// Depth returns the depth of the queue
func (d *diskQueue) Depth() int64 {
	depth, ok := <-d.depthChan
	if !ok {
		// ioLoop exited
		depth = d.depth
	}
	return depth
}

// ReadChan returns the receive-only []byte channel for reading data
func (d *diskQueue) ReadChan() <-chan []byte {
	return d.readChan
}

func (d *diskQueue) PeekChan() <-chan []byte {
	return d.peekChan
}

// Put writes a []byte to the queue
func (d *diskQueue) Put(data []byte) error {
	d.RLock()
	defer d.RUnlock()

	if d.exitFlag == 1 {
		return errors.New("exiting")
	}

	d.writeChan <- data
	return <-d.writeResponseChan
}

// Close cleans up the queue and persists metadata
func (d *diskQueue) Close() error {
	err := d.exit(false)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *diskQueue) Delete() error {
	return d.exit(true)
}

func (d *diskQueue) exit(deleted bool) error {
	d.Lock()
	defer d.Unlock()

	d.exitFlag = 1

	if deleted {
		d.logger.Info("deleting", zap.String("name", d.name))
	} else {
		d.logger.Info("closing", zap.String("name", d.name))
	}

	close(d.exitChan)
	// ensure that ioLoop has exited
	<-d.exitSyncChan

	close(d.depthChan)

	if d.readFile != nil {
		_ = d.readFile.Close()
		d.readFile = nil
	}

	if d.writeFile != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
	}

	return nil
}

// Empty destructively clears out any pending data in the queue
// by fast forwarding read positions and removing intermediate files
func (d *diskQueue) Empty() error {
	d.RLock()
	defer d.RUnlock()

	if d.exitFlag == 1 {
		return errors.New("exiting")
	}

	d.logger.Info("emptying", zap.String("name", d.name))

	d.emptyChan <- 1
	return <-d.emptyResponseChan
}

func (d *diskQueue) deleteAllFiles() error {
	err := d.skipToNextRWFile()

	innerErr := os.Remove(d.metaDataFileName())
	if innerErr != nil && !os.IsNotExist(innerErr) {
		d.logger.Error(" failed to remove metadata file", zap.String("name", d.name), zap.Error(innerErr))
		return innerErr
	}

	return err
}

func (d *diskQueue) skipToNextRWFile() error {
	var err error

	if d.readFile != nil {
		_ = d.readFile.Close()
		d.readFile = nil
	}

	if d.writeFile != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
	}

	for i := d.readFileNum; i <= d.writeFileNum; i++ {
		fn := d.fileName(i)
		innerErr := os.Remove(fn)
		if innerErr != nil && !os.IsNotExist(innerErr) {
			d.logger.Error(" failed to remove data file", zap.String("name", d.name), zap.Error(innerErr))
			err = innerErr
		}
	}

	d.writeFileNum++
	d.writePos = 0
	d.readFileNum = d.writeFileNum
	d.readPos = 0
	d.nextReadFileNum = d.writeFileNum
	d.nextReadPos = 0
	d.depth = 0

	return err
}

// readOne performs a low level filesystem read for a single []byte
// while advancing read positions and rolling files, if necessary
func (d *diskQueue) readOne() ([]byte, error) {
	var err error
	var msgSize int64

	if d.readFile == nil {
		curFileName := d.fileName(d.readFileNum)
		d.readFile, err = os.OpenFile(curFileName, os.O_RDONLY, 0o600)
		if err != nil {
			return nil, err
		}

		d.logger.Info("readOne() opened", zap.String("name", d.name), zap.String("filename", curFileName))

		if d.readPos > 0 {
			_, err = d.readFile.Seek(d.readPos, 0)
			if err != nil {
				_ = d.readFile.Close()
				d.readFile = nil
				return nil, err
			}
		}

		// for "complete" files (i.e. not the "current" file), maxBytesPerFileRead
		// should be initialized to the file's size, or default to maxBytesPerFile
		d.maxBytesPerFileRead = d.maxBytesPerFile
		if d.readFileNum < d.writeFileNum {
			stat, err := d.readFile.Stat()
			if err == nil {
				d.maxBytesPerFileRead = stat.Size()
			}
		}

		d.reader = bufio.NewReader(d.readFile)
	}

	err = binary.Read(d.reader, binary.BigEndian, &msgSize)
	if err != nil {
		_ = d.readFile.Close()
		d.readFile = nil
		return nil, err
	}

	if msgSize < d.minMsgSize || msgSize > d.maxMsgSize {
		// this file is corrupt and we have no reasonable guarantee on
		// where a new message should begin
		_ = d.readFile.Close()
		d.readFile = nil
		return nil, fmt.Errorf("invalid message read size (%d)", msgSize)
	}

	readBuf := make([]byte, msgSize)
	_, err = io.ReadFull(d.reader, readBuf)
	if err != nil {
		_ = d.readFile.Close()
		d.readFile = nil
		return nil, err
	}

	totalBytes := 8 + msgSize

	// we only advance next* because we have not yet sent this to consumers
	// (where readFileNum, readPos will actually be advanced)
	d.nextReadPos = d.readPos + totalBytes
	d.nextReadFileNum = d.readFileNum

	// we only consider rotating if we're reading a "complete" file
	// and since we cannot know the size at which it was rotated, we
	// rely on maxBytesPerFileRead rather than maxBytesPerFile
	if d.readFileNum <= d.writeFileNum && d.nextReadPos >= d.maxBytesPerFileRead {
		if d.readFile != nil {
			_ = d.readFile.Close()
			d.readFile = nil
		}

		d.nextReadFileNum++
		d.nextReadPos = 0
	}

	return readBuf, nil
}

// writeOne performs a low level filesystem write for a single []byte
// while advancing write positions and rolling files, if necessary
func (d *diskQueue) writeOne(data []byte) error {
	var err error

	dataLen := int64(len(data))
	totalBytes := 8 + dataLen

	if dataLen < d.minMsgSize || dataLen > d.maxMsgSize {
		return fmt.Errorf("invalid message write size (%d) minMsgSize=%d maxMsgSize=%d", dataLen, d.minMsgSize, d.maxMsgSize)
	}

	// will not wrap-around if maxBytesPerFile + maxMsgSize < Int64Max
	if d.writePos > 0 && d.writePos+totalBytes > d.maxBytesPerFile {
		if d.readFileNum == d.writeFileNum {
			d.maxBytesPerFileRead = d.writePos
		}

		d.writeFileNum++
		d.writePos = 0

		// sync every time we start writing to a new file
		err = d.sync()
		if err != nil {
			d.logger.Error(" failed to sync - %s", zap.String("name", d.name), zap.Error(err))
		}

		if d.writeFile != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
		}
	}

	if d.writeFile == nil {
		curFileName := d.fileName(d.writeFileNum)
		d.writeFile, err = os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0o600)
		if err != nil {
			return err
		}

		d.logger.Info("writeOne() opened", zap.String("name", d.name), zap.String("filename", curFileName))

		if d.writePos > 0 {
			_, err = d.writeFile.Seek(d.writePos, 0)
			if err != nil {
				_ = d.writeFile.Close()
				d.writeFile = nil
				return err
			}
		}
	}

	d.writeBuf.Reset()
	err = binary.Write(&d.writeBuf, binary.BigEndian, dataLen)
	if err != nil {
		return err
	}

	_, err = d.writeBuf.Write(data)
	if err != nil {
		return err
	}

	// only write to the file once
	_, err = d.writeFile.Write(d.writeBuf.Bytes())
	if err != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
		return err
	}

	d.writePos += totalBytes
	d.depth++

	return err
}

// sync fsyncs the current writeFile and persists metadata
func (d *diskQueue) sync() error {
	if d.writeFile != nil {
		err := d.writeFile.Sync()
		if err != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
			return err
		}
	}

	err := d.persistMetaData()
	if err != nil {
		return err
	}

	d.needSync = false
	return nil
}

// retrieveMetaData initializes state from the filesystem
func (d *diskQueue) retrieveMetaData() error {
	var f *os.File
	var err error

	fileName := d.metaDataFileName()
	f, err = os.OpenFile(fileName, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	var depth int64
	_, err = fmt.Fscanf(f, "%d\n%d,%d\n%d,%d\n",
		&depth,
		&d.readFileNum, &d.readPos,
		&d.writeFileNum, &d.writePos)
	if err != nil {
		return err
	}
	d.depth = depth
	d.nextReadFileNum = d.readFileNum
	d.nextReadPos = d.readPos

	// if the metadata was not sync'd at the last shutdown of nsqd
	// then the actual file size might actually be larger than the writePos,
	// in which case the safest thing to do is skip to the next file for
	// writes, and let the reader salvage what it can from the messages in the
	// diskqueue beyond the metadata's likely also stale readPos
	fileName = d.fileName(d.writeFileNum)
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	if d.writePos < fileSize {
		d.logger.Warn(fmt.Sprintf("metadata %s writePos %d < file size of %d, skipping to new file", fileName, d.writePos, fileSize),
			zap.String("name", d.name))
		d.writeFileNum += 1
		d.writePos = 0
		if d.writeFile != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
		}
	}

	return nil
}

// persistMetaData atomically writes state to the filesystem
func (d *diskQueue) persistMetaData() error {
	var f *os.File
	var err error

	fileName := d.metaDataFileName()
	tmpFileName := fmt.Sprintf("%s.%d.tmp", fileName, rand.Int())

	// write to tmp file
	f, err = os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f, "%d\n%d,%d\n%d,%d\n",
		d.depth,
		d.readFileNum, d.readPos,
		d.writeFileNum, d.writePos)
	if err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Sync()
	_ = f.Close()

	// atomically rename
	return os.Rename(tmpFileName, fileName)
}

func (d *diskQueue) metaDataFileName() string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.meta.dat"), d.name)
}

func (d *diskQueue) fileName(fileNum int64) string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.%06d.dat"), d.name, fileNum)
}

func (d *diskQueue) checkTailCorruption(depth int64) {
	if d.readFileNum < d.writeFileNum || d.readPos < d.writePos {
		return
	}

	// we've reached the end of the diskqueue
	// if depth isn't 0 something went wrong
	if depth != 0 {
		if depth < 0 {
			d.logger.Error(fmt.Sprintf("negative depth at tail (%d), metadata corruption, resetting 0...", depth),
				zap.String("name", d.name))
		} else {
			d.logger.Error(
				fmt.Sprintf("positive depth at tail (%d), data loss, resetting 0...", depth),
				zap.String("name", d.name),
			)
		}
		// force set depth 0
		d.depth = 0
		d.needSync = true
	}

	if d.readFileNum != d.writeFileNum || d.readPos != d.writePos {
		if d.readFileNum > d.writeFileNum {
			d.logger.Error(fmt.Sprintf("readFileNum > writeFileNum (%d > %d), corruption, skipping to next writeFileNum and resetting 0...", d.readFileNum, d.writeFileNum),
				zap.String("name", d.name))
		}

		if d.readPos > d.writePos {
			d.logger.Error(fmt.Sprintf("readPos > writePos (%d > %d), corruption, skipping to next writeFileNum and resetting 0...", d.readPos, d.writePos),
				zap.String("name", d.name))
		}

		if err := d.skipToNextRWFile(); err != nil {
			d.logger.Error("error skipping to next file", zap.Error(err))
		}
		d.needSync = true
	}
}

func (d *diskQueue) moveForward() {
	oldReadFileNum := d.readFileNum
	d.readFileNum = d.nextReadFileNum
	d.readPos = d.nextReadPos
	d.depth--

	// see if we need to clean up the old file
	if oldReadFileNum != d.nextReadFileNum {
		// sync every time we start reading from a new file
		d.needSync = true

		fn := d.fileName(oldReadFileNum)
		err := os.Remove(fn)
		if err != nil {
			d.logger.Error(" failed to Remove", zap.String("name", d.name), zap.String("filename", fn), zap.Error(err))
		}
	}

	d.checkTailCorruption(d.depth)
}

func (d *diskQueue) handleReadError(err error) {
	// jump to the next read file and rename the current (bad) file
	if d.readFileNum == d.writeFileNum {
		// if you can't properly read from the current write file it's safe to
		// assume that something is fucked and we should skip the current file too
		if d.writeFile != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
		}
		d.writeFileNum++
		d.writePos = 0
	}

	if err != io.EOF {
		badFn := d.fileName(d.readFileNum)
		badRenameFn := badFn + ".bad"

		d.logger.Warn("jump to next file and saving bad file as %s",
			zap.String("name", d.name), zap.String("badRenameFilename", badRenameFn))

		err := os.Rename(badFn, badRenameFn)
		if err != nil {
			d.logger.Error(
				"failed to rename bad diskqueue file",
				zap.String("name", d.name), zap.String("badFilename", badFn), zap.String("badRenameFilename", badRenameFn),
			)
		}
	}

	d.readFileNum++
	d.readPos = 0
	d.nextReadFileNum = d.readFileNum
	d.nextReadPos = 0

	// significant state change, schedule a sync on the next iteration
	d.needSync = true

	d.checkTailCorruption(d.depth)
}

// ioLoop provides the backend for exposing a go channel (via ReadChan())
// in support of multiple concurrent queue consumers
//
// it works by looping and branching based on whether or not the queue has data
// to read and blocking until data is either read or written over the appropriate
// go channels
//
// conveniently this also means that we're asynchronously reading from the filesystem
func (d *diskQueue) ioLoop() {
	var dataRead []byte
	var err error
	var count int64
	var r chan []byte
	var p chan []byte

	syncTicker := time.NewTicker(d.syncTimeout)

	for {
		// dont sync all the time :)
		if count == d.syncEvery {
			d.needSync = true
		}

		if d.needSync {
			err = d.sync()
			if err != nil {
				d.logger.Error(" failed to sync - %s", zap.String("name", d.name), zap.Error(err))
			}
			count = 0
		}

		if d.readFileNum < d.writeFileNum || (d.readFileNum == d.writeFileNum && d.readPos < d.writePos) {
			if d.nextReadPos == d.readPos {
				dataRead, err = d.readOne()
				if err != nil {
					d.logger.Error(fmt.Sprintf(" reading at %d of %s", d.readPos, d.fileName(d.readFileNum)),
						zap.String("name", d.name), zap.Error(err))
					d.handleReadError(err)
					continue
				}
			}
			r = d.readChan
			p = d.peekChan
		} else {
			r = nil
			p = nil
		}

		select {
		// the Go channel spec dictates that nil channel operations (read or write)
		// in a select are skipped, we set r to d.readChan only when there is data to read
		case p <- dataRead:
		case r <- dataRead:
			count++
			// moveForward sets needSync flag if a file is removed
			d.moveForward()
		case d.depthChan <- d.depth:
		case <-d.emptyChan:
			d.emptyResponseChan <- d.deleteAllFiles()
			count = 0
		case dataWrite := <-d.writeChan:
			count++
			d.writeResponseChan <- d.writeOne(dataWrite)
		case <-syncTicker.C:
			if count == 0 {
				// avoid sync when there's no activity
				continue
			}
			d.needSync = true
		case <-d.exitChan:
			goto exit
		}
	}

exit:
	d.logger.Info("closing ... ioLoop", zap.String("name", d.name))
	syncTicker.Stop()
	d.exitSyncChan <- 1
}
