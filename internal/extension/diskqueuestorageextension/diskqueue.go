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

package diskqueuestorageextension

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension/internal/compression"
)

var bufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

var gzipWriterPool = compression.NewGZIPWriterPool()

var gzipReaderPool = compression.NewGZIPReaderPool()

type message struct {
	consumeCallback func()
	payload         []byte
}

type metadata struct {
	FileNum int64
	Pos     int64
}

type callback struct {
	pos     int64
	fileNum int64
}

// diskQueue implements a filesystem backed FIFO queue
type diskQueue struct {
	writeFile             *os.File
	writeChan             chan []byte
	exitChan              chan int
	peekRequestChan       chan struct{}
	waitForWriteChan      chan struct{}
	logger                *zap.Logger
	peekFile              *os.File
	metadataFile          *os.File
	peekMetadataFile      *os.File
	peekChan              chan message
	writeResponseChan     chan error
	callbackChan          chan callback
	dataPath              string
	name                  string
	metadata              metadata
	peekMetadata          metadata
	exitWG                sync.WaitGroup
	maxBytesPerFile       int64
	syncTimeout           time.Duration
	syncEvery             int64
	metadataTruncateEvery int
	metadataWrites        int
	peekMetadataWrites    int
	exitFlag              atomic.Bool
}

// newQueue instantiates an instance of diskQueue, retrieving metadata
// from the filesystem and starting the read ahead goroutine
func newQueue(name, dataPath string, maxBytesPerFile int64,
	syncEvery int64, syncTimeout time.Duration, logger *zap.Logger,
) (*diskQueue, error) {
	d := diskQueue{
		name:                  name,
		dataPath:              dataPath,
		maxBytesPerFile:       maxBytesPerFile,
		peekChan:              make(chan message),
		writeChan:             make(chan []byte),
		writeResponseChan:     make(chan error),
		exitChan:              make(chan int),
		callbackChan:          make(chan callback),
		peekRequestChan:       make(chan struct{}),
		waitForWriteChan:      make(chan struct{}),
		syncEvery:             syncEvery,
		syncTimeout:           syncTimeout,
		logger:                logger,
		metadataTruncateEvery: 1000,
	}

	// no need to lock here, nothing else could possibly be touching this instance
	m, err := d.retrieveMetaData(d.metaDataFilePath())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	d.metadata = *m
	m, err = d.retrieveMetaData(d.peekMetaDataFilePath())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	d.peekMetadata = *m
	d.exitWG.Go(d.readLoop)
	d.exitWG.Go(d.writeLoop)
	return &d, nil
}

func (d *diskQueue) put(data []byte) error {
	if d.exitFlag.Load() {
		return errors.New("exiting")
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	if _, err := gzipWriterPool.Write(buf, data); err != nil {
		return err
	}
	d.writeChan <- buf.Bytes()
	return <-d.writeResponseChan
}

func (d *diskQueue) close() error {
	d.logger.Debug("closing", zap.String("name", d.name))
	close(d.exitChan)

	d.exitFlag.Store(true)
	d.exitWG.Wait()

	close(d.peekChan)

	_ = d.sync()
	_ = d.syncPeek()

	if d.writeFile != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
	}

	if d.peekFile != nil {
		_ = d.peekFile.Close()
		d.peekFile = nil
	}

	if d.metadataFile != nil {
		_ = d.metadataFile.Close()
		d.metadataFile = nil
	}

	if d.peekMetadataFile != nil {
		_ = d.peekMetadataFile.Close()
		d.peekMetadataFile = nil
	}

	return nil
}

func (d *diskQueue) peek() chan message {
	d.peekRequestChan <- struct{}{}
	return d.peekChan
}

