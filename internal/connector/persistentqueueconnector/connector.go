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

package persistentqueueconnector

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-diskqueue"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/connector/persistentqueueconnector/internal"
)

var (
	metricsMarshaler    = &pmetric.ProtoMarshaler{}
	metricsUnmarshaler  = &pmetric.ProtoUnmarshaler{}
	logsMarshaler       = &plog.ProtoMarshaler{}
	logsUnmarshaler     = &plog.ProtoUnmarshaler{}
	tracesMarshaler     = &ptrace.ProtoMarshaler{}
	tracesUnmarshaler   = &ptrace.ProtoUnmarshaler{}
	profilesMarshaler   = &pprofile.ProtoMarshaler{}
	profilesUnmarshaler = &pprofile.ProtoUnmarshaler{}

	versionByte = byte(0)
	logByte     = byte(0)
	metricByte  = byte(1)
	traceByte   = byte(2)
	profileByte = byte(3)

	bufPool = sync.Pool{New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	}}
)

type persistentqueue struct {
	nextLogs     consumer.Logs
	nextMetrics  consumer.Metrics
	nextTraces   consumer.Traces
	nextProfiles xconsumer.Profiles
	queue        diskqueue.Interface
	config       *Config
	shutdownChan chan struct{}
	run          chan struct{}
	settings     connector.Settings
	wg           sync.WaitGroup
	limit        atomic.Int32
	limitEnabled bool
}

func (b *persistentqueue) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (b *persistentqueue) Start(_ context.Context, _ component.Host) error {
	b.shutdownChan = make(chan struct{})
	b.run = make(chan struct{}, 1)
	if err := os.MkdirAll(b.config.Path, 0o755); err != nil {
		return err
	}

	maxSize := b.config.ThroughputLimit
	if maxSize == 0 {
		maxSize = 10_000_000
	}

	q := internal.New(
		"processor",
		b.config.Path,
		1024*1024*1024,
		0,
		maxSize,
		int64(1*time.Second),
		100*time.Millisecond,
		b.settings.Logger,
	)
	b.queue = q

	b.limitEnabled = b.config.ThroughputLimit > 0
	if b.limitEnabled {
		b.limit = atomic.Int32{}
		b.limit.Store(b.config.ThroughputLimit)
	}
	b.tryRun()
	b.scheduleLimit()

	b.wg.Go(b.consumeLoop)
	return nil
}

func (b *persistentqueue) Shutdown(_ context.Context) error {
	close(b.shutdownChan)
	b.wg.Wait()
	if b.queue != nil {
		return b.queue.Close()
	}
	return nil
}

