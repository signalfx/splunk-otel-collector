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

package splunk

import (
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
)

// Just a dummy interface that covers all types of HEC inputs
type logEntry interface{}

type eventdata struct {
	Category   event.Category    `json:"category"`
	EventType  string            `json:"eventType"`
	Meta       map[string]string `json:"meta"`
	Dimensions map[string]string `json:"dimensions"`
	Properties map[string]string `json:"properties"`
}

type spandata struct {
	Meta           map[string]string   `json:"meta"`
	Debug          *bool               `json:"debug,omitempty"`
	Name           *string             `json:"name,omitempty"`
	Duration       *int64              `json:"duration,omitempty"`
	Kind           *string             `json:"kind,omitempty"`
	ParentID       *string             `json:"parentID,omitempty"`
	Shared         *bool               `json:"shared,omitempty"`
	Tags           map[string]string   `json:"tags"`
	Annotations    []*trace.Annotation `json:"annotations,omitempty"`
	ID             string              `json:"id"`
	TraceID        string              `json:"traceID"`
	LocalEndpoint  *trace.Endpoint     `json:"localEndpoint,omitempty"`
	RemoteEndpoint *trace.Endpoint     `json:"remoteEndpoint,omitempty"`
}

// This is the format that the HEC input accepts
type logMetric struct {
	Time       int64             `json:"time"`                 // epoch time
	Host       string            `json:"host"`                 // hostname
	Source     string            `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string            `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string            `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      string            `json:"event"`                // type of event: this is a metric.
	Fields     map[string]string `json:"fields"`               // metric data
}

// This is the format that the HEC input accepts
type logEvent struct {
	Time       int64     `json:"time"`                 // epoch time
	Host       string    `json:"host"`                 // hostname
	Source     string    `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string    `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string    `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      eventdata `json:"event"`                // event data
}

// This is the format that the HEC input accepts
type logSpan struct {
	Time       int64    `json:"time"`                 // epoch time
	Host       string   `json:"host"`                 // hostname
	Source     string   `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string   `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string   `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      spandata `json:"event"`                // event data -- span
}
