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

package telemetry

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// PDataToResourceLogs returns a ResourceLogs item generated from plog.Logs content.
func PDataToResourceLogs(pdataLogs ...plog.Logs) (ResourceLogs, error) {
	resourceLogs := ResourceLogs{}
	for _, pdataLog := range pdataLogs {
		pdataRLs := pdataLog.ResourceLogs()
		numRM := pdataRLs.Len()
		for i := 0; i < numRM; i++ {
			rl := ResourceLog{}
			pdataRL := pdataRLs.At(i)
			sanitizedAttrs := sanitizeAttributes(pdataRL.Resource().Attributes().AsRaw())
			rl.Resource.Attributes = &sanitizedAttrs
			pdataSLs := pdataRL.ScopeLogs()
			for j := 0; j < pdataSLs.Len(); j++ {
				sls := ScopeLogs{Logs: []Log{}}
				pdataSL := pdataSLs.At(j)
				attrs := pdataSL.Scope().Attributes().AsRaw()
				sls.Scope = InstrumentationScope{
					Name:       pdataSL.Scope().Name(),
					Version:    pdataSL.Scope().Version(),
					Attributes: &attrs,
				}
				for k := 0; k < pdataSL.LogRecords().Len(); k++ {
					pdLR := pdataSL.LogRecords().At(k)
					log := Log{}
					log.Body = pdLR.Body().AsRaw()
					lAttrs := sanitizeAttributes(pdLR.Attributes().AsRaw())
					log.Attributes = &lAttrs
					log.Timestamp = pdLR.Timestamp().AsTime()
					log.ObservedTimestamp = pdLR.ObservedTimestamp().AsTime()
					sn := pdLR.SeverityNumber()
					log.Severity = &sn
					log.SeverityText = pdLR.SeverityText()
					sls.Logs = append(sls.Logs, log)
				}
				rl.ScopeLogs = append(rl.ScopeLogs, sls)
			}
			resourceLogs.ResourceLogs = append(resourceLogs.ResourceLogs, rl)
		}
	}
	return resourceLogs, nil
}

func PDataLogs() plog.Logs {
	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	attrs := resourceLogs.Resource().Attributes()
	attrs.PutBool("bool", true)
	attrs.PutString("string", "a_string")
	attrs.PutInt("int", 123)
	attrs.PutDouble("double", 123.45)
	attrs.PutEmpty("null")

	scopeLogs := resourceLogs.ScopeLogs()
	slOne := scopeLogs.AppendEmpty()
	slOne.Scope().SetName("an_instrumentation_scope_name")
	slOne.Scope().SetVersion("an_instrumentation_scope_version")
	slOneLogs := slOne.LogRecords()
	slOneLogOne := slOneLogs.AppendEmpty()

	slOneLogOne.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1970, 1, 2, 3, 4, 5, 6, time.UTC)))
	slOneLogOne.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1990, 1, 2, 3, 4, 5, 6, time.UTC)))
	slOneLogOne.SetSeverityNumber(plog.SeverityNumberTrace)
	slOneLogOne.SetSeverityText("severity-text-1")
	slOneLogOne.Attributes().PutString("attribute_name_1", "attribute_value_1")
	slOneLogOne.Attributes().PutString("attribute_name_2", "attribute_value_2")
	slOneLogOne.Body().SetStr("body - string value")

	slOneLogTwo := slOneLogs.AppendEmpty()
	slOneLogTwo.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1971, 1, 2, 3, 4, 5, 6, time.UTC)))
	slOneLogTwo.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1991, 1, 2, 3, 4, 5, 6, time.UTC)))
	slOneLogTwo.SetSeverityNumber(plog.SeverityNumberDebug)
	slOneLogTwo.SetSeverityText("severity-text-2")
	slOneLogTwo.Attributes().PutString("attribute_name_3", "attribute_value_3")
	slOneLogTwo.Attributes().PutString("attribute_name_4", "attribute_value_4")
	slOneLogTwo.Body().SetBool(true)

	scopeLogs.AppendEmpty().Scope().SetName("an_instrumentation_scope_without_version_or_logs")

	slThreeLogs := scopeLogs.AppendEmpty().LogRecords()
	slThreeLogOne := slThreeLogs.AppendEmpty()
	slThreeLogOne.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1972, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogOne.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1992, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogOne.SetSeverityNumber(plog.SeverityNumberInfo)
	slThreeLogOne.SetSeverityText("severity-text-3")
	slThreeLogOne.Attributes().PutString("attribute_name_5", "attribute_value_5")
	slThreeLogOne.Attributes().PutString("attribute_name_6", "attribute_value_6")
	slThreeLogOne.Body().SetInt(1)

	slThreeLogTwo := slThreeLogs.AppendEmpty()
	slThreeLogTwo.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1973, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogTwo.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1993, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogTwo.SetSeverityNumber(plog.SeverityNumberWarn)
	slThreeLogTwo.SetSeverityText("severity-text-4")
	slThreeLogTwo.Attributes().PutString("attribute_name_7", "attribute_value_7")
	slThreeLogTwo.Attributes().PutString("attribute_name_8", "attribute_value_8")
	slThreeLogTwo.Body().SetDouble(1.234)

	slThreeLogThree := slThreeLogs.AppendEmpty()
	slThreeLogThree.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1974, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogThree.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1994, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogThree.SetSeverityNumber(plog.SeverityNumberError)
	slThreeLogThree.SetSeverityText("severity-text-5")
	slThreeLogThree.Attributes().PutString("attribute_name_9", "attribute_value_9")
	slThreeLogThree.Attributes().PutString("attribute_name_10", "attribute_value_10")
	slThreeLogThree.Body().SetEmptyMap().FromRaw(map[string]any{"one.key": "one.val", "two.key": 2})

	slThreeLogFour := slThreeLogs.AppendEmpty()
	slThreeLogFour.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(1975, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogFour.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Date(1995, 1, 2, 3, 4, 5, 6, time.UTC)))
	slThreeLogFour.SetSeverityNumber(plog.SeverityNumberFatal)
	slThreeLogFour.SetSeverityText("severity-text-6")
	slThreeLogFour.Attributes().PutString("attribute_name_11", "attribute_value_11")
	slThreeLogFour.Attributes().PutString("attribute_name_12", "attribute_value_12")
	slThreeLogFour.Body().SetEmptyBytes().FromRaw([]byte("bytes"))
	return logs
}
