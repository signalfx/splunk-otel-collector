package jaegerprotobuf

// The following code is adapted from the Jaeger Thrift To SignalFx code
//	https://github.com/signalfx/gateway/blob/master/protocol/signalfx/trace_jaeger.go

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
)

// Constants as variables so it is easy to get a pointer to them
var (
	ClientKind   = "CLIENT"
	ServerKind   = "SERVER"
	ProducerKind = "PRODUCER"
	ConsumerKind = "CONSUMER"
)

var pads = []string{
	"",
	"0",
	"00",
	"000",
	"0000",
	"00000",
	"000000",
	"0000000",
	"00000000",
	"000000000",
	"0000000000",
	"00000000000",
	"000000000000",
	"0000000000000",
	"00000000000000",
	"000000000000000",
}

// The way IDs get converted to strings in some of the jaeger code, leading 0s
// can be dropped, which will cause the ids to fail validation on our backend.
func padID(id string) string {
	var expectedLen int
	switch {
	case len(id) < 16:
		expectedLen = 16
	case len(id) > 16 && len(id) < 32:
		expectedLen = 32
	default:
		return id
	}

	return pads[expectedLen-len(id)] + id
}

func convertKind(tag *model.KeyValue) *string {
	var kind *string
	switch tag.GetVStr() {
	case string(ext.SpanKindRPCClientEnum):
		kind = &ClientKind
	case string(ext.SpanKindRPCServerEnum):
		kind = &ServerKind
	case string(ext.SpanKindProducerEnum):
		kind = &ProducerKind
	case string(ext.SpanKindConsumerEnum):
		kind = &ConsumerKind
	}
	return kind
}

func convertPeerIPv4(tag *model.KeyValue) string {
	switch tag.VType {
	case model.ValueType_STRING:
		if ip := net.ParseIP(tag.GetVStr()); ip != nil {
			return ip.To4().String()
		}
	case model.ValueType_INT64:
		localIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(localIP, uint32(tag.GetVInt64()))
		return localIP.String()
	}
	return ""
}

func convertPeerPort(tag *model.KeyValue) int32 {
	switch tag.VType {
	case model.ValueType_STRING:
		if port, err := strconv.ParseUint(tag.GetVStr(), 10, 16); err == nil {
			return int32(port)
		}
	case model.ValueType_INT64:
		return int32(tag.GetVInt64())
	}
	return 0
}

func tagValueToString(tag *model.KeyValue) string {
	switch tag.VType {
	case model.ValueType_STRING:
		return tag.GetVStr()
	case model.ValueType_FLOAT64:
		return strconv.FormatFloat(tag.GetVFloat64(), 'f', -1, 64)
	case model.ValueType_BOOL:
		if tag.GetVBool() {
			return "true"
		}
		return "false"
	case model.ValueType_INT64:
		return strconv.FormatInt(tag.GetVInt64(), 10)
	default:
		return ""
	}
}

// processJaegerTags processes special tags that get converted to the kind and remote endpoint
// fields, and throw the rest of the tags into a map that becomes the Zipkin
// Tags field.
// nolint: gocyclo
func processJaegerTags(s *model.Span) (*string, *trace.Endpoint, map[string]string) {
	var kind *string
	var remote *trace.Endpoint
	tags := make(map[string]string, len(s.Tags))

	ensureRemote := func() {
		if remote == nil {
			remote = &trace.Endpoint{}
		}
	}

	for i := range s.Tags {
		switch s.Tags[i].Key {
		case string(ext.PeerHostIPv4):
			ip := convertPeerIPv4(&s.Tags[i])
			if ip == "" {
				continue
			}
			ensureRemote()
			remote.Ipv4 = pointer.String(ip)
		// ipv6 host is always string
		case string(ext.PeerHostIPv6):
			if s.Tags[i].VStr != "" {
				ensureRemote()
				remote.Ipv6 = pointer.String(s.Tags[i].VStr)
			}
		case string(ext.PeerPort):
			port := convertPeerPort(&s.Tags[i])
			if port == 0 {
				continue
			}
			ensureRemote()
			remote.Port = &port
		case string(ext.PeerService):
			ensureRemote()
			remote.ServiceName = pointer.String(s.Tags[i].VStr)
		case string(ext.SpanKind):
			kind = convertKind(&s.Tags[i])
		default:
			val := tagValueToString(&s.Tags[i])
			if val != "" {
				tags[s.Tags[i].Key] = val
			}
		}
	}
	return kind, remote, tags
}

