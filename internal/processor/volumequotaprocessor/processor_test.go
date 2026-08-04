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

package volumequotaprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestDetermineStartRatesPerfect(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalSpanCount: 1000},
			{globalSpanCount: 1000},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 1.0, newEpoch.globalSpanSamplingRate)
}

func TestDetermineStartRatesUnderRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalSpanCount: 200},
			{globalSpanCount: 300},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 1.0, newEpoch.globalSpanSamplingRate)
}

func TestDetermineStartRatesOverRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalSpanCount: 2000},
			{globalSpanCount: 2000},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 0.5, newEpoch.globalSpanSamplingRate)
}

func TestDetermineStartRatesTracesPerfect(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Traces: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalTracesCount: 1000},
			{globalTracesCount: 1000},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 1.0, newEpoch.globalTracesSamplingRate)
}

func TestDetermineStartRatesTracesUnderRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Traces: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalTracesCount: 200},
			{globalTracesCount: 300},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 1.0, newEpoch.globalTracesSamplingRate)
}

func TestDetermineStartRatesTracesOverRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Traces: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalTracesCount: 2000},
			{globalTracesCount: 2000},
		},
	}
	newEpoch := &epoch{}
	p.determineStartSamplingRate(newEpoch)
	require.Equal(t, 0.5, newEpoch.globalTracesSamplingRate)
}

func TestRotateEpoch(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
		},
	}

	p.rotateEpoch()

	require.NotNil(t, p.currentEpoch)
}

func TestRotateEpochWithPreviousEpochs(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
		},
		lookbackEpochs: []epoch{
			{globalSpanCount: 2000},
			{globalSpanCount: 2000},
		},
	}

	p.rotateEpoch()

	require.NotNil(t, p.currentEpoch)
	require.Equal(t, 0.5, p.currentEpoch.globalSpanSamplingRate)
	require.Equal(t, 1.0, p.currentEpoch.globalTracesSamplingRate)
	require.Len(t, p.currentEpoch.servicesSpansSamplingRate, 0)
	require.Len(t, p.currentEpoch.servicesTracesSamplingRate, 0)
}

func TestRotateEpochWithPreviousEpochServiceLimits(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			Lookback: 2,
			Limits: Limits{
				Spans: map[string]int64{
					"foo": 1000,
				},
				Traces: map[string]int64{
					"foo": 500,
				},
			},
		},
		lookbackEpochs: []epoch{
			{
				serviceSpansCount: map[string]int64{
					"foo": 500,
					"bar": 500,
				},
				serviceTracesCount: map[string]int64{
					"foo": 750,
				},
			},
			{
				serviceSpansCount: map[string]int64{
					"foo": 2000,
					"bar": 500,
				},
				serviceTracesCount: map[string]int64{
					"foo": 750,
				},
			},
		},
	}

	p.rotateEpoch()

	require.NotNil(t, p.currentEpoch)
	require.Equal(t, 1.0, p.currentEpoch.globalSpanSamplingRate)
	require.Equal(t, 1.0, p.currentEpoch.globalTracesSamplingRate)
	require.Len(t, p.currentEpoch.servicesSpansSamplingRate, 1)
	require.Len(t, p.currentEpoch.servicesTracesSamplingRate, 1)
	require.Equal(t, 0.8, p.currentEpoch.servicesSpansSamplingRate["foo"])
	require.Equal(t, 2.0/3, p.currentEpoch.servicesTracesSamplingRate["foo"])
}

func TestRotateEpochWithCompleteConfigNoPreviousEpochs(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000, Traces: 400},
			Limits: Limits{
				Spans: map[string]int64{
					"foo": 100,
				},
				Traces: map[string]int64{
					"bar": 50,
				},
			},
			Lookback: 2,
		},
		lookbackEpochs: []epoch{},
	}

	p.rotateEpoch()

	require.NotNil(t, p.currentEpoch)
	require.Equal(t, 1.0, p.currentEpoch.globalSpanSamplingRate)
	require.Equal(t, 1.0, p.currentEpoch.globalTracesSamplingRate)
	require.Len(t, p.currentEpoch.servicesSpansSamplingRate, 1)
	require.Len(t, p.currentEpoch.servicesTracesSamplingRate, 1)
}

