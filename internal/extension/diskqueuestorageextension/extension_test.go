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
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
)

var _ component.Host = (*hostWithExtensions)(nil)

type hostWithExtensions struct {
	extensions map[component.ID]component.Component
}

func (h hostWithExtensions) GetExtensions() map[component.ID]component.Component {
	return h.extensions
}

func TestExtensionAsPersistentQueue(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	recf := otlpreceiver.NewFactory()
	rCfg := recf.CreateDefaultConfig().(*otlpreceiver.Config)
	listenerForFreePort, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	require.NoError(t, listenerForFreePort.Close())
	rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint = listenerForFreePort.Addr().String()
	sink := &consumertest.LogsSink{}
	receiverSettings := receivertest.NewNopSettings(component.MustNewType("otlp"))
	receiverSettings.Logger = logger
	reclogs, err := recf.CreateLogs(t.Context(), receiverSettings, rCfg, sink)
	require.NoError(t, err)
	require.NoError(t, reclogs.Start(t.Context(), componenttest.NewNopHost()))
	f := otlpexporter.NewFactory()
	cfg := f.CreateDefaultConfig().(*otlpexporter.Config)
	extID := component.MustNewIDWithName("disk_queue_storage", "my")
	cfg.QueueConfig.GetOrInsertDefault().StorageID = &extID
	cfg.QueueConfig.GetOrInsertDefault().WaitForResult = true
	cfg.ClientConfig.Endpoint = rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint
	cfg.ClientConfig.TLS.Insecure = true
	exporterSettings := exportertest.NewNopSettings(component.MustNewType("otlp"))
	exporterSettings.Logger = logger
	l, err := f.CreateLogs(t.Context(), exporterSettings, cfg)
	require.NoError(t, err)
	extConfig := createDefaultConfig().(*Config)
	extConfig.Path = t.TempDir()
	extensionSettings := extensiontest.NewNopSettings(component.MustNewType("disk_queue_storage"))
	extensionSettings.Logger = logger
	ext := newDiskQueueStorageExtension(extensionSettings, extConfig)
	require.NoError(t, l.Start(t.Context(), hostWithExtensions{
		extensions: map[component.ID]component.Component{
			extID: ext,
		},
	}))
	for i := range 10 {
		logs := plog.NewLogs()
		logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr(fmt.Sprintf("hello world %d", i))
		require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		lrs := map[string]struct{}{}
		var logs []string
		for _, l := range sink.AllLogs() {
			log := l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Body().Str()
			lrs[log] = struct{}{}
			logs = append(logs, log)
		}
		require.Len(tt, lrs, 10, strings.Join(logs, ","))
	}, 2*time.Second, 100*time.Millisecond)
	require.NoError(t, l.Shutdown(t.Context()))
	require.NoError(t, reclogs.Shutdown(t.Context()))
	require.NoError(t, ext.Shutdown(t.Context()))
}

func TestExtensionAsPersistentQueueWithWorkers(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	recf := otlpreceiver.NewFactory()
	rCfg := recf.CreateDefaultConfig().(*otlpreceiver.Config)
	listenerForFreePort, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	require.NoError(t, listenerForFreePort.Close())
	rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint = listenerForFreePort.Addr().String()
	sink := &consumertest.LogsSink{}
	receiverSettings := receivertest.NewNopSettings(component.MustNewType("otlp"))
	receiverSettings.Logger = logger
	reclogs, err := recf.CreateLogs(t.Context(), receiverSettings, rCfg, sink)
	require.NoError(t, err)
	require.NoError(t, reclogs.Start(t.Context(), componenttest.NewNopHost()))
	f := otlpexporter.NewFactory()
	cfg := f.CreateDefaultConfig().(*otlpexporter.Config)
	extID := component.MustNewIDWithName("disk_queue_storage", "my")
	cfg.QueueConfig.GetOrInsertDefault().StorageID = &extID
	cfg.QueueConfig.GetOrInsertDefault().WaitForResult = true
	cfg.ClientConfig.Endpoint = rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint
	cfg.ClientConfig.TLS.Insecure = true
	cfg.QueueConfig.GetOrInsertDefault().NumConsumers = 32
	exporterSettings := exportertest.NewNopSettings(component.MustNewType("otlp"))
	exporterSettings.Logger = logger
	l, err := f.CreateLogs(t.Context(), exporterSettings, cfg)
	require.NoError(t, err)
	extConfig := createDefaultConfig().(*Config)
	extConfig.Path = t.TempDir()
	extensionSettings := extensiontest.NewNopSettings(component.MustNewType("disk_queue_storage"))
	extensionSettings.Logger = logger
	ext := newDiskQueueStorageExtension(extensionSettings, extConfig)
	require.NoError(t, l.Start(t.Context(), hostWithExtensions{
		extensions: map[component.ID]component.Component{
			extID: ext,
		},
	}))
	for i := range 10 {
		logs := plog.NewLogs()
		logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr(fmt.Sprintf("hello world %d", i))
		require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		lrs := map[string]struct{}{}
		for _, l := range sink.AllLogs() {
			log := l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Body().Str()
			lrs[log] = struct{}{}
		}
		require.Len(tt, lrs, 10)
	}, 1*time.Second, 100*time.Millisecond)
	require.NoError(t, l.Shutdown(t.Context()))
	require.NoError(t, reclogs.Shutdown(t.Context()))
	require.NoError(t, ext.Shutdown(t.Context()))
}

