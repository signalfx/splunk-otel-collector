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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestPDataToResourceLogsHappyPath(t *testing.T) {
	resourceLogs, err := PDataToResourceLogs(PDataLogs())
	assert.NoError(t, err)
	require.NotNil(t, resourceLogs)

	rls := resourceLogs.ResourceLogs
	assert.Len(t, rls, 1)
	rl := rls[0]
	attrs := *rl.Resource.Attributes
	assert.True(t, attrs["bool"].(bool))
	assert.Equal(t, "a_string", attrs["string"].(string))
	assert.Equal(t, 123, attrs["int"])
	assert.Equal(t, 123.45, attrs["double"])
	assert.Nil(t, attrs["null"])

	scopeLogs := rl.ScopeLogs
	assert.Len(t, scopeLogs, 3)
	assert.Equal(t, "an_instrumentation_scope_name", scopeLogs[0].Scope.Name)
	assert.Equal(t, "an_instrumentation_scope_version", scopeLogs[0].Scope.Version)

	require.Len(t, scopeLogs[0].Logs, 2)

	slOneLogOne := scopeLogs[0].Logs[0]
	assert.Equal(t, "body - string value", slOneLogOne.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_1": "attribute_value_1",
		"attribute_name_2": "attribute_value_2",
	}, *slOneLogOne.Attributes)
	assert.Equal(t, "severity-text-1", slOneLogOne.SeverityText)
	assert.Equal(t, plog.SeverityNumberTrace, *slOneLogOne.Severity)

	slOneLogTwo := scopeLogs[0].Logs[1]
	assert.Equal(t, true, slOneLogTwo.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_3": "attribute_value_3",
		"attribute_name_4": "attribute_value_4",
	}, *slOneLogTwo.Attributes)
	assert.Equal(t, "severity-text-2", slOneLogTwo.SeverityText)
	assert.Equal(t, plog.SeverityNumberDebug, *slOneLogTwo.Severity)

	assert.Equal(t, "an_instrumentation_scope_without_version_or_logs", scopeLogs[1].Scope.Name)
	assert.Equal(t, "", scopeLogs[1].Scope.Version)
	assert.Len(t, scopeLogs[1].Logs, 0)

	assert.Equal(t, "", scopeLogs[2].Scope.Name)
	assert.Equal(t, "", scopeLogs[2].Scope.Version)
	require.Len(t, scopeLogs[2].Logs, 4)

	slThreeLogOne := scopeLogs[2].Logs[0]
	assert.Equal(t, int64(1), slThreeLogOne.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_5": "attribute_value_5",
		"attribute_name_6": "attribute_value_6",
	}, *slThreeLogOne.Attributes)
	assert.Equal(t, "severity-text-3", slThreeLogOne.SeverityText)
	assert.Equal(t, plog.SeverityNumberInfo, *slThreeLogOne.Severity)

	slThreeLogTwo := scopeLogs[2].Logs[1]
	assert.Equal(t, 1.234, slThreeLogTwo.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_7": "attribute_value_7",
		"attribute_name_8": "attribute_value_8",
	}, *slThreeLogTwo.Attributes)
	assert.Equal(t, "severity-text-4", slThreeLogTwo.SeverityText)
	assert.Equal(t, plog.SeverityNumberWarn, *slThreeLogTwo.Severity)

	slThreeLogThree := scopeLogs[2].Logs[2]
	assert.Equal(t, map[string]any{"one.key": "one.val", "two.key": int64(2)}, slThreeLogThree.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_9":  "attribute_value_9",
		"attribute_name_10": "attribute_value_10",
	}, *slThreeLogThree.Attributes)
	assert.Equal(t, "severity-text-5", slThreeLogThree.SeverityText)
	assert.Equal(t, plog.SeverityNumberError, *slThreeLogThree.Severity)

	slThreeLogFour := scopeLogs[2].Logs[3]
	assert.Equal(t, []byte("bytes"), slThreeLogFour.Body)
	assert.Equal(t, map[string]any{
		"attribute_name_11": "attribute_value_11",
		"attribute_name_12": "attribute_value_12",
	}, *slThreeLogFour.Attributes)
	assert.Equal(t, "severity-text-6", slThreeLogFour.SeverityText)
	assert.Equal(t, plog.SeverityNumberFatal, *slThreeLogFour.Severity)
}
