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

package pqprocessor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-diskqueue"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var (
	marshaler   = &plog.ProtoMarshaler{}
	unmarshaler = &plog.ProtoUnmarshaler{}
)

type pqprocessor struct {
	next         consumer.Logs
	queue        diskqueue.Interface
	config       *Config
	shutdownChan chan struct{}
	run          chan struct{}
	settings     processor.Settings
	wg           sync.WaitGroup
	limit        atomic.Int32
	limitEnabled bool
}

func (b *pqprocessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (b *pqprocessor) Start(_ context.Context, _ component.Host) error {
	b.shutdownChan = make(chan struct{})
	b.run = make(chan struct{}, 1)
	if err := os.MkdirAll(b.config.Folder, 0o755); err != nil {
		return err
	}

	maxSize := b.config.Bandwidth
	if maxSize == 0 {
		maxSize = 10_000_000
	}

	q := diskqueue.New(
		"processor",
		b.config.Folder,
		1024*1024*1024,
		0,
		maxSize,
		int64(1*time.Second),
		100*time.Millisecond,
		func(lvl diskqueue.LogLevel, f string, args ...interface{}) {
			switch lvl {
			case diskqueue.DEBUG:
				b.settings.Logger.Debug(fmt.Sprintf(f, args...))
			case diskqueue.INFO:
				b.settings.Logger.Info(fmt.Sprintf(f, args...))
			case diskqueue.WARN:
				b.settings.Logger.Warn(fmt.Sprintf(f, args...))
			case diskqueue.ERROR:
				b.settings.Logger.Error(fmt.Sprintf(f, args...))
			case diskqueue.FATAL:
				b.settings.Logger.Fatal(fmt.Sprintf(f, args...))
			}
		},
	)
	b.queue = q

	b.limitEnabled = b.config.Bandwidth > 0
	if b.limitEnabled {
		b.limit = atomic.Int32{}
		b.limit.Store(b.config.Bandwidth)
	}
	b.tryRun()
	b.scheduleLimit()

	b.wg.Go(b.consumeLoop)
	return nil
}

func (b *pqprocessor) Shutdown(_ context.Context) error {
	close(b.shutdownChan)
	b.wg.Wait()
	if b.queue != nil {
		return b.queue.Close()
	}
	return nil
}

func (b *pqprocessor) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	marshaled, err := marshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	if b.config.Bandwidth != 0 && int32(len(marshaled)) > b.config.Bandwidth {
		return fmt.Errorf("cannot consume object, size is larger than allowed bandwidth: %d bytes, max %d bytes", len(marshaled), b.config.Bandwidth)
	}
	err = b.queue.Put(marshaled)
	if err != nil {
		return err
	}

	return nil
}

func (b *pqprocessor) tryRun() {
	select {
	case b.run <- struct{}{}:
	default:
	}
}

func (b *pqprocessor) scheduleLimit() {
	b.wg.Go(func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-b.shutdownChan:
				return
			case <-ticker.C:
				// reset limit
				if b.limitEnabled {
					b.limit.Store(b.config.Bandwidth)
				}
				b.tryRun()
			}
		}
	})
}

func (b *pqprocessor) consumeLoop() {
	for {
		select {
		case <-b.shutdownChan:
			return
		case <-b.run:
			b.innerLoop()
		}
	}
}

func (b *pqprocessor) innerLoop() {
InnerLoop:
	for {
		select {
		case <-b.shutdownChan:
			return
		case newMessage := <-b.queue.PeekChan():
			msg, err := unmarshaler.UnmarshalLogs(newMessage)
			if err != nil {
				b.settings.Logger.Error("error unmarshaling logs", zap.Error(err))
				<-b.queue.ReadChan()
				continue InnerLoop
			}

			err = b.next.ConsumeLogs(context.Background(), msg)
			if err != nil {
				b.settings.Logger.Error("error consuming logs", zap.Error(err))
			} else {
				<-b.queue.ReadChan()
				if b.limitEnabled {
					b.limit.Add(-int32(len(newMessage)))
				}
			}
			if b.limitEnabled && b.limit.Load()-int32(len(newMessage)) < 0 {
				break InnerLoop
			}
		}
	}
}
