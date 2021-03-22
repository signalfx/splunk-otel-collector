// Copyright 2021 Splunk, Inc.
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
	"strconv"
	"strings"
	"testing"

	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
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

func TestSFxSpansToPDataTraces(t *testing.T) {
	tests := []struct {
		sfxSpans       func(_ *testing.T) []*trace.Span
		expectedTraces func(_ *testing.T) pdata.Traces
		name           string
	}{
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
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) pdata.Traces {
				traces := newPDataSpan(
					t,
					serviceName,
					"some operation name",
					"0123456789abcdef0123456789abcdef",
					"0123456789abcdef",
					"123456789abcdef0",
					pdata.SpanKindSERVER,
					1000, 3000,
					map[string]interface{}{
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
				nil, nil,
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) pdata.Traces {
				traces := newPDataSpan(
					t,
					"",
					"some operation name",
					"0",
					"1",
					"0",
					pdata.SpanKindCLIENT,
					1000, 3000,
					map[string]interface{}{
						"some tag": "some tag value",
					},
					map[uint64]string{
						1000: "some annotation",
					},
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
			)
			return []*trace.Span{&span}
		},
			func(t *testing.T) pdata.Traces {
				traces := newPDataSpan(
					t,
					serviceName,
					"some operation name",
					"0123456789abcdef",
					"12345678",
					"23456789",
					pdata.SpanKindCONSUMER,
					1000, 3000,
					map[string]interface{}{
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

			c := NewConverter(logger)
			pdataTraces := c.SpansToPDataTraces(test.sfxSpans(tt))
			require.NotNil(t, pdataTraces)

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

func TestNilSFxToZipkinSpanConversion(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	c := NewConverter(logger)
	traces := c.SpansToPDataTraces([]*trace.Span{nil})
	assert.Zero(t, traces.ResourceSpans().Len())
	assert.Equal(t, 0, logs.Len())
}

func TestInvalidSFxToZipkinSpanModelConversion(t *testing.T) {
	invalidAsZipkinJSON := trace.Span{}
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	c := NewConverter(logger)
	traces := c.SpansToPDataTraces([]*trace.Span{&invalidAsZipkinJSON})
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

	c := NewConverter(logger)
	traces := c.SpansToPDataTraces([]*trace.Span{&invalidAsPData})
	// error is logged but core conversion persists
	assert.Equal(t, 1, traces.ResourceSpans().Len())

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "error converting SFx spans to pdata.Traces", entry.Message)
	assert.Equal(t, "invalid length for ID", entry.ContextMap()["error"])
}

func assertSpansAreEqual(t *testing.T, expectedResourceSpans, resourceSpans pdata.ResourceSpansSlice) {
	for i := 0; i < resourceSpans.Len(); i++ {
		resourceSpan := resourceSpans.At(i)
		expectedResourceSpan := expectedResourceSpans.At(i)

		resource := resourceSpan.Resource()
		expectedResource := expectedResourceSpan.Resource()
		assert.Equal(t, expectedResource.Attributes(), resource.Attributes())

		illSpansSlice := resourceSpan.InstrumentationLibrarySpans()
		expectedIllSpansSlice := expectedResourceSpan.InstrumentationLibrarySpans()
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
				assert.Equal(t, expectedSpan.StartTime(), span.StartTime())
				assert.Equal(t, expectedSpan.EndTime(), span.EndTime())
				assert.Equal(t, expectedSpan.Status(), span.Status())
				assert.Equal(t, expectedSpan.Links(), span.Links())

				updateMap := func(m map[string]interface{}) func(k string, v pdata.AttributeValue) {
					return func(k string, v pdata.AttributeValue) {
						switch v.Type() {
						case pdata.AttributeValueSTRING:
							m[k] = v.StringVal()
						case pdata.AttributeValueINT:
							m[k] = v.IntVal()
						}
					}
				}

				attributeMap := map[string]interface{}{}
				span.Attributes().ForEach(updateMap(attributeMap))

				expectedAttributeMap := map[string]interface{}{}
				expectedSpan.Attributes().ForEach(updateMap(expectedAttributeMap))

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
	kind pdata.SpanKind,
	startTime, endTime uint64,
	attributes map[string]interface{}, events map[uint64]string,
) pdata.Traces {
	td := pdata.NewTraces()
	td.ResourceSpans().Resize(1)
	rs := td.ResourceSpans().At(0)
	if serviceName != "" {
		rs.Resource().Attributes().InsertString("service.name", serviceName)
	}
	rs.InstrumentationLibrarySpans().Resize(1)
	ils := rs.InstrumentationLibrarySpans().At(0)
	ils.Spans().Resize(1)
	span := ils.Spans().At(0)

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

	span.SetTraceID(pdata.NewTraceID(tID))
	span.SetSpanID(pdata.NewSpanID(sID))
	span.SetParentSpanID(pdata.NewSpanID(pID))

	span.SetKind(kind)
	span.SetStartTime(pdata.Timestamp(startTime))
	span.SetEndTime(pdata.Timestamp(endTime))
	attrs := span.Attributes()
	attrs.InitEmptyWithCapacity(len(attributes))
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
	spanEvents.Resize(len(events))
	var i int
	for ts, v := range events {
		spanEvents.At(i).SetTimestamp(pdata.Timestamp(ts))
		spanEvents.At(i).SetName(v)
		i++
	}

	return td
}
