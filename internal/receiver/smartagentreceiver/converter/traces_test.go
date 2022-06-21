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

package converter

import (
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/signalfx/golib/v3/trace"
	sfxConstants "github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

var (
	flse                     = false
	tru                      = true
	oneInt64           int64 = 1
	twoInt64           int64 = 2
	oneHundredInt32    int32 = 100
	twoHundredInt32    int32 = 200
	serviceName              = "some service"
	anotherServiceName       = "another service"
	loopback                 = "127.0.0.1"
	loopbackIPv6             = "::1"
	linkLocal                = "169.254.0.1"
	linkLocalIPv6            = "fe80::1"
)

type sfxToPDataTestCase struct {
	sfxSpans       func(_ *testing.T) []*trace.Span
	expectedTraces func(_ *testing.T) ptrace.Traces
	name           string
}

func TestSFxSpansToPDataTraces(t *testing.T) {
	tests := []sfxToPDataTestCase{
		{func(t *testing.T) []*trace.Span {
			localEndpoint := trace.Endpoint{
				ServiceName: &serviceName,
				Ipv4:        &loopback,
				Port:        &oneHundredInt32,
			}

			remoteEndpoint := trace.Endpoint{
				ServiceName: &anotherServiceName,
				Ipv4:        &linkLocal,
				Port:        &twoHundredInt32,
			}

			span := newSFxSpan(
				t,
				"some operation name",
				"server",
				"0123456789abcdef0123456789abcdef",
				"0123456789abcdef",
				"123456789abcdef0",
				&oneInt64, &twoInt64,
				&tru, &tru,
				map[string]string{
					"some tag":    "some tag value",
					"another tag": "another tag value",
				},
				map[*int64]string{
					&oneInt64: "some annotation",
					&twoInt64: "another annotation",
				},
				&localEndpoint, &remoteEndpoint,
				"127.0.0.1",
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) ptrace.Traces {
				traces := newPDataSpan(
					t,
					serviceName,
					"some operation name",
					"0123456789abcdef0123456789abcdef",
					"0123456789abcdef",
					"123456789abcdef0",
					ptrace.SpanKindServer,
					1000, 3000,
					map[string]any{
						"some tag":      "some tag value",
						"another tag":   "another tag value",
						"net.host.ip":   loopback,
						"net.host.port": 100,
						"net.peer.ip":   linkLocal,
						"net.peer.port": 200,
						"peer.service":  anotherServiceName,
					},
					map[uint64]string{
						1000: "some annotation",
						2000: "another annotation",
					},
					"127.0.0.1",
				)
				return traces
			},
			"fully populated span",
		},
		{func(t *testing.T) []*trace.Span {
			span := newSFxSpan(
				t,
				"some operation name",
				"client",
				"0",
				"1",
				"",
				&oneInt64, &twoInt64,
				&flse, &flse,
				map[string]string{
					"some tag": "some tag value",
				},
				map[*int64]string{
					&oneInt64: "some annotation",
				},
				nil, nil, "",
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) ptrace.Traces {
				traces := newPDataSpan(
					t,
					"",
					"some operation name",
					"0",
					"1",
					"0",
					ptrace.SpanKindClient,
					1000, 3000,
					map[string]any{
						"some tag": "some tag value",
					},
					map[uint64]string{
						1000: "some annotation",
					},
					"",
				)
				return traces
			},
			"missing endpoints with zero ids",
		},
		{func(t *testing.T) []*trace.Span {
			localEndpoint := trace.Endpoint{
				ServiceName: &serviceName,
				Ipv6:        &loopbackIPv6,
				Port:        &oneHundredInt32,
			}

			remoteEndpoint := trace.Endpoint{
				ServiceName: &anotherServiceName,
				Ipv6:        &linkLocalIPv6,
				Port:        &twoHundredInt32,
			}

			span := newSFxSpan(
				t,
				"some operation name",
				"consumer",
				"0123456789abcdef",
				"12345678",
				"23456789",
				&oneInt64, &twoInt64,
				nil, nil,
				map[string]string{
					"some tag":    "some tag value",
					"another tag": "another tag value",
				},
				map[*int64]string{
					&oneInt64: "some annotation",
					&twoInt64: "another annotation",
				},
				&localEndpoint, &remoteEndpoint,
				"127.0.0.1",
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) ptrace.Traces {
				traces := newPDataSpan(
					t,
					serviceName,
					"some operation name",
					"0123456789abcdef",
					"12345678",
					"23456789",
					ptrace.SpanKindConsumer,
					1000, 3000,
					map[string]any{
						"some tag":      "some tag value",
						"another tag":   "another tag value",
						"net.host.ip":   loopbackIPv6,
						"net.host.port": 100,
						"net.peer.ip":   linkLocalIPv6,
						"net.peer.port": 200,
						"peer.service":  anotherServiceName,
					},
					map[uint64]string{
						1000: "some annotation",
						2000: "another annotation",
					},
					"127.0.0.1",
				)
				return traces
			},
			"ipv6 endpoints span",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			obs, logs := observer.New(zap.DebugLevel)
			logger := zap.New(obs)

			pdataTraces, err := sfxSpansToPDataTraces(test.sfxSpans(tt), logger)
			assert.NoError(tt, err)

			expectedTraces := test.expectedTraces(tt)
			assert.Equal(tt, expectedTraces.SpanCount(), pdataTraces.SpanCount())

			resourceSpans := pdataTraces.ResourceSpans()
			expectedResourceSpans := expectedTraces.ResourceSpans()
			assert.Equal(tt, expectedResourceSpans.Len(), resourceSpans.Len())

			assertSpansAreEqual(tt, expectedResourceSpans, resourceSpans)

			assert.Zero(tt, logs.Len())
		})
	}
}

func TestSFxSpansWithDataSourceIPToPDataTraces(t *testing.T) {
	// translate a batch of two spans with different source IPs
	// and make sure they are batched separate under different resources.
	sfxSpans := []trace.Span{
		newSFxSpan(
			t,
			"op",
			"server",
			"0123456789abcdef0123456789abcdef",
			"0123456789abcdef",
			"123456789abcdef0",
			&oneInt64, &twoInt64,
			&tru, &tru,
			map[string]string{
				"some tag":    "some tag value",
				"another tag": "another tag value",
			},
			map[*int64]string{
				&oneInt64: "some annotation",
				&twoInt64: "another annotation",
			},
			nil, nil,
			"127.0.0.1",
		),
		newSFxSpan(
			t,
			"op",
			"server",
			"0123456789abcdef0123456789abcdef",
			"0123456789abcdef",
			"123456789abcdef0",
			&oneInt64, &twoInt64,
			&tru, &tru,
			map[string]string{
				"some tag":    "some tag value",
				"another tag": "another tag value",
			},
			map[*int64]string{
				&oneInt64: "some annotation",
				&twoInt64: "another annotation",
			},
			nil, nil,
			"127.0.0.2",
		),
	}
	sfxPtrSpans := []*trace.Span{
		&sfxSpans[0], &sfxSpans[1],
	}

	logger := zap.NewNop()
	pdataTraces, err := sfxSpansToPDataTraces(sfxPtrSpans, logger)
	assert.NoError(t, err)
	assert.Equal(t, pdataTraces.ResourceSpans().Len(), 2)

	resources := make([]pcommon.Resource, pdataTraces.ResourceSpans().Len())
	for i := 0; i < pdataTraces.ResourceSpans().Len(); i++ {
		resources[i] = pdataTraces.ResourceSpans().At(i).Resource()
	}

	// sort resources by ip
	sort.Slice(resources, func(i, j int) bool {
		if ip1, ok := resources[i].Attributes().Get("ip"); ok {
			if ip2, ok := resources[j].Attributes().Get("ip"); ok {
				return ip1.StringVal() < ip2.StringVal()
			}
		}
		return false
	})

	ip, exists := resources[0].Attributes().Get("ip")
	assert.True(t, exists)
	assert.Equal(t, ip, pcommon.NewValueString("127.0.0.1"))

	ip, exists = resources[1].Attributes().Get("ip")
	assert.True(t, exists)
	assert.Equal(t, ip, pcommon.NewValueString("127.0.0.2"))
}

func TestNilSFxSpanConversion(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{nil}, logger)
	assert.NoError(t, err)
	assert.Zero(t, traces.ResourceSpans().Len())
	assert.Equal(t, 0, logs.Len())
}

func TestInvalidSFxToZipkinSpanModelConversion(t *testing.T) {
	invalidAsZipkinJSON := trace.Span{}
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{&invalidAsZipkinJSON}, logger)
	assert.NoError(t, err)
	assert.Zero(t, traces.ResourceSpans().Len())

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "failed to unmarshall SFx span as Zipkin", entry.Message)
	assert.Equal(t, "valid traceId required", entry.ContextMap()["error"])
}

