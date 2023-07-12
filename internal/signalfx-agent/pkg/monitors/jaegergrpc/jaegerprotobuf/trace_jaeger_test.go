package jaegerprotobuf

import (
	"testing"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	. "github.com/smartystreets/goconvey/convey"
)

// JaegerProtoTestProcess returns a jaeger protobuf Process to use in tests
func JaegerProtoTestProcess() *model.Process {
	return &model.Process{
		ServiceName: "api",
		Tags: []model.KeyValue{
			{
				Key:   "hostname",
				VType: model.ValueType_STRING,
				VStr:  "api246-sjc1",
			},
			{
				Key:   "ip",
				VType: model.ValueType_STRING,
				VStr:  "10.53.69.61",
			},
			{
				Key:   "jaeger.version",
				VType: model.ValueType_STRING,
				VStr:  "Python-3.1.0",
			},
		},
	}
}

// JaegerProtoTestSpans returns an array of jaeger protobuf spans to use in tests
func JaegerProtoTestSpans() []*model.Span {
	return []*model.Span{
		{
			TraceID: model.NewTraceID(0, 5951113872249657919),
			SpanID:  model.NewSpanID(6585752),
			References: []model.SpanRef{
				model.NewChildOfRef(model.NewTraceID(0, 5951113872249657919), 6866147),
			},
			OperationName: "get",
			StartTime:     time.Unix(0, int64(1485467191639875*time.Microsecond)),
			Duration:      time.Microsecond * 22938,
			Tags: model.KeyValues{
				{
					Key:   "http.url",
					VType: model.ValueType_STRING,
					VStr:  "http://127.0.0.1:15598/client_transactions",
				},
				{
					Key:   "span.kind",
					VType: model.ValueType_STRING,
					VStr:  "server",
				},
				{
					Key:    "peer.port",
					VType:  model.ValueType_INT64,
					VInt64: 53931,
				},
				{
					Key:   "someBool",
					VType: model.ValueType_BOOL,
					VBool: true,
				},
				{
					Key:   "someFalseBool",
					VType: model.ValueType_BOOL,
					VBool: false,
				},
				{
					Key:      "someDouble",
					VType:    model.ValueType_FLOAT64,
					VFloat64: 129.8,
				},
				{
					Key:   "peer.service",
					VType: model.ValueType_STRING,
					VStr:  "rtapi",
				},
				{
					Key:    "peer.ipv4",
					VType:  model.ValueType_INT64,
					VInt64: 3224716605,
				},
			},
			Logs: []model.Log{
				{
					Timestamp: time.Unix(0, int64(1485467191639875*time.Microsecond)),
					Fields: []model.KeyValue{
						{
							Key:   "key1",
							VType: model.ValueType_STRING,
							VStr:  "value1",
						},
						{
							Key:   "key2",
							VType: model.ValueType_STRING,
							VStr:  "value2",
						},
					},
				},
				{
					Timestamp: time.Unix(0, int64(1485467191639875*time.Microsecond)),
					Fields: []model.KeyValue{
						{
							Key:   "event",
							VType: model.ValueType_STRING,
							VStr:  "nothing",
						},
					},
				},
			},
		},
		{
			TraceID: model.NewTraceID(1, 5951113872249657919),
			SpanID:  model.NewSpanID(27532398882098234),
			References: []model.SpanRef{
				model.NewChildOfRef(model.NewTraceID(1, 5951113872249657919), 6866147),
			},
			OperationName: "post",
			StartTime:     time.Unix(0, int64(1485467191639875*time.Microsecond)),
			Duration:      time.Microsecond * 22938,
			Tags: model.KeyValues{
				model.KeyValue{
					Key:   "span.kind",
					VType: model.ValueType_STRING,
					VStr:  "client",
				},
				model.KeyValue{
					Key:   "peer.port",
					VType: model.ValueType_STRING,
					VStr:  "53931",
				},
				model.KeyValue{
					Key:   "peer.ipv4",
					VType: model.ValueType_STRING,
					VStr:  "10.0.0.1",
				},
			},
		},
		{
			TraceID: model.NewTraceID(0, 5951113872249657919),
			SpanID:  model.NewSpanID(27532398882098234),
			References: []model.SpanRef{
				model.NewChildOfRef(model.NewTraceID(0, 5951113872249657919), 6866147),
			},
			OperationName: "post",
			StartTime:     time.Unix(0, int64(1485467191639875*time.Microsecond)),
			Duration:      time.Microsecond * 22938,
			Tags: model.KeyValues{
				{
					Key:   "span.kind",
					VType: model.ValueType_STRING,
					VStr:  "consumer",
				},
				{
					Key:   "peer.port",
					VType: model.ValueType_BOOL,
					VBool: true,
				},
				{
					Key:   "peer.ipv4",
					VType: model.ValueType_BOOL,
					VBool: false,
				},
			},
		},
		{
			TraceID: model.NewTraceID(0, 5951113872249657919),
			SpanID:  model.NewSpanID(27532398882098234),
			References: []model.SpanRef{
				model.NewFollowsFromRef(model.NewTraceID(0, 5951113872249657919), 6866148),
				model.NewChildOfRef(model.NewTraceID(0, 5951113872249657919), 6866147),
			},
			OperationName: "post",
			StartTime:     time.Unix(0, int64(1485467191639875*time.Microsecond)),
			Flags:         model.Flags(2),
			Duration:      time.Microsecond * 22938,
			Tags: model.KeyValues{
				{
					Key:   "span.kind",
					VType: model.ValueType_STRING,
					VStr:  "producer",
				},
				{
					Key:   "peer.ipv6",
					VType: model.ValueType_STRING,
					VStr:  "::1",
				},
			},
		},
		{
			TraceID: model.NewTraceID(0, 5951113872249657919),
			SpanID:  model.NewSpanID(27532398882098234),
			References: []model.SpanRef{
				model.NewChildOfRef(model.NewTraceID(0, 5951113872249657919), 6866147),
			},
			OperationName: "post",
			StartTime:     time.Unix(0, int64(1485467191639875*time.Microsecond)),
			Flags:         model.Flags(2),
			Duration:      time.Microsecond * 22938,
			Tags: model.KeyValues{
				{
					Key:   "span.kind",
					VType: model.ValueType_STRING,
					VStr:  "producer",
				},
				{
					Key:   "peer.ipv6",
					VType: model.ValueType_STRING,
					VStr:  "::1",
				},
				{
					Key:    "elements",
					VType:  model.ValueType_INT64,
					VInt64: 100,
				},
				{
					Key:     "binData",
					VType:   model.ValueType_BINARY,
					VBinary: []byte("abc"),
				},
				{
					Key:   "badType",
					VType: 9,
				},
			},
		},
	}
}