func (d *diskQueue) peekData() ([]byte, error) {
	var err error
	if d.peekFile == nil {
		curFileName := d.fileName(d.peekMetadata.FileNum)
		d.peekFile, err = os.OpenFile(curFileName, os.O_RDONLY, 0o600)
		if err != nil {
			return nil, err
		}
		d.logger.Debug("peekData() opened", zap.String("name", d.name), zap.String("filename", curFileName))
	}

	readSize := make([]byte, 8)
	_, err = d.peekFile.ReadAt(readSize, d.peekMetadata.Pos)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		_ = d.peekFile.Close()
		d.peekFile = nil
		return nil, err
	}
	size := binary.BigEndian.Uint64(readSize)
	readBuf := make([]byte, size)
	_, err = d.peekFile.ReadAt(readBuf, d.peekMetadata.Pos+8)
	if err != nil {
		_ = d.peekFile.Close()
		d.peekFile = nil
		return nil, err
	}

	return readBuf, nil
}

func (d *diskQueue) write(data []byte) error {
	dataLen := int64(len(data))

	if d.writeFile == nil {
		curFileName := d.fileName(d.metadata.FileNum)
		var err error
		d.writeFile, err = os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0o600)
		if err != nil {
			return err
		}

		d.logger.Debug("writeOne() opened", zap.String("name", d.name), zap.String("filename", curFileName))

		if d.metadata.Pos > 0 {
			_, err = d.writeFile.Seek(d.metadata.Pos, 0)
			if err != nil {
				_ = d.writeFile.Close()
				d.writeFile = nil
				return err
			}
		}
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(dataLen))
	buf.Write(b)
	buf.Write(data)
	_, err := d.writeFile.Write(buf.Bytes())
	bufPool.Put(buf)
	if err != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
		return err
	}

	writeFileNum := d.metadata.FileNum
	writePos := d.metadata.Pos + dataLen + 8

	// will not wrap-around if maxBytesPerFile + maxMsgSize < Int64Max
	if writePos > 0 && writePos > d.maxBytesPerFile {
		writeFileNum++
		writePos = 0

		// sync every time we start writing to a new file
		err := d.sync()
		if err != nil {
			d.logger.Error(" failed to sync - %s", zap.String("name", d.name), zap.Error(err))
		}

		if d.writeFile != nil {
			_ = d.writeFile.Close()
			d.writeFile = nil
		}
	}
	d.metadata.Pos = writePos
	d.metadata.FileNum = writeFileNum

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

	return nil
}

