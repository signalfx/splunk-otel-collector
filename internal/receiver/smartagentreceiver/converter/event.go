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
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

const (
	// SFxEventCategoryKey key for splunk event category,
	SFxEventCategoryKey = "com.splunk.signalfx.event_category"
	// SFxEventPropertiesKey key for splunk event properties.
	SFxEventPropertiesKey = "com.splunk.signalfx.event_properties"
)

// eventToLog converts a SFx event to a pdata.Logs entry suitable for consumption by LogConsumer.
// based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/5de076e9773bdb7617b544a57fa0a4b848cec92c/receiver/signalfxreceiver/signalfxv2_event_to_logdata.go#L27
func sfxEventToPDataLogs(event *event.Event, logger *zap.Logger) pdata.Logs {
	logs, lr := newLogs()

	lr.SetName(event.EventType)

	var unixNano int64
	if !event.Timestamp.IsZero() {
		unixNano = event.Timestamp.UnixNano()
	}
	lr.SetTimestamp(pdata.Timestamp(unixNano))

	// size for event category and dimension attributes
	attrsCapacity := 1 + len(event.Dimensions)
	if len(event.Properties) > 0 {
		attrsCapacity++
	}
	attrs := lr.Attributes()
	attrs.Clear()
	attrs.EnsureCapacity(attrsCapacity)

	if event.Category == 0 {
		// This attribute must be present or SFx exporter will not know it's an event
		attrs.InsertNull(SFxEventCategoryKey)
	} else {
		attrs.InsertInt(SFxEventCategoryKey, int64(event.Category))
	}

	for k, v := range event.Dimensions {
		attrs.InsertString(k, v)
	}

	if len(event.Properties) > 0 {
		propMapVal := pdata.NewAttributeValueMap()
		propMap := propMapVal.MapVal()
		propMap.Clear()
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
				propMap.InsertString(property, v)
			case bool:
				propMap.InsertBool(property, v)
			case int:
				propMap.InsertInt(property, int64(v))
			case int8:
				propMap.InsertInt(property, int64(v))
			case int16:
				propMap.InsertInt(property, int64(v))
			case int32:
				propMap.InsertInt(property, int64(v))
			case int64:
				propMap.InsertInt(property, v)
			case float32:
				propMap.InsertDouble(property, float64(v))
			case float64:
				propMap.InsertDouble(property, v)
			default:
				// Default to string representation.
				propMap.InsertString(property, fmt.Sprintf("%v", value))
			}
		}

		attrs.Insert(SFxEventPropertiesKey, propMapVal)
	}

	return logs
}

func newLogs() (pdata.Logs, pdata.LogRecord) {
	ld := pdata.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().InstrumentationLibraryLogs().AppendEmpty().Logs().AppendEmpty()

	return ld, lr
}
