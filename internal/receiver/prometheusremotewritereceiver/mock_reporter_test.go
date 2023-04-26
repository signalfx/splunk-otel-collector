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
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// mockReporter provides a reporter that provides some useful functionalities for
// tests (e.g.: wait for certain number of messages).
type mockReporter struct {
	TotalSuccessMetrics *int32
	OpsSuccess          *sync.WaitGroup
	OpsStarted          *sync.WaitGroup
	OpsFailed           *sync.WaitGroup
	Errors              []error
}

var _ reporter = (*mockReporter)(nil)

func (m *mockReporter) AddExpectedError(newCalls int) int {
	m.OpsFailed.Add(newCalls)
	atomic.AddInt32(m.TotalSuccessMetrics, int32(newCalls))
	return int(atomic.LoadInt32(m.TotalSuccessMetrics))
}

func (m *mockReporter) AddExpectedSuccess(newCalls int) int {
	m.OpsSuccess.Add(newCalls)
	atomic.AddInt32(m.TotalSuccessMetrics, int32(newCalls))
	return int(atomic.LoadInt32(m.TotalSuccessMetrics))
}

func (m *mockReporter) AddExpectedStart(newCalls int) int {
	m.OpsStarted.Add(newCalls)
	atomic.AddInt32(m.TotalSuccessMetrics, int32(newCalls))
	return int(atomic.LoadInt32(m.TotalSuccessMetrics))
}

// newMockReporter returns a new instance of a mockReporter.
func newMockReporter() *mockReporter {
	successCalls := new(int32)
	m := mockReporter{
		OpsSuccess:          &sync.WaitGroup{},
		OpsFailed:           &sync.WaitGroup{},
		OpsStarted:          &sync.WaitGroup{},
		TotalSuccessMetrics: successCalls,
	}
	return &m
}

func (m *mockReporter) StartMetricsOp(ctx context.Context) context.Context {
	m.OpsStarted.Done()
	return ctx
}

func (m *mockReporter) OnError(_ context.Context, _ string, err error) {
	m.Errors = append(m.Errors, err)
	m.OpsFailed.Done()
}

func (m *mockReporter) OnMetricsProcessed(_ context.Context, numReceivedMessages int, _ error) {
	atomic.AddInt32(m.TotalSuccessMetrics, int32(numReceivedMessages))
	m.OpsSuccess.Done()
}

func (m *mockReporter) OnDebugf(template string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(template, args...))
}

// WaitAllOnMetricsProcessedCalls blocks until the number of expected calls
// specified at creation of the otelReporter is completed.
func (m *mockReporter) WaitAllOnMetricsProcessedCalls(timeout time.Duration) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()
	allDone := &sync.WaitGroup{}
	allDone.Add(3)

	done := make(chan string)

	go func() {
		m.OpsFailed.Wait()
		allDone.Done()
		done <- "done with failed"
	}()
	go func() {
		m.OpsSuccess.Wait()
		allDone.Done()
		done <- "done with success"
	}()
	go func() {
		m.OpsStarted.Wait()
		allDone.Done()
		done <- "done with started"
	}()
	go func() {
		allDone.Wait()
		cancel()
	}()
	var completed []string
	for {
		select {
		case completedOps := <-done:
			completed = append(completed, completedOps)
		case <-time.After(timeout):
			return fmt.Errorf("took too long to return. Ones that did: %s", completed)
		case <-ctx.Done():
			return nil
		}
	}
}
