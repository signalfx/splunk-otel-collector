// Copyright Splunk, Inc.
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

package telemetry

import (
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// PDataToResourceTraces returns a ResourceTraces item generated from ptrace.Traces content.
func PDataToResourceTraces(pdataTraces ...ptrace.Traces) (ResourceTraces, error) {
	resourceTraces := ResourceTraces{}
	for _, pdataTrace := range pdataTraces {
		pdataResourceSpans := pdataTrace.ResourceSpans()
		for i := 0; i < pdataResourceSpans.Len(); i++ {
			resourceSpans := ResourceSpans{}
			pdataResourceSpan := pdataResourceSpans.At(i)
			sanitizedAttrs := sanitizeAttributes(pdataResourceSpan.Resource().Attributes().AsRaw())
			resourceSpans.Resource.Attributes = &sanitizedAttrs
			pdataScopedSpans := pdataResourceSpan.ScopeSpans()
			for j := 0; j < pdataScopedSpans.Len(); j++ {
				scopedSpans := ScopeSpans{Spans: []Span{}}
				pdataScopedSpan := pdataScopedSpans.At(j)
				attrs := sanitizeAttributes(pdataScopedSpan.Scope().Attributes().AsRaw())
				scopedSpans.Scope = InstrumentationScope{
					Name:       pdataScopedSpan.Scope().Name(),
					Version:    pdataScopedSpan.Scope().Version(),
					Attributes: &attrs,
				}
				pdataSpans := pdataScopedSpan.Spans()
				for k := 0; k < pdataSpans.Len(); k++ {
					pdataSpan := pdataSpans.At(k)
					spanAttrs := sanitizeAttributes(pdataSpan.Attributes().AsRaw())
					span := Span{
						Name:       pdataSpan.Name(),
						Attributes: &spanAttrs,
					}
					scopedSpans.Spans = append(scopedSpans.Spans, span)
				}
				resourceSpans.ScopeSpans = append(resourceSpans.ScopeSpans, scopedSpans)
			}
			resourceTraces.ResourceSpans = append(resourceTraces.ResourceSpans, resourceSpans)
		}
	}
	return resourceTraces, nil
}

func PDataTraces() ptrace.Traces {
	traces := ptrace.NewTraces()

	resourceTraces := traces.ResourceSpans().AppendEmpty()
	resourceTracesAttrs := resourceTraces.Resource().Attributes()
	resourceTracesAttrs.PutBool("bool", true)
	resourceTracesAttrs.PutStr("string", "a_string")
	resourceTracesAttrs.PutInt("int", 123)
	resourceTracesAttrs.PutDouble("double", 123.45)
	resourceTracesAttrs.PutEmpty("null")

	scopeSpans := resourceTraces.ScopeSpans()
	scopeOne := scopeSpans.AppendEmpty()
	scopeOne.Scope().SetName("instrumentation_scope_one")
	scopeOne.Scope().SetVersion("instrumentation_scope_one_version")

	scopeOneSpans := scopeOne.Spans()
	scopeOneSpanOne := scopeOneSpans.AppendEmpty()
	scopeOneSpanOne.SetName("span_one")
	scopeOneSpanOne.SetKind(ptrace.SpanKindClient)
	scopeOneSpanOneAttrs := scopeOneSpanOne.Attributes()
	scopeOneSpanOneAttrs.PutBool("bool", true)
	scopeOneSpanOneAttrs.PutStr("string", "a_string")
	scopeOneSpanOneAttrs.PutInt("int", 123)
	scopeOneSpanOneAttrs.PutDouble("double", 123.45)
	scopeOneSpanOneAttrs.PutEmpty("null")

	scopeOneSpanTwo := scopeOneSpans.AppendEmpty()
	scopeOneSpanTwo.SetName("span_two")
	scopeOneSpanTwo.SetKind(ptrace.SpanKindServer)

	scopeTwo := scopeSpans.AppendEmpty()
	scopeTwo.Scope().SetName("instrumentation_scope_two")
	scopeTwo.Scope().SetVersion("instrumentation_scope_two_version")

	scopeTwoSpans := scopeTwo.Spans()
	scopeTwoSpanOne := scopeTwoSpans.AppendEmpty()
	scopeTwoSpanOne.SetName("span_one")
	scopeTwoSpanOne.SetKind(ptrace.SpanKindClient)

	return traces
}
