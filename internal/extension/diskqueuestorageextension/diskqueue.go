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
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type message struct {
	consumeCallback func()
	payload         func() []byte
}

type metadata struct {
	Segments    []*segment
	PeekFileNum int64
	PeekPos     int64
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
	writeResponseChan     chan error
	writeChan             chan []byte
	exitChan              chan int
	recomputePeekChan     chan struct{}
	peekRequestChan       chan struct{}
	logger                *zap.Logger
	writeFile             *os.File
	peekFile              *os.File
	metadataFile          *os.File
	peekChan              chan message
	dataPath              string
	name                  string
	metadataLock          sync.Mutex
	metadata              metadata
	maxBytesPerFile       int64
	syncTimeout           time.Duration
	syncEvery             int64
	exitFlag              atomic.Bool
	exitWG                sync.WaitGroup
	metadataTruncateEvery int
	metadataWrites        int
	opCount               atomic.Int64
	callbackChan          chan callback
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
	d.exitWG.Go(d.writeLoop)
	d.exitWG.Go(d.callbackLoop)
	d.exitWG.Go(d.ioLoop)
	return &d
}

var bufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

func (d *diskQueue) put(data []byte) error {
	if d.exitFlag.Load() {
		return errors.New("exiting")
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(buf)
	defer gzipWriterPool.Put(zw)
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
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
		curFileName := d.fileName(d.metadata.PeekFileNum)
		d.peekFile, err = os.OpenFile(curFileName, os.O_RDONLY, 0o600)
		if err != nil {
			return nil, err
		}
		d.logger.Debug("peekOne() opened", zap.String("name", d.name), zap.String("filename", curFileName))
	}

	var msgSize int64
	for _, s := range d.metadata.Segments {
		if s.FileNum == d.metadata.PeekFileNum && s.Pos == d.metadata.PeekPos {
			msgSize = s.MessageLen
			break
		}
	}

	readBuf := make([]byte, msgSize)
	_, err = d.peekFile.ReadAt(readBuf, d.metadata.PeekPos)
	if err != nil {
		_ = d.peekFile.Close()
		d.peekFile = nil
		return nil, err
	}

	return readBuf, nil
}