func TestInvalidSFxToPDataConversion(t *testing.T) {
	invalidAsPData := trace.Span{
		TraceID: "12345",
		ID:      "12345",
		Tags: map[string]string{
			// appears only way to generate error is in link content.
			"otlp.link.0": "|i|n|v|a|l|i|d|",
		},
	}
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{&invalidAsPData}, logger)
	assert.EqualError(t, err, "invalid length for ID")
	assert.Equal(t, 0, traces.ResourceSpans().Len())
	require.Equal(t, 0, logs.Len())
}

func assertSpansAreEqual(t *testing.T, expectedResourceSpans, resourceSpans ptrace.ResourceSpansSlice) {
	for i := 0; i < resourceSpans.Len(); i++ {
		resourceSpan := resourceSpans.At(i)
		expectedResourceSpan := expectedResourceSpans.At(i)

		resource := resourceSpan.Resource()
		expectedResource := expectedResourceSpan.Resource()
		assert.Equal(t, expectedResource.Attributes(), resource.Attributes())

		illSpansSlice := resourceSpan.ScopeSpans()
		expectedIllSpansSlice := expectedResourceSpan.ScopeSpans()
		assert.Equal(t, expectedIllSpansSlice.Len(), illSpansSlice.Len())

		for j := 0; j < illSpansSlice.Len(); j++ {
			illSpans := illSpansSlice.At(j)
			expectedIllSpans := expectedIllSpansSlice.At(j)
			assert.Equal(t, expectedIllSpans.Spans().Len(), illSpans.Spans().Len())

			for k := 0; k < illSpans.Spans().Len(); k++ {
				span := illSpans.Spans().At(k)
				expectedSpan := expectedIllSpans.Spans().At(k)
				assert.Equal(t, expectedSpan.Name(), span.Name())
				assert.Equal(t, expectedSpan.Kind(), span.Kind())
				assert.Equal(t, expectedSpan.TraceID(), span.TraceID())
				assert.Equal(t, expectedSpan.SpanID(), span.SpanID())
				assert.Equal(t, expectedSpan.ParentSpanID(), span.ParentSpanID())
				assert.Equal(t, expectedSpan.StartTimestamp(), span.StartTimestamp())
				assert.Equal(t, expectedSpan.EndTimestamp(), span.EndTimestamp())
				assert.Equal(t, expectedSpan.Status(), span.Status())
				assert.Equal(t, expectedSpan.Links(), span.Links())

				updateMap := func(m map[string]any) func(k string, v pcommon.Value) bool {
					return func(k string, v pcommon.Value) bool {
						switch v.Type() {
						case pcommon.ValueTypeString:
							m[k] = v.StringVal()
						case pcommon.ValueTypeInt:
							m[k] = v.IntVal()
						}
						return true
					}
				}

				attributeMap := map[string]any{}
				span.Attributes().Range(updateMap(attributeMap))

				expectedAttributeMap := map[string]any{}
				expectedSpan.Attributes().Range(updateMap(expectedAttributeMap))

				assert.Equal(t, expectedAttributeMap, attributeMap)

				eventMap := map[uint64]string{}
				for l := 0; l < span.Events().Len(); l++ {
					event := span.Events().At(l)
					eventMap[uint64(event.Timestamp())] = event.Name()
					assert.Zero(t, event.Attributes().Len())
				}

				expectedEventMap := map[uint64]string{}
				for l := 0; l < expectedSpan.Events().Len(); l++ {
					event := expectedSpan.Events().At(l)
					expectedEventMap[uint64(event.Timestamp())] = event.Name()
				}

				assert.Equal(t, expectedEventMap, eventMap)
			}
		}
	}
}

