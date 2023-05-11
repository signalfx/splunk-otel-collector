package utils

import (
	"testing"

	"github.com/signalfx/golib/v3/trace"
	"gotest.tools/assert"
)

func TestSpanClone(t *testing.T) {
	name := "foo"
	localhost := "127.0.0.1"
	parentID := "parentID"
	kind := "server"
	ts := int64(300000)
	duration := int64(40)
	boolfalse := false
	annotations := []*trace.Annotation{{
		Timestamp: &duration,
		Value:     &name,
	}}
	span := trace.Span{
		TraceID:   "12345",
		Name:      &name,
		ParentID:  &parentID,
		ID:        "myID",
		Kind:      &kind,
		Timestamp: &ts,
		Duration:  &duration,
		Debug:     &boolfalse,
		Shared:    &boolfalse,
		LocalEndpoint: &trace.Endpoint{
			ServiceName: &name,
			Ipv4:        &localhost,
		},
		RemoteEndpoint: nil,
		Annotations:    annotations,
		Tags:           map[string]string{"foo": "bar"},
		Meta:           map[interface{}]interface{}{"foo": "bar"},
	}
	clone := CloneSpan(&span)
	assert.Equal(t, span.TraceID, clone.TraceID)
	assert.Equal(t, *span.ParentID, *clone.ParentID)
	assert.Equal(t, *span.LocalEndpoint.ServiceName, *clone.LocalEndpoint.ServiceName)
	assert.Equal(t, *span.LocalEndpoint.Ipv4, *clone.LocalEndpoint.Ipv4)
	assert.DeepEqual(t, span.Tags, clone.Tags)
	assert.Equal(t, *span.Shared, *clone.Shared)
}

func TestSpanIncompleteClone(t *testing.T) {
	name := "foo"
	localhost := "127.0.0.1"
	parentID := "parentID"
	kind := "server"
	ts := int64(300000)
	duration := int64(40)
	boolfalse := false
	annotations := []*trace.Annotation{{
		Timestamp: &duration,
		Value:     &name,
	}}
	span := trace.Span{
		TraceID:   "12345",
		Name:      &name,
		ParentID:  &parentID,
		ID:        "myID",
		Kind:      &kind,
		Timestamp: &ts,
		Duration:  &duration,
		Debug:     nil,
		Shared:    &boolfalse,
		LocalEndpoint: &trace.Endpoint{
			ServiceName: &name,
			Ipv4:        &localhost,
		},
		RemoteEndpoint: nil,
		Annotations:    annotations,
		Tags:           map[string]string{"foo": "bar"},
		Meta:           map[interface{}]interface{}{"foo": "bar"},
	}
	clone := CloneSpan(&span)
	assert.Equal(t, span.TraceID, clone.TraceID)
	assert.Equal(t, *span.ParentID, *clone.ParentID)
	assert.Equal(t, *span.LocalEndpoint.ServiceName, *clone.LocalEndpoint.ServiceName)
	assert.Equal(t, *span.LocalEndpoint.Ipv4, *clone.LocalEndpoint.Ipv4)
	assert.DeepEqual(t, span.Tags, clone.Tags)
	assert.Equal(t, *span.Shared, *clone.Shared)
}
