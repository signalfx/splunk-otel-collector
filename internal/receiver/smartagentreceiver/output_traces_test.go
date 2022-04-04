// Copyright OpenTelemetry Authors
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

package smartagentreceiver

import (
	"fmt"
	"testing"

	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSendSpansWithoutNextTracesConsumer(t *testing.T) {
	output := NewOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		nil, componenttest.NewNopHost(), newReceiverCreateSettings(),
	)

	output.SendSpans(&trace.Span{TraceID: "12345678", ID: "23456789"}) // doesn't panic
}

func TestExtraSpanTags(t *testing.T) {
	output := NewOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.Empty(t, output.extraSpanTags)

	output.RemoveExtraSpanTag("not_a_known_tag")

	output.AddExtraSpanTag("a_tag", "a_value")
	assert.Equal(t, "a_value", output.extraSpanTags["a_tag"])

	cp, ok := output.Copy().(*Output)
	require.True(t, ok)
	assert.Equal(t, "a_value", cp.extraSpanTags["a_tag"])

	cp.RemoveExtraSpanTag("a_tag")
	assert.Empty(t, cp.extraSpanTags["a_tag"])
	assert.Equal(t, "a_value", output.extraSpanTags["a_tag"])

	cp.AddExtraSpanTag("another_tag", "another_tag_value")
	assert.Equal(t, "another_tag_value", cp.extraSpanTags["another_tag"])
	assert.Empty(t, output.extraSpanTags["another_tag"])
}

func TestDefaultSpanTags(t *testing.T) {
	output := NewOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.Empty(t, output.defaultSpanTags)

	output.RemoveDefaultSpanTag("not_a_known_tag")

	output.AddDefaultSpanTag("a_tag", "a_value")
	assert.Equal(t, "a_value", output.defaultSpanTags["a_tag"])

	cp, ok := output.Copy().(*Output)
	require.True(t, ok)
	assert.Equal(t, "a_value", cp.defaultSpanTags["a_tag"])

	cp.RemoveDefaultSpanTag("a_tag")
	assert.Empty(t, cp.defaultSpanTags["a_tag"])
	assert.Equal(t, "a_value", output.defaultSpanTags["a_tag"])

	cp.AddDefaultSpanTag("another_tag", "another_tag_value")
	assert.Equal(t, "another_tag_value", cp.defaultSpanTags["another_tag"])
	assert.Empty(t, output.defaultSpanTags["another_tag"])
}

func TestSendSpansWithDefaultAndExtraSpanTags(t *testing.T) {
	tracesSink := consumertest.TracesSink{}
	output := NewOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		&tracesSink, componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	output.AddExtraSpanTag("will_be_overridden", "added extra value (want)")
	output.AddDefaultSpanTag("wont_be_overridden", "added default value")

	sfxSpan0 := trace.Span{
		TraceID: "12345678",
		ID:      "23456789",
		Tags: map[string]string{
			"will_be_overridden": "span-provided value (don't want)",
			"some_tag":           "some_value",
		},
	}

	sfxSpan1 := trace.Span{
		TraceID: "34567890",
		ID:      "45678901",
		Tags: map[string]string{
			"wont_be_overridden": "span-provided value (want)",
		},
	}

	sfxSpan2 := trace.Span{
		TraceID: "56789012",
		ID:      "67890123",
	}

	// core zipkin pdata translation reverses span order
	output.SendSpans(&sfxSpan2, &sfxSpan1, &sfxSpan0)

	received := tracesSink.AllTraces()
	require.Equal(t, 1, len(received))
	trace := received[0]
	assert.Equal(t, 3, trace.SpanCount())
	resourceSpans := trace.ResourceSpans()
	require.Equal(t, 1, resourceSpans.Len())
	illSpansSlice := resourceSpans.At(0).ScopeSpans()
	require.Equal(t, 1, illSpansSlice.Len())
	illSpans := illSpansSlice.At(0).Spans()
	require.Equal(t, 3, illSpans.Len())

	span0 := illSpans.At(0)
	require.Equal(t, "00000000000000000000000012345678", span0.TraceID().HexString())
	require.Equal(t, "0000000023456789", span0.SpanID().HexString())

	extraValue, containsExtraValue := span0.Attributes().Get("will_be_overridden")
	require.True(t, containsExtraValue)
	assert.Equal(t, "added extra value (want)", extraValue.StringVal())

	defaultValue, containsDefaultValue := span0.Attributes().Get("wont_be_overridden")
	require.True(t, containsDefaultValue)
	assert.Equal(t, "added default value", defaultValue.StringVal())

	value, containsValue := span0.Attributes().Get("some_tag")
	require.True(t, containsValue)
	assert.Equal(t, "some_value", value.StringVal())

	span1 := illSpans.At(1)
	require.Equal(t, "00000000000000000000000034567890", span1.TraceID().HexString())
	require.Equal(t, "0000000045678901", span1.SpanID().HexString())

	extraValue, containsExtraValue = span1.Attributes().Get("will_be_overridden")
	require.True(t, containsExtraValue)
	assert.Equal(t, "added extra value (want)", extraValue.StringVal())

	defaultValue, containsDefaultValue = span1.Attributes().Get("wont_be_overridden")
	require.True(t, containsDefaultValue)
	assert.Equal(t, "span-provided value (want)", defaultValue.StringVal())

	span2 := illSpans.At(2)
	require.Equal(t, "00000000000000000000000056789012", span2.TraceID().HexString())
	require.Equal(t, "0000000067890123", span2.SpanID().HexString())

	extraValue, containsExtraValue = span2.Attributes().Get("will_be_overridden")
	require.True(t, containsExtraValue)
	assert.Equal(t, "added extra value (want)", extraValue.StringVal())

	defaultValue, containsDefaultValue = span2.Attributes().Get("wont_be_overridden")
	require.True(t, containsDefaultValue)
	assert.Equal(t, "added default value", defaultValue.StringVal())
}

func TestSendSpansWithConsumerError(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	rcs := newReceiverCreateSettings()
	rcs.Logger = logger

	err := fmt.Errorf("desired error")
	tracesConsumer := consumertest.NewErr(err)
	output := NewOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		tracesConsumer, componenttest.NewNopHost(), rcs,
	)

	output.SendSpans(&trace.Span{TraceID: "12345678", ID: "23456789"})

	// first log will be about lack of dimension client
	require.Equal(t, 2, logs.Len())
	entry := logs.All()[1]
	assert.Equal(t, "SendSpans has failed", entry.Message)
	assert.Equal(t, "desired error", entry.ContextMap()["error"])
}
