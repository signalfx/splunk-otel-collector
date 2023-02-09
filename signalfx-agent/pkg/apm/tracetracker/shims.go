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

package tracetracker

import (
	"github.com/signalfx/golib/v3/trace"
)

var (
	_ SpanList = (*spanListWrap)(nil)
	_ Span     = (*spanWrap)(nil)
)

// Span is a generic interface for accessing span metadata.
type Span interface {
	Environment() (string, bool)
	ServiceName() (string, bool)
	Tag(string) (string, bool)
	NumTags() int
}

// SpanList is a generic interface for accessing a list of spans.
type SpanList interface {
	Len() int
	At(i int) Span
}

type spanWrap struct {
	*trace.Span
}

func (s spanWrap) Environment() (string, bool) {
	env, ok := s.Tags["environment"]
	return env, ok
}

func (s spanWrap) ServiceName() (string, bool) {
	if s.LocalEndpoint == nil || s.LocalEndpoint.ServiceName == nil || *s.LocalEndpoint.ServiceName == "" {
		return "", false
	}
	return *s.LocalEndpoint.ServiceName, true
}

func (s spanWrap) Tag(tag string) (string, bool) {
	t, ok := s.Tags[tag]
	return t, ok
}

func (s spanWrap) NumTags() int {
	return len(s.Tags)
}

type spanListWrap struct {
	spans []*trace.Span
}

func (s spanListWrap) Len() int {
	return len(s.spans)
}

func (s spanListWrap) At(i int) Span {
	return spanWrap{Span: s.spans[i]}
}
