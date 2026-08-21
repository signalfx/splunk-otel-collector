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
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
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
	extId := component.MustNewIDWithName("disk_queue_storage", "my")
	cfg.QueueConfig.GetOrInsertDefault().StorageID = &extId
	cfg.QueueConfig.GetOrInsertDefault().WaitForResult = true
	// Use multiple consumers so that dispatched items complete out of
	// order; this exercises the storage client's key-addressed Get/Delete
	// rather than a simple single-item head-of-queue path.
	cfg.QueueConfig.GetOrInsertDefault().NumConsumers = 4
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
	require.NoError(t, l.Start(t.Context(), hostWithExtensions{
		extensions: map[component.ID]component.Component{
			extId: newDiskQueueStorageExtension(extensionSettings, extConfig),
		},
	}))
	logs := plog.NewLogs()
	logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("hello world")
	for range 10 {
		require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		require.Len(tt, sink.AllLogs(), 10)
	}, 1*time.Second, 100*time.Millisecond)
	require.NoError(t, l.Shutdown(t.Context()))
	require.NoError(t, reclogs.Shutdown(t.Context()))
}
