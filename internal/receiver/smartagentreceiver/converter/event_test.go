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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func TestEventToPDataLogs(tt *testing.T) {
	for _, test := range []struct {
		event       event.Event
		expectedLog plog.Logs
		name        string
	}{
		{
			name:  "event zero value",
			event: event.Event{},
			expectedLog: newExpectedLog(map[string]pcommon.Value{
				"com.splunk.signalfx.event_category": pcommon.NewValueEmpty(),
			}, 0),
		},
		{
			name: "complete event",
			event: event.Event{
				EventType: "some_event_type",
				Category:  1,
				Dimensions: map[string]string{
					"dimension_name": "dimension_value",
				},
				Properties: map[string]any{
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
			expectedLog: newExpectedLog(func() map[string]pcommon.Value {
				attrs := map[string]pcommon.Value{
					"com.splunk.signalfx.event_category": pcommon.NewValueInt(1),
					"com.splunk.signalfx.event_type":     pcommon.NewValueString("some_event_type"),
					"dimension_name":                     pcommon.NewValueString("dimension_value"),
				}
				properties := pcommon.NewValueMap()
				properties.FromRaw(
					map[string]any{
						"bool_property_name":    true,
						"string_property_name":  "some value",
						"int_property_name":     12345,
						"int8_property_name":    127,
						"int16_property_name":   23456,
						"int32_property_name":   34567,
						"int64_property_name":   45678,
						"float32_property_name": 12345.678,
						"float64_property_name": 23456.789,
					},
				)
				attrs["com.splunk.signalfx.event_properties"] = properties
				return attrs
			}(), 1000000001),
		},
		{
			name: "unsupported properties",
			event: event.Event{
				EventType: "some_event_type",
				Category:  10000000,
				Properties: map[string]any{
					"nil_property":    nil,
					"struct_property": struct{ field string }{"something"},
					"uint_property":   uint(12345),
				},
			},
			expectedLog: newExpectedLog(func() map[string]pcommon.Value {
				attrs := map[string]pcommon.Value{
					"com.splunk.signalfx.event_category": pcommon.NewValueInt(10000000),
					"com.splunk.signalfx.event_type":     pcommon.NewValueString("some_event_type"),
				}
				properties := pcommon.NewValueMap()
				properties.FromRaw(
					map[string]any{
						"struct_property": "{something}",
						"uint_property":   "12345",
					},
				)
				attrs["com.splunk.signalfx.event_properties"] = properties
				return attrs
			}(), 0),
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			log := sfxEventToPDataLogs(&test.event, zap.NewNop())
			assertLogsEqual(t, test.expectedLog, log)
		})
	}
}

func newExpectedLog(properties map[string]pcommon.Value, timestamp uint64) plog.Logs {
	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.Timestamp(timestamp))

	attrs := lr.Attributes()
	for k, v := range properties {
		v.CopyTo(attrs.PutEmpty(k))
	}

	return ld
}

func assertLogsEqual(t *testing.T, expected, received plog.Logs) {
	expectedLog := expected.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	receivedLog := received.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	assert.Equal(t, expectedLog.Timestamp(), receivedLog.Timestamp())

	assertAttributeMapContainsAll(t, expectedLog.Attributes(), receivedLog.Attributes())
	assertAttributeMapContainsAll(t, receivedLog.Attributes(), expectedLog.Attributes())
}

func assertAttributeMapContainsAll(t *testing.T, first, second pcommon.Map) {
	first.Range(func(firstKey string, firstValue pcommon.Value) bool {
		secondValue, ok := second.Get(firstKey)
		require.True(t, ok, fmt.Sprintf("first attribute %s not in second", firstKey))
		require.Equal(t, firstValue.Type(), secondValue.Type())
		if secondValue.Type() == pcommon.ValueTypeMap {
			assertAttributeMapContainsAll(t, firstValue.Map(), secondValue.Map())
			return true
		}

		if secondValue.Type() == pcommon.ValueTypeDouble {
			// account for float32 -> float64 precision
			assert.InDelta(t, firstValue.Double(), secondValue.Double(), .001)
			return true
		}

		assert.EqualValues(t, firstValue, secondValue,
			"second value doesn't match first for first key %s", firstKey,
		)
		return true
	})
}
