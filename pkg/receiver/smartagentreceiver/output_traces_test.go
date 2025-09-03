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

	"github.com/signalfx/golib/v3/trace" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSendSpansWithoutNextTracesConsumer(t *testing.T) {
	output, err := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		nil, componenttest.NewNopHost(), newReceiverCreateSettings("", t),
	)
	require.NoError(t, err)
	output.SendSpans(&trace.Span{TraceID: "12345678", ID: "23456789"}) // doesn't panic
}

func TestSendSpansWithConsumerError(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	rcs := newReceiverCreateSettings("", t)
	rcs.Logger = logger

	err := fmt.Errorf("desired error")
	tracesConsumer := consumertest.NewErr(err)
	output, err := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		tracesConsumer, componenttest.NewNopHost(), rcs,
	)
	require.NoError(t, err)
	output.SendSpans(&trace.Span{TraceID: "12345678", ID: "23456789"})

	// first log will be about lack of dimension client
	require.Equal(t, 2, logs.Len())
	entry := logs.All()[1]
	assert.Equal(t, "SendSpans has failed", entry.Message)
	assert.Equal(t, "desired error", entry.ContextMap()["error"])
}
