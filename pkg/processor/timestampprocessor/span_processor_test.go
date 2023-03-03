// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package timestampprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func Test_newSpanAttributesProcessor(t *testing.T) {
	now := time.Now().UTC()
	traces := ptrace.NewTraces()
	span := traces.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(now))
	e := span.Events().AppendEmpty()
	e.SetTimestamp(pcommon.NewTimestampFromTime(now))
	proc := newSpanAttributesProcessor(zap.NewNop(), offsetFn(1*time.Hour))
	newTraces, err := proc(context.Background(), traces)
	require.NoError(t, err)
	require.Equal(t, 1, newTraces.SpanCount())
	result := newTraces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
	require.Equal(t, now.Add(1*time.Hour), result.StartTimestamp().AsTime())
	require.Equal(t, now.Add(1*time.Hour), result.Events().At(0).Timestamp().AsTime())
}