func newSFxSpan(
	_ *testing.T,
	name, kind, traceID, spanID, parentID string,
	timestamp, duration *int64,
	debug, shared *bool, // These appear to have no adoption in Collector's zipkin translator
	tags map[string]string, annotations map[*int64]string,
	localEndpoint, remoteEndpoint *trace.Endpoint,
	dataSourceIP string,
) trace.Span {
	spanKind := strings.ToUpper(kind)
	pID := &parentID
	span := trace.Span{
		Name:           &name,
		Kind:           &spanKind,
		TraceID:        traceID,
		ID:             spanID,
		ParentID:       pID,
		Timestamp:      timestamp,
		Duration:       duration,
		Debug:          debug,
		Shared:         shared,
		LocalEndpoint:  localEndpoint,
		RemoteEndpoint: remoteEndpoint,
	}
	if dataSourceIP != "" {
		if span.Meta == nil {
			span.Meta = map[any]any{}
		}
		span.Meta[sfxConstants.DataSourceIPKey] = net.ParseIP(dataSourceIP)
	}

	if len(tags) > 0 {
		span.Tags = map[string]string{}
	}

	for k, v := range tags {
		span.Tags[k] = v
	}

	for k, v := range annotations {
		value := v
		span.Annotations = append(
			span.Annotations,
			&trace.Annotation{Timestamp: k, Value: &value},
		)
	}
	return span
}

