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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func Test_newLogAttributesProcessor(t *testing.T) {
	now := time.Now().UTC()
	logs := plog.NewLogs()
	lr := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(now))
	cfg := &Config{Offset: "+1h"}
	proc := newLogAttributesProcessor(zap.NewNop(), cfg.offsetFn())
	newLogs, err := proc(context.Background(), logs)
	require.NoError(t, err)
	require.Equal(t, 1, newLogs.LogRecordCount())
	result := newLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	require.True(t, result.ObservedTimestamp() == pcommon.Timestamp(0))
	require.Equal(t, now.Add(1*time.Hour), result.Timestamp().AsTime())
}
