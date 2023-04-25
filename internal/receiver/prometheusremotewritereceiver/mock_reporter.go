// Copyright 2020, OpenTelemetry Authors
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

package prometheusremotewritereceiver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// mockReporter provides a iReporter that provides some useful functionalities for
// tests (e.g.: wait for certain number of messages).
type mockReporter struct {
	TranslationErrors  []error
	wgMetricsProcessed sync.WaitGroup
	TotalCalls         uint32
	MessagesProcessed  uint32
}

var _ iReporter = (*mockReporter)(nil)

func (m *mockReporter) AddExpected(newCalls int) int {
	m.wgMetricsProcessed.Add(newCalls)
	atomic.AddUint32(&m.MessagesProcessed, uint32(newCalls))
	atomic.AddUint32(&m.TotalCalls, uint32(newCalls))
	return int(m.TotalCalls)
}

// newMockReporter returns a new instance of a mockReporter.
func newMockReporter(expectedOnMetricsProcessedCalls int) *mockReporter {
	m := mockReporter{}
	m.wgMetricsProcessed.Add(expectedOnMetricsProcessedCalls)
	return &m
}

func (m *mockReporter) StartMetricsOp(ctx context.Context) context.Context {
	return ctx
}

func (m *mockReporter) OnError(_ context.Context, err error) {
	m.TranslationErrors = append(m.TranslationErrors, err)
}

func (m *mockReporter) OnMetricsProcessed(_ context.Context, numReceivedMessages int, _ error) {
	atomic.AddUint32(&m.MessagesProcessed, uint32(numReceivedMessages))
	m.wgMetricsProcessed.Done()
}

func (m *mockReporter) OnDebugf(template string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(template, args...))
}

// WaitAllOnMetricsProcessedCalls blocks until the number of expected calls
// specified at creation of the reporter is completed.
func (m *mockReporter) WaitAllOnMetricsProcessedCalls(timeout time.Duration) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()

	go func() {
		m.wgMetricsProcessed.Wait()
		cancel()
	}()

	select {
	case <-time.After(timeout):
		return errors.New("took too long to return")
	case <-ctx.Done():
		return nil
	}
}