func (d *diskQueue) lastSegment() *segment {
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
	d.metadataLock.Lock()
	defer d.metadataLock.Unlock()
	var err error

	dataLen := int64(len(data))
	totalBytes := dataLen

	lastSegment := d.lastSegment()
	writeFileNum := lastSegment.FileNum
	writePos := lastSegment.Pos + lastSegment.MessageLen

	// will not wrap-around if maxBytesPerFile + maxMsgSize < Int64Max
	if writePos > 0 && writePos+totalBytes > d.maxBytesPerFile {
		writeFileNum++
		writePos = 0

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
		curFileName := d.fileName(writeFileNum)
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
	_, err = d.writeFile.Write(data)
	if err != nil {
		_ = d.writeFile.Close()
		d.writeFile = nil
		return err
	}

	d.metadata.Segments = append(d.metadata.Segments, &segment{
		FileNum:    writeFileNum,
		Pos:        writePos,
		MessageLen: totalBytes,
		Consumed:   false,
	})

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

// retrieveMetaData initializes state from the filesystem
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

func (d *diskQueue) peekForward() {
	d.metadataLock.Lock()
	defer d.metadataLock.Unlock()
	var lastSegment *segment
	for _, s := range d.metadata.Segments {
		if !s.Consumed && (s.FileNum > d.metadata.PeekFileNum || (s.FileNum == d.metadata.PeekFileNum && s.Pos > d.metadata.PeekPos)) {
			d.metadata.PeekFileNum = s.FileNum
			d.metadata.PeekPos = s.Pos
			return
		}
		lastSegment = s
	}
	// place at the tip.
	if lastSegment != nil {
		d.metadata.PeekFileNum = lastSegment.FileNum
		d.metadata.PeekPos = lastSegment.Pos + lastSegment.MessageLen
	}
}

func (d *diskQueue) moveForward(fileNum, pos, messageLen int64) {
	d.metadataLock.Lock()
	defer d.metadataLock.Unlock()
	consumedFiles := make(map[int64]bool, len(d.metadata.Segments))
	for _, s := range d.metadata.Segments {
		// consume the segment
		if s.FileNum != d.metadata.PeekFileNum && s.FileNum == fileNum && s.Pos == pos && s.MessageLen == messageLen {
			s.Consumed = true
		}
		if _, ok := consumedFiles[s.FileNum]; !ok {
			consumedFiles[s.FileNum] = s.Consumed
		} else {
			consumedFiles[s.FileNum] = consumedFiles[s.FileNum] && s.Consumed
		}
	}

	if len(consumedFiles) == 0 {
		return
	}

	for fileNum, consumed := range consumedFiles {
		if consumed {
			f := d.fileName(fileNum)
			err := os.Remove(f)
			if err != nil && !os.IsNotExist(err) {
				d.logger.Error(" failed to Remove", zap.String("name", d.name), zap.String("filename", f), zap.Error(err))
			}
		}

	}

	compactedSegments := make([]*segment, 0, len(d.metadata.Segments))
	for _, s := range d.metadata.Segments {
		if !consumedFiles[s.FileNum] {
			compactedSegments = append(compactedSegments, s)
		}
	}
	d.metadata.Segments = compactedSegments
}

func (d *diskQueue) writeLoop() {
	syncTicker := time.NewTicker(d.syncTimeout)
	defer syncTicker.Stop()
	for {
		select {
		case dataWrite := <-d.writeChan:
			d.opCount.Add(1)
			data := d.write(dataWrite)
			// if peek had caught up to the tip, recompute it now that a new entry is available.
			select {
			case d.peekRequestChan <- struct{}{}:
			default:
			}
			d.writeResponseChan <- data
		case <-syncTicker.C:
			if d.opCount.Load() == 0 {
				// avoid sync when there's no activity
				continue
			}
			if d.opCount.Load() == d.syncEvery {
				err := d.sync()
				if err != nil {
					d.logger.Error("failed to sync", zap.String("name", d.name), zap.Error(err))
				}
				d.opCount.Store(0)
			}
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) callbackLoop() {
	for {
		select {
		case c := <-d.callbackChan:
			d.moveForward(c.fileNum, c.pos, c.len)
			d.opCount.Add(1)
		case <-d.exitChan:
			return
		}
	}
}

func (d *diskQueue) ioLoop() {
	for {
		select {
		case <-d.peekRequestChan:
			d.metadataLock.Lock()
			lastSegment := d.lastSegment()
			var msg message
			if d.metadata.PeekFileNum < lastSegment.FileNum || (d.metadata.PeekFileNum == lastSegment.FileNum && d.metadata.PeekPos < lastSegment.Pos+lastSegment.MessageLen) {
				peekDataRead, err := d.peekData()
				messagePeekFileNum := d.metadata.PeekFileNum
				messagePeekPos := d.metadata.PeekPos
				d.metadataLock.Unlock()
				if err != nil {
					d.logger.Error(fmt.Sprintf("failed peeking at %d of %s", d.metadata.PeekPos, d.fileName(d.metadata.PeekFileNum)),
						zap.String("name", d.name), zap.Error(err))
					panic("foo")
				} else {
					msg = message{
						payload: func() []byte {
							r, err := gzip.NewReader(bytes.NewReader(peekDataRead))
							if err != nil {
								d.logger.Error("error creating reader", zap.Error(err))
								return []byte{}
							}
							defer func() {
								_ = r.Close()
							}()
							decompressed, err := io.ReadAll(r)
							if err != nil {
								d.logger.Error("error decompressing entry", zap.Error(err))
							}
							return decompressed
						},
						consumeCallback: func() {
							d.callbackChan <- callback{
								pos:     messagePeekPos,
								len:     int64(len(peekDataRead)),
								fileNum: messagePeekFileNum,
							}
						},
					}
				}
				select {
				case d.peekChan <- msg:
					d.peekForward()
					d.opCount.Add(1)
				case <-d.exitChan:
					return
				}
			} else {
				d.metadataLock.Unlock()
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
	d.metadataLock.Lock()
	err := e.Encode(d.metadata)
	d.metadataLock.Unlock()
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

func (d *diskQueue) fileName(fileNum int64) string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.%06d.dat"), d.name, fileNum)
}
