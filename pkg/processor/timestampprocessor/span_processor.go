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

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

func newSpanAttributesProcessor(_ *zap.Logger, offsetFn func(timestamp pcommon.Timestamp) pcommon.Timestamp) processorhelper.ProcessTracesFunc {

	return func(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
		for i := 0; i < traces.ResourceSpans().Len(); i++ {
			rs := traces.ResourceSpans().At(i)
			for j := 0; j < rs.ScopeSpans().Len(); j++ {
				ss := rs.ScopeSpans().At(j)
				for k := 0; k < ss.Spans().Len(); k++ {
					span := ss.Spans().At(k)
					span.SetStartTimestamp(offsetFn(span.StartTimestamp()))
					span.SetEndTimestamp(offsetFn(span.EndTimestamp()))
					for l := 0; l < span.Events().Len(); l++ {
						e := span.Events().At(l)
						e.SetTimestamp(offsetFn(e.Timestamp()))
					}
				}
			}
		}
		return traces, nil
	}

}
