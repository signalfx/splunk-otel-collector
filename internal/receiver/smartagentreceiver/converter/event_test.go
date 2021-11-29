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
	"fmt"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestEventToPDataLogs(tt *testing.T) {
	for _, test := range []struct {
		event       event.Event
		expectedLog pdata.Logs
		name        string
	}{
		{
			name:  "event zero value",
			event: event.Event{},
			expectedLog: newExpectedLog(
				"",
				map[string]pdata.AttributeValue{
					"com.splunk.signalfx.event_category": pdata.NewAttributeValueEmpty(),
				}, 0,
			),
		},
		{
			name: "complete event",
			event: event.Event{
				EventType: "some_event_type",
				Category:  1,
				Dimensions: map[string]string{
					"dimension_name": "dimension_value",
				},
				Properties: map[string]interface{}{
					"bool_property_name":    true,
					"string_property_name":  "some value",
					"int_property_name":     int(12345),
					"int8_property_name":    int8(127),
					"int16_property_name":   int16(23456),
					"int32_property_name":   int32(34567),
					"int64_property_name":   int64(45678),
					"float32_property_name": float32(12345.678),
					"float64_property_name": float64(23456.789),
				},
				Timestamp: time.Unix(1, 1),
			},
			expectedLog: newExpectedLog(
				"some_event_type",
				func() map[string]pdata.AttributeValue {
					attrs := map[string]pdata.AttributeValue{
						"com.splunk.signalfx.event_category": pdata.NewAttributeValueInt(1),
						"dimension_name":                     pdata.NewAttributeValueString("dimension_value"),
					}
					properties := pdata.NewAttributeValueMap()
					pdata.NewAttributeMapFromMap(
						map[string]pdata.AttributeValue{
							"bool_property_name":    pdata.NewAttributeValueBool(true),
							"string_property_name":  pdata.NewAttributeValueString("some value"),
							"int_property_name":     pdata.NewAttributeValueInt(12345),
							"int8_property_name":    pdata.NewAttributeValueInt(127),
							"int16_property_name":   pdata.NewAttributeValueInt(23456),
							"int32_property_name":   pdata.NewAttributeValueInt(34567),
							"int64_property_name":   pdata.NewAttributeValueInt(45678),
							"float32_property_name": pdata.NewAttributeValueDouble(12345.678),
							"float64_property_name": pdata.NewAttributeValueDouble(23456.789),
						},
					).CopyTo(properties.MapVal())
					attrs["com.splunk.signalfx.event_properties"] = properties
					return attrs
				}(), 1000000001,
			),
		},
		{
			name: "unsupported properties",
			event: event.Event{
				EventType: "some_event_type",
				Category:  10000000,
				Properties: map[string]interface{}{
					"nil_property":    nil,
					"struct_property": struct{ field string }{"something"},
					"uint_property":   uint(12345),
				},
			},
			expectedLog: newExpectedLog(
				"some_event_type",
				func() map[string]pdata.AttributeValue {
					attrs := map[string]pdata.AttributeValue{
						"com.splunk.signalfx.event_category": pdata.NewAttributeValueInt(10000000),
					}
					properties := pdata.NewAttributeValueMap()
					pdata.NewAttributeMapFromMap(
						map[string]pdata.AttributeValue{
							"struct_property": pdata.NewAttributeValueString("{something}"),
							"uint_property":   pdata.NewAttributeValueString("12345"),
						},
					).CopyTo(properties.MapVal())
					attrs["com.splunk.signalfx.event_properties"] = properties
					return attrs
				}(), 0,
			),
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			log := sfxEventToPDataLogs(&test.event, zap.NewNop())
			assertLogsEqual(t, test.expectedLog, log)
		})
	}
}

func newExpectedLog(eventType string, properties map[string]pdata.AttributeValue, timestamp uint64) pdata.Logs {
	ld := pdata.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().InstrumentationLibraryLogs().AppendEmpty().Logs().AppendEmpty()
	lr.SetName(eventType)
	lr.SetTimestamp(pdata.Timestamp(timestamp))

	attrs := lr.Attributes()
	for k, v := range properties {
		attrs.Insert(k, v)
	}

	return ld
}

func assertLogsEqual(t *testing.T, expected, received pdata.Logs) {
	expectedLog := expected.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)
	receivedLog := received.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)

	assert.Equal(t, expectedLog.Name(), receivedLog.Name())
	assert.Equal(t, expectedLog.Timestamp(), receivedLog.Timestamp())

	assertAttributeMapContainsAll(t, expectedLog.Attributes(), receivedLog.Attributes())
	assertAttributeMapContainsAll(t, receivedLog.Attributes(), expectedLog.Attributes())
}

func assertAttributeMapContainsAll(t *testing.T, first, second pdata.AttributeMap) {
	first.Range(func(firstKey string, firstValue pdata.AttributeValue) bool {
		secondValue, ok := second.Get(firstKey)
		require.True(t, ok, fmt.Sprintf("first attribute %s not in second", firstKey))
		require.Equal(t, firstValue.Type(), secondValue.Type())
		if secondValue.Type() == pdata.AttributeValueTypeMap {
			assertAttributeMapContainsAll(t, firstValue.MapVal(), secondValue.MapVal())
			return true
		}

		if secondValue.Type() == pdata.AttributeValueTypeDouble {
			// account for float32 -> float64 precision
			assert.InDelta(t, firstValue.DoubleVal(), secondValue.DoubleVal(), .001)
			return true
		}

		assert.EqualValues(t, firstValue, secondValue,
			"second value doesn't match first for first key %s", firstKey,
		)
		return true
	})
}
