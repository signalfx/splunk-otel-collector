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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
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
	Segments []*segment
}

type peekMetadata struct {
	PeekFileNum int64
	PeekPos     int64
	PeekLen     int64
}

type segment struct {
	FileNum    int64
	Pos        int64
	MessageLen int64
	Consumed   bool
}

type callback struct {
	pos     int64
	len     int64
	fileNum int64
}

// diskQueue implements a filesystem backed FIFO queue
type diskQueue struct {
	callbackChan          chan callback
	writeChan             chan []byte
	exitChan              chan int
	recomputePeekChan     chan struct{}
	peekRequestChan       chan struct{}
	logger                *zap.Logger
	writeFile             *os.File
	peekFile              *os.File
	metadataFile          *os.File
	peekMetadataFile      *os.File
	peekChan              chan message
	writeResponseChan     chan error
	dataPath              string
	name                  string
	metadata              metadata
	lastSegment           *segment
	peekMetadata          peekMetadata
	exitWG                sync.WaitGroup
	maxBytesPerFile       int64
	syncTimeout           time.Duration
	syncEvery             int64
	metadataTruncateEvery int
	metadataWrites        int
	peekMetadataWrites    int
	metadataLock          sync.RWMutex
	exitFlag              atomic.Bool
}

// newQueue instantiates an instance of diskQueue, retrieving metadata
// from the filesystem and starting the read ahead goroutine
func newQueue(name, dataPath string, maxBytesPerFile int64,
	syncEvery int64, syncTimeout time.Duration, logger *zap.Logger,
) *diskQueue {
	d := diskQueue{
		name:                  name,
		dataPath:              dataPath,
		maxBytesPerFile:       maxBytesPerFile,
		peekChan:              make(chan message),
		writeChan:             make(chan []byte),
		writeResponseChan:     make(chan error),
		exitChan:              make(chan int),
		recomputePeekChan:     make(chan struct{}),
		callbackChan:          make(chan callback),
		peekRequestChan:       make(chan struct{}, 1),
		syncEvery:             syncEvery,
		syncTimeout:           syncTimeout,
		logger:                logger,
		metadataTruncateEvery: 1000,
	}

	// no need to lock here, nothing else could possibly be touching this instance
	err := d.retrieveMetaData()
	if err != nil && !os.IsNotExist(err) {
		d.logger.Error(" failed to retrieveMetaData", zap.String("name", d.name), zap.Error(err))
	}
	err = d.retrievePeekMetaData()
	if err != nil && !os.IsNotExist(err) {
		d.logger.Error(" failed to retrievePeekMetaData", zap.String("name", d.name), zap.Error(err))
	}
	d.lastSegment = d.computeLastSegment()
	d.exitWG.Go(d.writeLoop)
	d.exitWG.Go(d.ioLoop)
	return &d
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
	select {
	case d.peekRequestChan <- struct{}{}:
	default:
	}
	return d.peekChan
}

func (d *diskQueue) peekData() ([]byte, error) {
	var err error
	if d.peekFile == nil {
		curFileName := d.fileName(d.peekMetadata.PeekFileNum)
		d.peekFile, err = os.OpenFile(curFileName, os.O_RDONLY, 0o600)
		if err != nil {
			return nil, err
		}
		d.logger.Debug("peekData() opened", zap.String("name", d.name), zap.String("filename", curFileName))
	}

	_, err = d.peekFile.Seek(d.peekMetadata.PeekPos, 0)
	if err != nil {
		_ = d.peekFile.Close()
		d.peekFile = nil
		return nil, err
	}
	readBuf := make([]byte, d.peekMetadata.PeekLen)
	_, err = d.peekFile.ReadAt(readBuf, d.peekMetadata.PeekPos)
	if err != nil {
		_ = d.peekFile.Close()
		d.peekFile = nil
		return nil, err
	}

	return readBuf, nil
}

func (d *diskQueue) computeLastSegment() *segment {
	if len(d.metadata.Segments) == 0 {
		return &segment{
			FileNum:    0,
			Pos:        0,
			MessageLen: 0,
			Consumed:   true,
		}
	}
	return d.metadata.Segments[len(d.metadata.Segments)-1]
}

func (d *diskQueue) write(data []byte) error {
	dataLen := int64(len(data))
	totalBytes := dataLen

	lastSegment := d.lastSegment
	writeFileNum := lastSegment.FileNum
	writePos := lastSegment.Pos + lastSegment.MessageLen

	// will not wrap-around if maxBytesPerFile + maxMsgSize < Int64Max
	if writePos > 0 && writePos+totalBytes > d.maxBytesPerFile {
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

	if d.writeFile == nil {
		curFileName := d.fileName(writeFileNum)
		var err error
		d.writeFile, err = os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0o600)
		if err != nil {
			return err
		}

		d.logger.Debug("writeOne() opened", zap.String("name", d.name), zap.String("filename", curFileName))

		if writePos > 0 {
			_, err = d.writeFile.Seek(writePos, 0)
			if err != nil {
				_ = d.writeFile.Close()
				d.writeFile = nil
				return err
			}
		}
	}

	// only write to the file once
	_, err := d.writeFile.Write(data)
	if err != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
		return err
	}

	d.metadataLock.Lock()
	defer d.metadataLock.Unlock()
	newSegment := &segment{
		FileNum:    writeFileNum,
		Pos:        writePos,
		MessageLen: totalBytes,
		Consumed:   false,
	}
	d.metadata.Segments = append(d.metadata.Segments, newSegment)
	d.lastSegment = newSegment

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
	e := json.NewEncoder(buf)
	buf.WriteString(separator)
	err := e.Encode(d.peekMetadata)
	if err != nil {
		return err
	}
	_, err = d.peekMetadataFile.Write(buf.Bytes())
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