type statefulLogSink struct {
	sink         *consumertest.LogsSink
	allowTraffic atomic.Bool
}

func (s *statefulLogSink) Capabilities() consumer.Capabilities {
	return s.sink.Capabilities()
}

func (s *statefulLogSink) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	if s.allowTraffic.Load() {
		return s.sink.ConsumeLogs(ctx, logs)
	}
	return consumererror.NewRetryableError(errors.New("no"))
}

func BenchmarkExtensionAsPersistentQueueWithWorkers(b *testing.B) {
	for _, extType := range []string{"diskqueue", "bbolt"} {
		for _, traffic := range []string{"ok", "draining"} {
			for _, volume := range []int{100, 200, 300, 400, 500, 1000, 5000, 10000} {
				b.Run(fmt.Sprintf("bench-%s-%s-%d", extType, traffic, volume), func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						logger := zap.NewNop()
						recf := otlpreceiver.NewFactory()
						rCfg := recf.CreateDefaultConfig().(*otlpreceiver.Config)
						listenerForFreePort, err := net.Listen("tcp", "localhost:0")
						require.NoError(b, err)
						require.NoError(b, listenerForFreePort.Close())
						rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint = listenerForFreePort.Addr().String()
						rCfg.Protocols.GRPC.GetOrInsertDefault().MaxConcurrentStreams = 1024
						var logConsumer consumer.Logs
						sink := &consumertest.LogsSink{}
						if traffic == "ok" {
							logConsumer = sink
						} else {
							logConsumer = &statefulLogSink{
								sink: sink,
							}
							logConsumer.(*statefulLogSink).allowTraffic.Store(false)
						}
						receiverSettings := receivertest.NewNopSettings(component.MustNewType("otlp"))
						receiverSettings.Logger = logger
						reclogs, err := recf.CreateLogs(b.Context(), receiverSettings, rCfg, logConsumer)
						require.NoError(b, err)
						require.NoError(b, reclogs.Start(b.Context(), componenttest.NewNopHost()))
						f := otlpexporter.NewFactory()
						cfg := f.CreateDefaultConfig().(*otlpexporter.Config)
						cfg.QueueConfig.GetOrInsertDefault().NumConsumers = 32
						cfg.ClientConfig.TLS.Insecure = true
						cfg.QueueConfig.GetOrInsertDefault().WaitForResult = true
						cfg.ClientConfig.Endpoint = rCfg.Protocols.GRPC.GetOrInsertDefault().NetAddr.Endpoint
						exporterSettings := exportertest.NewNopSettings(component.MustNewType("otlp"))
						exporterSettings.Logger = logger
						var l exporter.Logs
						var ext extension.Extension
						switch extType {
						case "diskqueue":
							extID := component.MustNewIDWithName("disk_queue_storage", "my")
							cfg.QueueConfig.GetOrInsertDefault().StorageID = &extID
							l, err = f.CreateLogs(b.Context(), exporterSettings, cfg)
							require.NoError(b, err)
							extConfig := createDefaultConfig().(*Config)
							extConfig.Path = b.TempDir()
							extensionSettings := extensiontest.NewNopSettings(component.MustNewType("disk_queue_storage"))
							extensionSettings.Logger = logger
							ext = newDiskQueueStorageExtension(extensionSettings, extConfig)
							require.NoError(b, l.Start(b.Context(), hostWithExtensions{
								extensions: map[component.ID]component.Component{
									extID: ext,
								},
							}))
						case "bbolt":
							extID := component.MustNewIDWithName("file_storage", "my")
							cfg.QueueConfig.GetOrInsertDefault().StorageID = &extID
							l, err = f.CreateLogs(b.Context(), exporterSettings, cfg)
							require.NoError(b, err)
							extF := filestorage.NewFactory()
							extConfig := extF.CreateDefaultConfig().(*filestorage.Config)
							extConfig.Directory = b.TempDir()
							extConfig.FSync = true
							extensionSettings := extensiontest.NewNopSettings(component.MustNewType("file_storage"))
							extensionSettings.Logger = logger
							ext, err = extF.Create(b.Context(), extensionSettings, extConfig)
							require.NoError(b, err)
							require.NoError(b, l.Start(b.Context(), hostWithExtensions{
								extensions: map[component.ID]component.Component{
									extID: ext,
								},
							}))
						}

						for i := range volume {
							logs := plog.NewLogs()
							logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr(fmt.Sprintf("hello world %d", i))
							require.NoError(b, l.ConsumeLogs(b.Context(), logs))
						}
						if traffic == "draining" {
							logConsumer.(*statefulLogSink).allowTraffic.Store(true)
						}
						assert.EventuallyWithT(b, func(tt *assert.CollectT) {
							lrs := map[string]struct{}{}
							for _, l := range sink.AllLogs() {
								log := l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Body().Str()
								lrs[log] = struct{}{}
							}
							require.Len(tt, lrs, volume)
						}, 60*time.Second, 100*time.Millisecond)

						require.NoError(b, l.Shutdown(b.Context()))
						require.NoError(b, reclogs.Shutdown(b.Context()))
						require.NoError(b, ext.Shutdown(b.Context()))
					}
				})
			}
		}
	}
}