func newPDataSpan(
	t *testing.T,
	serviceName, name, traceID, spanID, parentID string,
	kind ptrace.SpanKind,
	startTime, endTime uint64,
	attributes map[string]any, events map[uint64]string,
	dataSourceIP string,
) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	if serviceName != "" {
		rs.Resource().Attributes().InsertString("service.name", serviceName)
	}
	if dataSourceIP != "" {
		rs.Resource().Attributes().InsertString("ip", dataSourceIP)
	}
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()

	span.SetName(name)

	decodedTraceID, err := hex.DecodeString(fmt.Sprintf("%032s", traceID))
	require.NoError(t, err)
	tID := [16]byte{}
	copy(tID[:], decodedTraceID)

	decodedSpanID, err := hex.DecodeString(fmt.Sprintf("%016s", spanID))
	require.NoError(t, err)
	sID := [8]byte{}
	copy(sID[:], decodedSpanID)

	decodedParentID, err := hex.DecodeString(fmt.Sprintf("%016s", parentID))
	require.NoError(t, err)
	pID := [8]byte{}
	copy(pID[:], decodedParentID)

	span.SetTraceID(pcommon.NewTraceID(tID))
	span.SetSpanID(pcommon.NewSpanID(sID))
	span.SetParentSpanID(pcommon.NewSpanID(pID))

	span.SetKind(kind)
	span.SetStartTimestamp(pcommon.Timestamp(startTime))
	span.SetEndTimestamp(pcommon.Timestamp(endTime))
	attrs := span.Attributes()
	attrs.Clear()
	attrs.EnsureCapacity(len(attributes))
	for k, val := range attributes {
		switch v := val.(type) {
		case string:
			attrs.InsertString(k, v)
		default:
			vInt, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			require.NoError(t, err)
			attrs.InsertInt(k, vInt)
		}
	}

	spanEvents := span.Events()
	spanEvents.EnsureCapacity(len(events))
	for ts, v := range events {
		spanEvent := spanEvents.AppendEmpty()
		spanEvent.SetTimestamp(pcommon.Timestamp(ts))
		spanEvent.SetName(v)
	}

	return td
}