func (d *diskQueue) syncPeek() error {
	fileName := d.peekMetaDataFilePath()
	if d.peekMetadataFile == nil {
		f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		d.peekMetadataFile = f
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	buf.WriteString(separator)
	buf.Write(binary.BigEndian.AppendUint64(nil, uint64(d.peekMetadata.FileNum)))
	buf.Write(binary.BigEndian.AppendUint64(nil, uint64(d.peekMetadata.Pos)))
	_, err := d.peekMetadataFile.Write(buf.Bytes())
	if err != nil {
		_ = d.peekMetadataFile.Close()
		d.peekMetadataFile = nil
		return err
	}
	err = d.peekMetadataFile.Sync()
	if err != nil {
		_ = d.peekMetadataFile.Close()
		d.peekMetadataFile = nil
		return err
	}
	d.peekMetadataWrites++

	if d.peekMetadataWrites%d.metadataTruncateEvery == 0 {
		_ = d.peekMetadataFile.Close()
		d.peekMetadataFile = nil
		d.peekMetadataWrites = 0
	}

	return nil
}

func (d *diskQueue) retrieveMetaData(fileName string) (*metadata, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0o600)
	if err != nil {
		return &metadata{}, err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Seek(-16, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	fileNum := binary.BigEndian.Uint64(buf.Bytes()[0:8])
	pos := binary.BigEndian.Uint64(buf.Bytes()[8:])
	return &metadata{
		FileNum: int64(fileNum),
		Pos:     int64(pos),
	}, nil
}

func (d *diskQueue) writeLoop() {
	syncTicker := time.NewTicker(d.syncTimeout)
	defer syncTicker.Stop()
	opCount := int64(0)
	for {
		select {
		case dataWrite := <-d.writeChan:
			opCount++
			err := d.write(dataWrite)
			d.writeResponseChan <- err
			select {
			case d.waitForWriteChan <- struct{}{}:
			default:
			}
		case <-syncTicker.C:
			if opCount == 0 {
				// avoid sync when there's no activity
				continue
			}
			if opCount == d.syncEvery {
				err := d.sync()
				if err != nil {
					d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(err))
				}

				opCount = 0
			}
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) readLoop() {
	syncTicker := time.NewTicker(d.syncTimeout)
	peekOps := int64(0)
	callbacks := map[int64]int{}
	defer syncTicker.Stop()
	var p chan struct{}
	for {
		select {
		case <-p:
			p = nil
			if d.readOne(callbacks) {
				peekOps++
			} else {
				p = d.waitForWriteChan
			}
		case <-d.peekRequestChan:
			if d.readOne(callbacks) {
				peekOps++
			} else {
				p = d.waitForWriteChan
			}
		case c := <-d.callbackChan:
			callbacks[c.fileNum]--
			if c.fileNum != d.peekMetadata.FileNum && callbacks[c.fileNum] == 0 {
				f := d.fileName(c.fileNum)
				err := os.Remove(f)
				if err != nil && !os.IsNotExist(err) {
					d.logger.Error(" failed to Remove", zap.String("name", d.name), zap.String("filename", f), zap.Error(err))
				}
			}
		case <-syncTicker.C:
			if peekOps == 0 {
				// avoid sync when there's no activity
				continue
			}
			if peekOps == d.syncEvery {
				err := d.syncPeek()
				if err != nil {
					d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(err))
				}
				peekOps = 0
			}
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) readOne(callbacks map[int64]int) bool {
	peekData, err := d.peekData()
	if err != nil {
		d.logger.Error("error peeking", zap.Error(err))
		return true
	}
	// caught to the head of the queue.
	if len(peekData) == 0 {
		return false
	}
	messagePeekFileNum := d.peekMetadata.FileNum
	messagePeekPos := d.peekMetadata.Pos

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	_, err = gzipReaderPool.Read(buf, peekData)
	if err != nil {
		d.logger.Error("error decompressing entry", zap.Error(err))
		return true
	}
	callbacks[messagePeekFileNum]++
	msg := message{
		payload: buf.Bytes(),
		consumeCallback: func() {
			d.callbackChan <- callback{
				pos:     messagePeekPos,
				fileNum: messagePeekFileNum,
			}
			bufPool.Put(buf)
		},
	}
	select {
	case d.peekChan <- msg:
		if d.peekMetadata.Pos+int64(len(peekData)+8) > d.maxBytesPerFile {
			d.peekMetadata.Pos = 0
			d.peekMetadata.FileNum++
		} else {
			d.peekMetadata.Pos = d.peekMetadata.Pos + int64(len(peekData)+8)
		}
	case <-d.exitChan:
	}
	return true
}

func (d *diskQueue) persistMetaData() error {
	fileName := d.metaDataFilePath()
	if d.metadataFile == nil {
		f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		d.metadataFile = f
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	buf.WriteString(separator)
	buf.Write(binary.BigEndian.AppendUint64(nil, uint64(d.metadata.FileNum)))
	buf.Write(binary.BigEndian.AppendUint64(nil, uint64(d.metadata.Pos)))
	_, err := d.metadataFile.Write(buf.Bytes())
	if err != nil {
		_ = d.metadataFile.Close()
		d.metadataFile = nil
		return err
	}
	err = d.metadataFile.Sync()
	if err != nil {
		_ = d.metadataFile.Close()
		d.metadataFile = nil
		return err
	}
	d.metadataWrites++

	if d.metadataWrites%d.metadataTruncateEvery == 0 {
		_ = d.metadataFile.Close()
		d.metadataFile = nil
		d.metadataWrites = 0
	}

	return nil
}

func (d *diskQueue) metaDataFilePath() string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.meta.dat"), d.name)
}

func (d *diskQueue) peekMetaDataFilePath() string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.peek.dat"), d.name)
}

func (d *diskQueue) fileName(fileNum int64) string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.%06d.dat"), d.name, fileNum)
}
