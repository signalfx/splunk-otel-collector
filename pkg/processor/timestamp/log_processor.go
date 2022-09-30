// Copyright  Splunk, Inc.
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

package timestamp

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

func newLogAttributesProcessor(_ *zap.Logger, offsetFn func(timestamp pcommon.Timestamp) pcommon.Timestamp) processorhelper.ProcessLogsFunc {
	return func(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
		for i := 0; i < logs.ResourceLogs().Len(); i++ {
			rs := logs.ResourceLogs().At(i)
			for j := 0; j < rs.ScopeLogs().Len(); j++ {
				ss := rs.ScopeLogs().At(j)
				for k := 0; k < ss.LogRecords().Len(); k++ {
					log := ss.LogRecords().At(k)
					log.SetTimestamp(offsetFn(log.Timestamp()))
					log.SetObservedTimestamp(offsetFn(log.ObservedTimestamp()))
				}
			}
		}
		return logs, nil
	}
}
