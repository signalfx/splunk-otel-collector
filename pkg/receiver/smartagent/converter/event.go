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

	"github.com/signalfx/golib/v3/event"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

const (
	// sfxEventCategoryKey key for splunk event category,
	sfxEventCategoryKey = "com.splunk.signalfx.event_category"
	// sfxEventPropertiesKey key for splunk event properties.
	sfxEventPropertiesKey = "com.splunk.signalfx.event_properties"
	// sfxEventType key for splunk event type
	sfxEventType = "com.splunk.signalfx.event_type"
)

// eventToLog converts a SFx event to a plog.Logs entry suitable for consumption by LogConsumer.
// based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/5de076e9773bdb7617b544a57fa0a4b848cec92c/receiver/signalfxreceiver/signalfxv2_event_to_logdata.go#L27
func sfxEventToPDataLogs(event *event.Event, logger *zap.Logger) plog.Logs {
	logs, lr := newLogs()

	var unixNano int64
	if !event.Timestamp.IsZero() {
		unixNano = event.Timestamp.UnixNano()
	}
	lr.SetTimestamp(pcommon.Timestamp(unixNano))

	// size for event category and dimension attributes
	attrsCapacity := 2 + len(event.Dimensions)
	if len(event.Properties) > 0 {
		attrsCapacity++
	}
	attrs := lr.Attributes()
	attrs.Clear()
	attrs.EnsureCapacity(attrsCapacity)

	for k, v := range event.Dimensions {
		attrs.PutString(k, v)
	}

	if event.Category == 0 {
		// This attribute must be present or SFx exporter will not know it's an event
		attrs.PutEmpty(sfxEventCategoryKey)
	} else {
		attrs.PutInt(sfxEventCategoryKey, int64(event.Category))
	}

	if event.EventType != "" {
		attrs.PutString(sfxEventType, event.EventType)
	}

	if len(event.Properties) > 0 {
		propMap := attrs.PutEmptyMap(sfxEventPropertiesKey)
		propMap.EnsureCapacity(len(event.Properties))

		for property, value := range event.Properties {
			if value == nil {
				logger.Debug("property with nil value will not be reported", zap.String("property", property))
				continue
			}

			switch v := value.(type) {
			// https://github.com/signalfx/com_signalfx_metrics_protobuf/blob/master/model/signalfx_metrics.pb.go#L567
			// bool, float64, int64, and string are only supported types.
			case string:
				propMap.PutString(property, v)
			case bool:
				propMap.PutBool(property, v)
			case int:
				propMap.PutInt(property, int64(v))
			case int8:
				propMap.PutInt(property, int64(v))
			case int16:
				propMap.PutInt(property, int64(v))
			case int32:
				propMap.PutInt(property, int64(v))
			case int64:
				propMap.PutInt(property, v)
			case float32:
				propMap.PutDouble(property, float64(v))
			case float64:
				propMap.PutDouble(property, v)
			default:
				// Default to string representation.
				propMap.PutString(property, fmt.Sprintf("%v", value))
			}
		}
	}

	return logs
}

func newLogs() (plog.Logs, plog.LogRecord) {
	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()

	return ld, lr
}