func SFXTestSpans() []*trace.Span {
	return []*trace.Span{
		{
			TraceID:  "52969a8955571a3f",
			ParentID: pointer.String("000000000068c4e3"),
			ID:       "0000000000647d98",
			Name:     pointer.String("get"),
			Kind:     &ServerKind,
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("api"),
				Ipv4:        pointer.String("10.53.69.61"),
			},
			RemoteEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("rtapi"),
				Ipv4:        pointer.String("192.53.69.61"),
				Port:        pointer.Int32(53931),
			},
			Timestamp: pointer.Int64(1485467191639875),
			Duration:  pointer.Int64(22938),
			Debug:     nil,
			Shared:    nil,
			Annotations: []*trace.Annotation{
				{Timestamp: pointer.Int64(1485467191639875), Value: pointer.String("{\"key1\":\"value1\",\"key2\":\"value2\"}")},
				{Timestamp: pointer.Int64(1485467191639875), Value: pointer.String("nothing")},
			},
			Tags: map[string]string{
				"http.url":       "http://127.0.0.1:15598/client_transactions",
				"someBool":       "true",
				"someFalseBool":  "false",
				"someDouble":     "129.8",
				"hostname":       "api246-sjc1",
				"jaeger.version": "Python-3.1.0",
			},
		},
		{
			TraceID:  "000000000000000152969a8955571a3f",
			ParentID: pointer.String("000000000068c4e3"),
			ID:       "0061d092272e8c3a",
			Name:     pointer.String("post"),
			Kind:     &ClientKind,
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("api"),
				Ipv4:        pointer.String("10.53.69.61"),
			},
			RemoteEndpoint: &trace.Endpoint{
				Ipv4: pointer.String("10.0.0.1"),
				Port: pointer.Int32(53931),
			},
			Timestamp:   pointer.Int64(1485467191639875),
			Duration:    pointer.Int64(22938),
			Debug:       nil,
			Shared:      nil,
			Annotations: []*trace.Annotation{},
			Tags: map[string]string{
				"hostname":       "api246-sjc1",
				"jaeger.version": "Python-3.1.0",
			},
		},
		{
			TraceID:  "52969a8955571a3f",
			ParentID: pointer.String("000000000068c4e3"),
			ID:       "0061d092272e8c3a",
			Name:     pointer.String("post"),
			Kind:     &ConsumerKind,
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("api"),
				Ipv4:        pointer.String("10.53.69.61"),
			},
			RemoteEndpoint: nil,
			Timestamp:      pointer.Int64(1485467191639875),
			Duration:       pointer.Int64(22938),
			Debug:          nil,
			Shared:         nil,
			Annotations:    []*trace.Annotation{},
			Tags: map[string]string{
				"hostname":       "api246-sjc1",
				"jaeger.version": "Python-3.1.0",
			},
		},
		{
			TraceID:  "52969a8955571a3f",
			ParentID: pointer.String("000000000068c4e3"),
			ID:       "0061d092272e8c3a",
			Name:     pointer.String("post"),
			Kind:     &ProducerKind,
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("api"),
				Ipv4:        pointer.String("10.53.69.61"),
			},
			RemoteEndpoint: &trace.Endpoint{
				Ipv6: pointer.String("::1"),
			},
			Timestamp:   pointer.Int64(1485467191639875),
			Duration:    pointer.Int64(22938),
			Debug:       pointer.Bool(true),
			Shared:      nil,
			Annotations: []*trace.Annotation{},
			Tags: map[string]string{
				"hostname":       "api246-sjc1",
				"jaeger.version": "Python-3.1.0",
			},
		},
		{
			TraceID:  "52969a8955571a3f",
			ParentID: pointer.String("000000000068c4e3"),
			ID:       "0061d092272e8c3a",
			Name:     pointer.String("post"),
			Kind:     &ProducerKind,
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("api"),
				Ipv4:        pointer.String("10.53.69.61"),
			},
			RemoteEndpoint: &trace.Endpoint{
				Ipv6: pointer.String("::1"),
			},
			Timestamp:   pointer.Int64(1485467191639875),
			Duration:    pointer.Int64(22938),
			Debug:       pointer.Bool(true),
			Shared:      nil,
			Annotations: []*trace.Annotation{},
			Tags: map[string]string{
				"elements":       "100",
				"hostname":       "api246-sjc1",
				"jaeger.version": "Python-3.1.0",
			},
		},
	}
}

// TestJaegerTraceDecoder tests the ability to decode jaeger protobuf to SignalFx
func TestJaegerTraceDecoder(t *testing.T) {
	Convey("Spans should be decoded properly", t, func() {
		var batch = model.Batch{
			Process: JaegerProtoTestProcess(),
			Spans:   JaegerProtoTestSpans(),
		}

		spans := JaegerProtoBatchToSFX(&batch)

		So(spans, ShouldResemble, SFXTestSpans())
	})
}