func (d *diskQueue) retrieveMetaData() error {
	var f *os.File
	var err error

	fileName := d.metaDataFilePath()
	f, err = os.OpenFile(fileName, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	lastMetadataUpdate := b[bytes.LastIndex(b, []byte(separator)):]
	d.metadataLock.Lock()
	err = json.Unmarshal(lastMetadataUpdate, &d.metadata)
	d.metadataLock.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (d *diskQueue) retrievePeekMetaData() error {
	var f *os.File
	var err error

	fileName := d.peekMetaDataFilePath()
	f, err = os.OpenFile(fileName, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	lastMetadataUpdate := b[bytes.LastIndex(b, []byte(separator)):]
	err = json.Unmarshal(lastMetadataUpdate, &d.peekMetadata)
	if err != nil {
		return err
	}

	return nil
}

func (d *diskQueue) peekForward() {
	d.metadataLock.RLock()
	defer d.metadataLock.RUnlock()
	var lastSegment *segment
	for _, s := range d.metadata.Segments {
		if !s.Consumed && (s.FileNum > d.peekMetadata.PeekFileNum || (s.FileNum == d.peekMetadata.PeekFileNum && s.Pos > d.peekMetadata.PeekPos)) {
			d.peekMetadata.PeekFileNum = s.FileNum
			d.peekMetadata.PeekPos = s.Pos
			d.peekMetadata.PeekLen = s.MessageLen
			return
		}
		lastSegment = s
	}
	// place at the tip.
	if lastSegment != nil {
		d.peekMetadata.PeekFileNum = lastSegment.FileNum
		d.peekMetadata.PeekPos = lastSegment.Pos + lastSegment.MessageLen
		// set when we write the next segment
		d.peekMetadata.PeekLen = 0
	}
}

func (d *diskQueue) moveForward(fileNum, pos, messageLen int64) {
	d.metadataLock.Lock()
	if d.peekMetadata.PeekFileNum == fileNum {
		d.metadataLock.Unlock()
		return
	}
	allFiles := map[int64]struct{}{}
	segmentIndex := 0
	for i, s := range d.metadata.Segments {
		// consume the segment
		if s.FileNum == fileNum && s.Pos == pos && s.MessageLen == messageLen {
			segmentIndex = i
		} else {
			allFiles[s.FileNum] = struct{}{}
		}
	}
	d.metadata.Segments = slices.Delete(d.metadata.Segments, segmentIndex, segmentIndex+1)
	d.metadataLock.Unlock()

	// file was not used anywhere else, now delete.
	if _, ok := allFiles[fileNum]; !ok {
		f := d.fileName(fileNum)
		err := os.Remove(f)
		if err != nil && !os.IsNotExist(err) {
			d.logger.Error(" failed to Remove", zap.String("name", d.name), zap.String("filename", f), zap.Error(err))
		}
	}
}

func (d *diskQueue) writeLoop() {
	syncTicker := time.NewTicker(d.syncTimeout)
	defer syncTicker.Stop()
	opCount := int64(0)
	for {
		select {
		case dataWrite := <-d.writeChan:
			opCount++
			data := d.write(dataWrite)
			// if peek had caught up to the tip, recompute it now that a new entry is available.
			select {
			case d.peekRequestChan <- struct{}{}:
			default:
			}
			d.writeResponseChan <- data
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
		case c := <-d.callbackChan:
			d.moveForward(c.fileNum, c.pos, c.len)
			opCount++
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) ioLoop() {
	syncTicker := time.NewTicker(d.syncTimeout)
	defer syncTicker.Stop()
	peekOps := int64(0)
	for {
		select {
		case <-d.peekRequestChan:
			d.metadataLock.RLock()
			lastSegment := *d.lastSegment
			d.metadataLock.RUnlock()
			var msg message
			if !(d.peekMetadata.PeekFileNum < lastSegment.FileNum || (d.peekMetadata.PeekFileNum == lastSegment.FileNum && d.peekMetadata.PeekPos < lastSegment.Pos+lastSegment.MessageLen)) {
				continue
			}
			if d.peekMetadata.PeekLen == 0 {
				d.peekMetadata.PeekLen = lastSegment.MessageLen
			}
			peekData, err := d.peekData()
			if err != nil {
				d.logger.Error("error peeking", zap.Error(err))
				continue
			}
			messageLen := int64(len(peekData))
			messagePeekFileNum := d.peekMetadata.PeekFileNum
			messagePeekPos := d.peekMetadata.PeekPos

			buf := bufPool.Get().(*bytes.Buffer)
			buf.Reset()
			_, err = gzipReaderPool.Read(buf, peekData)
			if err != nil {
				d.logger.Error("error decompressing entry", zap.Error(err))
				continue
			}
			msg = message{
				payload: buf.Bytes(),
				consumeCallback: func() {
					d.callbackChan <- callback{
						pos:     messagePeekPos,
						len:     messageLen,
						fileNum: messagePeekFileNum,
					}
					bufPool.Put(buf)
				},
			}
			select {
			case d.peekChan <- msg:
				d.peekForward()
				peekOps++
			case <-d.exitChan:
				return
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
	e := json.NewEncoder(buf)
	buf.WriteString(separator)
	d.metadataLock.RLock()
	err := e.Encode(d.metadata)
	d.metadataLock.RUnlock()
	if err != nil {
		return err
	}
	_, err = d.metadataFile.Write(buf.Bytes())
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
