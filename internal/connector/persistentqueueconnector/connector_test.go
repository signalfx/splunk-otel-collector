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
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func TestQueue(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Path = t.TempDir()
	sink := &consumertest.LogsSink{}
	p, err := createLogs(
		t.Context(),
		connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter")),
		cfg,
		sink,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	require.NoError(t, p.ConsumeLogs(t.Context(), plog.NewLogs()))
	time.Sleep(2 * time.Second)
	require.Len(t, sink.AllLogs(), 1)
}

func createLd() plog.Logs {
	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.Body().SetStr("The quick brown fox")
	return ld
}

func createLd100k() plog.Logs {
	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.Body().SetStr(string(bytes.Repeat([]byte{'a'}, 100_000)))
	return ld
}

func createLd500() plog.Logs {
	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.Body().SetStr(string(bytes.Repeat([]byte{'a'}, 500)))
	return ld
}

func TestQueueMessageTooBig(t *testing.T) {
	sink := &consumertest.LogsSink{}
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 10
	cfg.Path = t.TempDir()
	p, err := createLogs(
		t.Context(),
		connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter")),
		cfg,
		sink,
	)
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	require.Error(t, p.ConsumeLogs(t.Context(), createLd()))
}

func TestQueueMessageSlowedDown(t *testing.T) {
	sink := &consumertest.LogsSink{}
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 39
	cfg.Path = t.TempDir()
	logger, _ := zap.NewDevelopment()
	settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
	settings.Logger = logger
	p, err := createLogs(
		t.Context(),
		settings,
		cfg,
		sink,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	require.NoError(t, p.ConsumeLogs(t.Context(), createLd()))
	require.NoError(t, p.ConsumeLogs(t.Context(), createLd()))
	require.NoError(t, p.ConsumeLogs(t.Context(), createLd()))
	time.Sleep(500 * time.Millisecond)
	require.Len(t, sink.AllLogs(), 1)
	time.Sleep(3 * time.Second)
	require.Len(t, sink.AllLogs(), 3)
}

func TestPush1MIn128KOut(t *testing.T) {
	sink := &consumertest.LogsSink{}
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 128000
	cfg.Path = t.TempDir()
	logger, _ := zap.NewDevelopment()
	settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
	settings.Logger = logger
	p, err := createLogs(
		t.Context(),
		settings,
		cfg,
		sink,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	for range 1_000_000 / 100_000 {
		require.NoError(t, p.ConsumeLogs(t.Context(), createLd100k()))
	}
	time.Sleep(5 * time.Second)
	found := len(sink.AllLogs())
	require.True(t, found >= 4 && found <= 6)
	time.Sleep(5 * time.Second)
	require.Len(t, sink.AllLogs(), 10)
}

func TestNoLimit(t *testing.T) {
	sink := &consumertest.LogsSink{}
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 0
	cfg.Path = t.TempDir()
	logger, _ := zap.NewDevelopment()
	settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
	settings.Logger = logger
	p, err := createLogs(
		t.Context(),
		settings,
		cfg,
		sink,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	for range 1_000_000 / 100_000 {
		require.NoError(t, p.ConsumeLogs(t.Context(), createLd100k()))
	}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		assert.Len(tt, sink.AllLogs(), 10)
	}, 1*time.Second, 10*time.Millisecond)
}

type mockLogsConsumer struct {
	received  []plog.Logs
	rejecting atomic.Bool
}

func (m *mockLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (m *mockLogsConsumer) ConsumeLogs(_ context.Context, logs plog.Logs) error {
	if m.rejecting.Load() {
		return errors.New("rejecting")
	}
	m.received = append(m.received, logs)
	return nil
}

func TestRetryForever(t *testing.T) {
	m := &mockLogsConsumer{
		rejecting: atomic.Bool{},
	}
	m.rejecting.Store(true)
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 128000
	cfg.Path = t.TempDir()
	logger, _ := zap.NewDevelopment()
	settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
	settings.Logger = logger
	p, err := createLogs(
		t.Context(),
		settings,
		cfg,
		m,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	for range 1_000_000 / 100_000 {
		require.NoError(t, p.ConsumeLogs(t.Context(), createLd100k()))
	}
	require.Empty(t, m.received)
	m.rejecting.Store(false)
	time.Sleep(10 * time.Second)
	require.Len(t, m.received, 10)
}

func TestRetryForeverStartAndStop(t *testing.T) {
	m := &mockLogsConsumer{
		rejecting: atomic.Bool{},
	}
	m.rejecting.Store(true)
	cfg := createDefaultConfig().(*Config)
	cfg.ThroughputLimit = 128000
	cfg.Path = t.TempDir()
	logger, _ := zap.NewDevelopment()
	settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
	settings.Logger = logger
	p, err := createLogs(
		t.Context(),
		settings,
		cfg,
		m,
	)
	require.NoError(t, err)
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	}()
	for range 1_000_000 / 100_000 {
		require.NoError(t, p.ConsumeLogs(t.Context(), createLd100k()))
	}
	require.Empty(t, m.received)
	require.NoError(t, p.Shutdown(context.WithoutCancel(t.Context())))
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	m.rejecting.Store(false)
	time.Sleep(10 * time.Second)
	require.Len(t, m.received, 10)
}

func BenchmarkNoLimit(b *testing.B) {
	for _, scenario := range []struct {
		fn   func() plog.Logs
		name string
		load int
	}{
		{
			name: "500",
			fn:   createLd500,
			load: 1,
		},
		{
			name: "500",
			fn:   createLd500,
			load: 10,
		},
		{
			name: "500",
			fn:   createLd500,
			load: 100,
		},
		{
			name: "500",
			fn:   createLd500,
			load: 1000,
		},
		{
			name: "500",
			fn:   createLd500,
			load: 10000,
		},
		{
			name: "100k",
			fn:   createLd100k,
			load: 1,
		},
		{
			name: "100k",
			fn:   createLd100k,
			load: 10,
		},
		{
			name: "100k",
			fn:   createLd100k,
			load: 100,
		},
		{
			name: "100k",
			fn:   createLd100k,
			load: 1000,
		},
		{
			name: "100k",
			fn:   createLd100k,
			load: 10000,
		},
		{
			name: "empty",
			fn:   createLd,
			load: 1,
		},
		{
			name: "empty",
			fn:   createLd,
			load: 10,
		},
		{
			name: "empty",
			fn:   createLd,
			load: 100,
		},

		{
			name: "empty",
			fn:   createLd,
			load: 1000,
		},
		{
			name: "empty",
			fn:   createLd,
			load: 10000,
		},
	} {
		b.Run(fmt.Sprintf("%s-load-%d", scenario.name, scenario.load), func(b *testing.B) {
			m := &mockLogsConsumer{}
			cfg := createDefaultConfig().(*Config)
			cfg.ThroughputLimit = 0
			cfg.Path = b.TempDir()
			logger, _ := zap.NewDevelopment()
			settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
			settings.Logger = logger
			p, err := createLogs(
				b.Context(),
				settings,
				cfg,
				m,
			)
			require.NoError(b, err)
			require.NoError(b, p.Start(b.Context(), componenttest.NewNopHost()))
			defer func() {
				require.NoError(b, p.Shutdown(context.WithoutCancel(b.Context())))
			}()
			b.ReportAllocs()
			for b.Loop() {
				for i := 0; i < scenario.load; i++ {
					require.NoError(b, p.ConsumeLogs(b.Context(), scenario.fn()))
				}
				assert.Eventually(b, func() bool {
					return len(m.received) == scenario.load
				}, 1*time.Second, 1*time.Millisecond, "bleh", len(m.received))
				m.received = nil
			}
		})
	}
}

func BenchmarkWithLimits(b *testing.B) {
	for _, scenario := range []struct {
		fn    func() plog.Logs
		name  string
		load  int
		limit int
	}{
		{
			name:  "100k",
			fn:    createLd100k,
			load:  1,
			limit: 128000,
		},
		{
			name:  "empty",
			fn:    createLd,
			load:  1,
			limit: 128000,
		},
		{
			name:  "100k",
			fn:    createLd100k,
			load:  10,
			limit: 128000,
		},
		{
			name:  "empty",
			fn:    createLd,
			load:  10,
			limit: 128000,
		},
		{
			name:  "empty",
			fn:    createLd,
			load:  100,
			limit: 128000,
		},
		{
			name:  "100k",
			fn:    createLd100k,
			load:  1000,
			limit: 128000,
		},
		{
			name:  "empty",
			fn:    createLd,
			load:  1000,
			limit: 128000,
		},
		{
			name:  "empty",
			fn:    createLd,
			load:  10000,
			limit: 128000,
		},
	} {
		b.Run(fmt.Sprintf("%s-load-%d-limit-%d", scenario.name, scenario.load, scenario.limit), func(b *testing.B) {
			m := &mockLogsConsumer{}
			cfg := createDefaultConfig().(*Config)
			cfg.ThroughputLimit = int32(scenario.limit) //nolint:gosec // disable G115
			cfg.Path = b.TempDir()
			logger, _ := zap.NewDevelopment()
			settings := connectortest.NewNopSettings(component.MustNewType("bandwidth_limiter"))
			settings.Logger = logger
			p, err := createLogs(
				b.Context(),
				settings,
				cfg,
				m,
			)
			require.NoError(b, err)
			require.NoError(b, p.Start(b.Context(), componenttest.NewNopHost()))
			defer func() {
				require.NoError(b, p.Shutdown(context.WithoutCancel(b.Context())))
			}()
			b.ReportAllocs()
			for b.Loop() {
				for i := 0; i < scenario.load; i++ {
					require.NoError(b, p.ConsumeLogs(b.Context(), scenario.fn()))
				}
				require.Eventually(b, func() bool {
					return len(m.received) == scenario.load
				}, 5*time.Second, 1*time.Millisecond, "received %d messages, expected: %d", len(m.received), scenario.load)
				m.received = nil
			}
		})
	}
}