func (b *persistentqueue) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	marshaled, err := logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	buf.WriteByte(versionByte)
	buf.WriteByte(logByte)
	buf.Write(marshaled)
	int32len := int32(len(marshaled)) //nolint:gosec // disable G115
	if b.config.ThroughputLimit != 0 && int32len > b.config.ThroughputLimit || int32len < 0 {
		return fmt.Errorf("cannot consume object, size is larger than allowed bandwidth: %d bytes, max %d bytes", len(marshaled), b.config.ThroughputLimit)
	}
	err = b.queue.Put(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (b *persistentqueue) ConsumeTraces(_ context.Context, ld ptrace.Traces) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	marshaled, err := tracesMarshaler.MarshalTraces(ld)
	if err != nil {
		return err
	}
	buf.WriteByte(versionByte)
	buf.WriteByte(traceByte)
	buf.Write(marshaled)
	int32len := int32(len(marshaled)) //nolint:gosec // disable G115
	if b.config.ThroughputLimit != 0 && (int32len > b.config.ThroughputLimit || int32len < 0) {
		return fmt.Errorf("cannot consume object, size is larger than allowed bandwidth: %d bytes, max %d bytes", len(marshaled), b.config.ThroughputLimit)
	}
	err = b.queue.Put(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (b *persistentqueue) ConsumeProfiles(_ context.Context, pd pprofile.Profiles) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	marshaled, err := profilesMarshaler.MarshalProfiles(pd)
	if err != nil {
		return err
	}
	buf.WriteByte(versionByte)
	buf.WriteByte(profileByte)
	buf.Write(marshaled)
	int32len := int32(len(marshaled)) //nolint:gosec // disable G115
	if b.config.ThroughputLimit != 0 && (int32len > b.config.ThroughputLimit || int32len < 0) {
		return fmt.Errorf("cannot consume object, size is larger than allowed bandwidth: %d bytes, max %d bytes", len(marshaled), b.config.ThroughputLimit)
	}
	err = b.queue.Put(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (b *persistentqueue) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	marshaled, err := metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	buf.WriteByte(versionByte)
	buf.WriteByte(metricByte)
	buf.Write(marshaled)
	int32len := int32(len(marshaled)) //nolint:gosec // disable G115
	if b.config.ThroughputLimit != 0 && (int32len > b.config.ThroughputLimit || int32len < 0) {
		return fmt.Errorf("cannot consume object, size is larger than allowed bandwidth: %d bytes, max %d bytes", len(marshaled), b.config.ThroughputLimit)
	}
	err = b.queue.Put(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (b *persistentqueue) tryRun() {
	select {
	case b.run <- struct{}{}:
	default:
	}
}

func (b *persistentqueue) scheduleLimit() {
	b.wg.Go(func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-b.shutdownChan:
				return
			case <-ticker.C:
				// reset limit
				if b.limitEnabled {
					b.limit.Store(b.config.ThroughputLimit)
				}
				b.tryRun()
			}
		}
	})
}

func (b *persistentqueue) consumeLoop() {
	for {
		select {
		case <-b.shutdownChan:
			return
		case <-b.run:
			b.innerLoop()
		}
	}
}

func (b *persistentqueue) innerLoop() {
InnerLoop:
	for {
		select {
		case <-b.shutdownChan:
			return
		case newMessage := <-b.queue.PeekChan():
			var err error
			var logs plog.Logs
			var traces ptrace.Traces
			var metrics pmetric.Metrics
			var profiles pprofile.Profiles
			dataType := ""
			switch newMessage[1] {
			case logByte:
				logs, err = logsUnmarshaler.UnmarshalLogs(newMessage[2:])
				dataType = "logs"
			case traceByte:
				traces, err = tracesUnmarshaler.UnmarshalTraces(newMessage[2:])
				dataType = "logs"
			case metricByte:
				metrics, err = metricsUnmarshaler.UnmarshalMetrics(newMessage[2:])
				dataType = "logs"
			case profileByte:
				profiles, err = profilesUnmarshaler.UnmarshalProfiles(newMessage[2:])
				dataType = "logs"
			}
			if err != nil {
				b.settings.Logger.Error("error unmarshaling "+dataType, zap.Error(err))
				<-b.queue.ReadChan()
				continue InnerLoop
			}
			switch newMessage[1] {
			case logByte:
				err = b.nextLogs.ConsumeLogs(context.Background(), logs)
			case traceByte:
				err = b.nextTraces.ConsumeTraces(context.Background(), traces)
			case metricByte:
				err = b.nextMetrics.ConsumeMetrics(context.Background(), metrics)
			case profileByte:
				err = b.nextProfiles.ConsumeProfiles(context.Background(), profiles)
			}
			if err != nil {
				b.settings.Logger.Error("error consuming "+dataType, zap.Error(err))
			} else {
				<-b.queue.ReadChan()
				if b.limitEnabled {
					b.limit.Add(-int32(len(newMessage))) //nolint:gosec // disable G115
				}
			}
			if b.limitEnabled && b.limit.Load()-int32(len(newMessage)) < 0 { //nolint:gosec // disable G115
				break InnerLoop
			}
		}
	}
}