// materializeWithJSON converts log Fields into JSON string, or just the field
// value of the event field, if present.
func materializeWithJSON(logFields []model.KeyValue) ([]byte, error) {
	fields := make(map[string]string, len(logFields))
	for i := range logFields {
		fields[logFields[i].Key] = tagValueToString(&logFields[i])
	}
	if event, ok := fields["event"]; ok && len(fields) == 1 {
		return []byte(event), nil
	}
	return json.Marshal(fields)
}

// convertJaegerLogs
func convertJaegerLogs(logs []model.Log) []*trace.Annotation {
	annotations := make([]*trace.Annotation, 0, len(logs))
	for i := range logs {
		anno := trace.Annotation{
			Timestamp: pointer.Int64(logs[i].Timestamp.UnixNano() / int64(time.Microsecond)),
		}
		if content, err := materializeWithJSON(logs[i].Fields); err == nil {
			anno.Value = pointer.String(string(content))
		}
		annotations = append(annotations, &anno)
	}
	return annotations
}

// jaegerProtoSpanToSFX takes a jaeger protobuf span and process and returns the equivalent SignalFx Spans
func jaegerProtoSpanToSFX(jSpan *model.Span, jProcess *model.Process) *trace.Span {
	var ptrParentID *string
	if jSpan.ParentSpanID() != 0 {
		ptrParentID = pointer.String(padID(strconv.FormatUint(uint64(jSpan.ParentSpanID()), 16)))
	}

	localEndpoint := &trace.Endpoint{
		ServiceName: &jProcess.ServiceName,
	}

	var ptrDebug *bool
	if jSpan.Flags&2 > 0 {
		ptrDebug = pointer.Bool(true)
	}

	// process jaeger span tags
	kind, remoteEndpoint, tags := processJaegerTags(jSpan)

	// process jaeger process tags
	for i := range jProcess.Tags {
		t := jProcess.Tags[i]
		if t.Key == "ip" && t.VStr != "" {
			localEndpoint.Ipv4 = pointer.String(t.VStr)
		} else {
			tags[t.Key] = tagValueToString(&t)
		}
	}

	// build the trace id from the low and high values and pad them
	// some jaeger code will drop leading zeros so we need to pad them
	traceID := padID(strconv.FormatUint(jSpan.TraceID.Low, 16))
	if jSpan.TraceID.High != 0 {
		traceID = padID(strconv.FormatUint(jSpan.TraceID.High, 16) + traceID)
	}

	span := &trace.Span{
		TraceID:        traceID,
		ID:             padID(strconv.FormatUint(uint64(jSpan.SpanID), 16)),
		ParentID:       ptrParentID,
		Debug:          ptrDebug,
		Name:           pointer.String(jSpan.OperationName),
		Timestamp:      pointer.Int64(jSpan.StartTime.UnixNano() / int64(time.Microsecond)),
		Duration:       pointer.Int64(jSpan.Duration.Microseconds()),
		Kind:           kind,
		LocalEndpoint:  localEndpoint,
		RemoteEndpoint: remoteEndpoint,
		Annotations:    convertJaegerLogs(jSpan.Logs),
		Tags:           tags,
	}
	return span
}

func JaegerProtoBatchToSFX(batch *model.Batch) []*trace.Span {
	spans := make([]*trace.Span, len(batch.Spans))
	for i := range batch.Spans {
		spans[i] = jaegerProtoSpanToSFX(batch.Spans[i], batch.Process)
	}

	return spans
}