func TestStartAndRunOnce(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1000},
			Lookback:     2,
			Epoch:        100 * time.Millisecond,
		},
		shutdownChan: make(chan struct{}),
	}
	require.NoError(t, p.Start(t.Context(), componenttest.NewNopHost()))
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		assert.Len(tt, p.lookbackEpochs, 2)
	}, 1*time.Second, 5*time.Millisecond)

	require.NoError(t, p.Shutdown(t.Context()))
	for _, l := range p.lookbackEpochs {
		require.Equal(t, 1.0, l.globalSpanSamplingRate)
	}
}

func TestRunOneSpanThrough(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1},
			Lookback:     2,
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())

	anotherOneSpanPayload := ptrace.NewTraces()
	secondSpan := anotherOneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	require.NoError(t, p.ConsumeTraces(t.Context(), anotherOneSpanPayload))
	v, ok = secondSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(0), v.Int())
}

func TestRunOneTraceThrough(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Traces: 1},
			Lookback:     2,
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetTraceID(pcommon.NewTraceIDEmpty())
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())

	anotherPayloadSameTrace := ptrace.NewTraces()
	secondSpan := anotherPayloadSameTrace.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	secondSpan.SetTraceID(pcommon.NewTraceIDEmpty())
	require.NoError(t, p.ConsumeTraces(t.Context(), anotherPayloadSameTrace))
	v, ok = secondSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())

	anotherPayloadAnotherTrace := ptrace.NewTraces()
	thirdSpan := anotherPayloadAnotherTrace.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	thirdSpan.SetTraceID([16]byte{0, 1})
	require.NoError(t, p.ConsumeTraces(t.Context(), anotherPayloadAnotherTrace))
	v, ok = thirdSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(0), v.Int())
}

func TestRunOneSpanInitialRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Spans: 1},
			Lookback:     2,
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	p.currentEpoch.globalSpanSamplingRate = 0.9
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(90), v.Int())
}

func TestRunOneTraceInitialRate(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			GlobalLimits: GlobalLimits{Traces: 1},
			Lookback:     2,
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	p.currentEpoch.globalTracesSamplingRate = 0.95
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(95), v.Int())
}

func TestRunOneServiceSpanThrough(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			Lookback: 2,
			Limits:   Limits{Spans: map[string]int64{"foo": 1}},
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutStr("service.name", "foo")
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())

	anotherOneSpanPayload := ptrace.NewTraces()
	secondSpan := anotherOneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	secondSpan.Attributes().PutStr("service.name", "foo")
	require.NoError(t, p.ConsumeTraces(t.Context(), anotherOneSpanPayload))
	v, ok = secondSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(0), v.Int())

	unrelatedSpanPayload := ptrace.NewTraces()
	thirdSpan := unrelatedSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	thirdSpan.Attributes().PutStr("service.name", "bar")
	require.NoError(t, p.ConsumeTraces(t.Context(), unrelatedSpanPayload))
	v, ok = thirdSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())
}

func TestRunOneServiceTraceThrough(t *testing.T) {
	p := volumeQuotaProcessor{
		config: &Config{
			Lookback: 2,
			Limits:   Limits{Traces: map[string]int64{"foo": 1}},
		},
		next: &consumertest.TracesSink{},
	}
	p.rotateEpoch()
	oneSpanPayload := ptrace.NewTraces()
	span := oneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutStr("service.name", "foo")
	span.SetTraceID([16]byte{1})
	require.NoError(t, p.ConsumeTraces(t.Context(), oneSpanPayload))
	v, ok := span.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())

	anotherOneSpanPayload := ptrace.NewTraces()
	secondSpan := anotherOneSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	secondSpan.Attributes().PutStr("service.name", "foo")
	secondSpan.SetTraceID([16]byte{2})
	require.NoError(t, p.ConsumeTraces(t.Context(), anotherOneSpanPayload))
	v, ok = secondSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(0), v.Int())

	unrelatedSpanPayload := ptrace.NewTraces()
	thirdSpan := unrelatedSpanPayload.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	thirdSpan.Attributes().PutStr("service.name", "bar")
	thirdSpan.SetTraceID([16]byte{1})
	require.NoError(t, p.ConsumeTraces(t.Context(), unrelatedSpanPayload))
	v, ok = thirdSpan.Attributes().Get("sampling.priority")
	require.True(t, ok)
	require.Equal(t, int64(100), v.Int())
}
