package jaegergrpc

import (
	"context"
	"net"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/api_v2"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/neotest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ClientKind   = "CLIENT"
	ServerKind   = "SERVER"
	ProducerKind = "PRODUCER"
	ConsumerKind = "CONSUMER"
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
			Meta: map[interface{}]interface{}{
				constants.DataSourceIPKey: net.ParseIP("127.0.0.1"),
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
			Meta: map[interface{}]interface{}{
				constants.DataSourceIPKey: net.ParseIP("127.0.0.1"),
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
			Meta: map[interface{}]interface{}{
				constants.DataSourceIPKey: net.ParseIP("127.0.0.1"),
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
			Meta: map[interface{}]interface{}{
				constants.DataSourceIPKey: net.ParseIP("127.0.0.1"),
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
			Meta: map[interface{}]interface{}{
				constants.DataSourceIPKey: net.ParseIP("127.0.0.1"),
			},
		},
	}
}

func TestMonitor_Configure(t *testing.T) {
	type args struct {
		conf  *Config
		batch model.Batch
	}
	tests := []struct {
		name    string
		args    args
		want    []*trace.Span
		wantErr bool
	}{
		{
			name: "simple test",
			args: args{
				conf: &Config{
					ListenAddress: "127.0.0.1:0",
				},
				batch: model.Batch{Spans: JaegerProtoTestSpans(), Process: JaegerProtoTestProcess()},
			},
			want: SFXTestSpans(),
		},
		{
			name: "tls test",
			args: args{
				conf: &Config{
					ListenAddress: "127.0.0.1:0",
					TLS: &TLSCreds{
						CertFile: path.Join(".", "testdata", "cert.pem"),
						KeyFile:  path.Join(".", "testdata", "key.pem"),
					},
				},
				batch: model.Batch{Spans: JaegerProtoTestSpans(), Process: JaegerProtoTestProcess()},
			},
			want: SFXTestSpans(),
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			var err error

			// test monitor with test output
			testOutput := neotest.NewTestOutput()
			m := &Monitor{Output: testOutput}

			if err = m.Configure(tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer m.Shutdown() // test clean up

			// get the address from the listener
			var address string
			for address == "" {
				m.listenerLock.Lock()
				if m.ln != nil {
					address = m.ln.Addr().String()
				}
				m.listenerLock.Unlock()
				runtime.Gosched()
			}

			// build the grpc client
			var conn *grpc.ClientConn
			if tt.args.conf.TLS != nil {
				var creds credentials.TransportCredentials
				creds, err = credentials.NewClientTLSFromFile(tt.args.conf.TLS.CertFile, "localhost")
				if (err != nil) != tt.wantErr {
					t.Errorf("credentials.NewClientTLSFromFile() error = %v, wantErr = %v", err, tt.wantErr)
					return
				}
				conn, err = grpc.Dial(address, grpc.WithTransportCredentials(creds))
			} else {
				conn, err = grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			}

			// handle to the grpc client connection error
			if (err != nil) != tt.wantErr {
				t.Errorf("grpc.Dial() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			// create the jaeger grpc client
			cli := api_v2.NewCollectorServiceClient(conn)

			// send the jaeger batch to the grpc server
			_, err = cli.PostSpans(context.Background(), &api_v2.PostSpansRequest{Batch: tt.args.batch})
			if (err != nil) != tt.wantErr {
				t.Errorf("PostSpans() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			// retrieve the spans received by the monitor
			spans := testOutput.FlushSpans()
			if !reflect.DeepEqual(spans, tt.want) {
				t.Log("FlushSpans() Spans received: ")
				for i := range spans {
					t.Log(spans[i])
				}
				t.Log("FlushSpans() Spans desired: ")
				for i := range tt.want {
					t.Log(tt.want[i])
				}
				t.Errorf("FlushSpans() got = %v, want = %v", spans, tt.want)
			}
		})
	}
}
