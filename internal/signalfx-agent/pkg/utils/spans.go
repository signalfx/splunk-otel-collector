package utils

import (
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
)

// CloneSpanSlice creates a clone of an array of trace spans.
func CloneSpanSlice(spans []*trace.Span) []*trace.Span {
	out := make([]*trace.Span, len(spans))
	for i := range spans {
		out[i] = CloneSpan(spans[i])
	}
	return out
}

// CloneSpan creates a copy of a span.
func CloneSpan(span *trace.Span) *trace.Span {
	newAnnotations := make([]*trace.Annotation, len(span.Annotations))
	for i, annotation := range span.Annotations {
		newAnnotations[i] = &(*annotation)
	}
	var parentID *string
	if span.ParentID == nil {
		parentID = nil
	} else {
		parentID = pointer.String(*span.ParentID)
	}
	var debug *bool
	if span.Debug != nil {
		debug = pointer.Bool(*span.Debug)
	} else {
		debug = nil
	}
	var shared *bool
	if span.Shared != nil {
		shared = pointer.Bool(*span.Shared)
	} else {
		shared = nil
	}

	return &trace.Span{
		TraceID:        span.TraceID,
		Name:           pointer.String(*span.Name),
		ParentID:       parentID,
		ID:             span.ID,
		Timestamp:      pointer.Int64(*span.Timestamp),
		Duration:       pointer.Int64(*span.Duration),
		Debug:          debug,
		Shared:         shared,
		LocalEndpoint:  cloneEndpoint(span.LocalEndpoint),
		RemoteEndpoint: cloneEndpoint(span.RemoteEndpoint),
		Annotations:    newAnnotations,
		Tags:           CloneStringMap(span.Tags),
		Meta:           CloneFullInterfaceMap(span.Meta),
	}
}

func cloneEndpoint(endpoint *trace.Endpoint) *trace.Endpoint {
	if endpoint == nil {
		return nil
	}

	var serviceName *string
	if endpoint.ServiceName != nil {
		serviceName = pointer.String(*endpoint.ServiceName)
	}
	var ipv4 *string
	if endpoint.Ipv4 != nil {
		ipv4 = pointer.String(*endpoint.Ipv4)
	}
	var ipv6 *string
	if endpoint.Ipv6 != nil {
		ipv6 = pointer.String(*endpoint.Ipv6)
	}
	var port *int32
	if endpoint.Port != nil {
		port = pointer.Int32(*endpoint.Port)
	}
	return &trace.Endpoint{
		ServiceName: serviceName,
		Ipv4:        ipv4,
		Ipv6:        ipv6,
		Port:        port,
	}
}
