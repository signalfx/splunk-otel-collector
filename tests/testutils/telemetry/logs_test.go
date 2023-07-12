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

//go:build testutils

package telemetry

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
)

func loadedResourceLogs(t *testing.T) ResourceLogs {
	resourceLogs, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "resource-logs.yaml"))
	require.NoError(t, err)
	require.NotNil(t, resourceLogs)
	return *resourceLogs
}

func TestResourceLogsYamlStringRep(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(".", "testdata", "logs", "resource-logs.yaml"))
	require.NoError(t, err)
	resourceLogs := loadedResourceLogs(t)
	require.Equal(t, string(b), fmt.Sprintf("%v", resourceLogs))
}

func TestLoadLogsHappyPath(t *testing.T) {
	resourceLogs := loadedResourceLogs(t)
	assert.Equal(t, 2, len(resourceLogs.ResourceLogs))

	firstRL := resourceLogs.ResourceLogs[0]
	firstRLAttrs := *firstRL.Resource.Attributes
	require.Equal(t, 2, len(firstRLAttrs))
	require.NotNil(t, firstRLAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRLAttrs["one_attr"])
	require.NotNil(t, firstRLAttrs["two_attr"])
	assert.Equal(t, "two_value", firstRLAttrs["two_attr"])

	assert.Equal(t, 2, len(firstRL.ScopeLogs))
	firstRLFirstSL := firstRL.ScopeLogs[0]
	require.NotNil(t, firstRLFirstSL)
	require.NotNil(t, firstRLFirstSL.Scope)
	assert.Equal(t, "without_logs", firstRLFirstSL.Scope.Name)
	assert.Equal(t, "some_version", firstRLFirstSL.Scope.Version)
	require.Nil(t, firstRLFirstSL.Logs)

	firstRLSecondSL := firstRL.ScopeLogs[1]
	require.NotNil(t, firstRLSecondSL)
	require.NotNil(t, firstRLSecondSL.Scope)
	assert.Empty(t, firstRLSecondSL.Scope.Name)
	assert.Empty(t, firstRLSecondSL.Scope.Version)
	require.NotNil(t, firstRLSecondSL.Logs)

	require.Equal(t, 2, len(firstRLSecondSL.Logs))
	firstRLSecondSLFirstLog := firstRLSecondSL.Logs[0]
	require.NotNil(t, firstRLSecondSLFirstLog)
	assert.Equal(t, "a string body", firstRLSecondSLFirstLog.Body)
	assert.Equal(t, plog.SeverityNumber(1), *firstRLSecondSLFirstLog.Severity)
	assert.Equal(t, "info", firstRLSecondSLFirstLog.SeverityText)
	firstRLSecondSLFirstLogAttrs := *firstRL.Resource.Attributes
	require.Equal(t, 2, len(firstRLSecondSLFirstLogAttrs))
	require.NotNil(t, firstRLSecondSLFirstLogAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRLSecondSLFirstLogAttrs["one_attr"])
	require.NotNil(t, firstRLSecondSLFirstLogAttrs["two_attr"])
	assert.Equal(t, "two_value", firstRLSecondSLFirstLogAttrs["two_attr"])

	firstRLSecondScopeLogSecondLog := firstRLSecondSL.Logs[1]
	require.NotNil(t, firstRLSecondScopeLogSecondLog)
	assert.Equal(t, 0, firstRLSecondScopeLogSecondLog.Body)
	assert.Empty(t, firstRLSecondScopeLogSecondLog.SeverityText)
	assert.Nil(t, firstRLSecondScopeLogSecondLog.Severity)
	assert.Nil(t, firstRLSecondScopeLogSecondLog.Attributes)

	secondRL := resourceLogs.ResourceLogs[1]
	require.Nil(t, secondRL.Resource.Attributes)

	assert.Equal(t, 1, len(secondRL.ScopeLogs))
	secondRLFirstSL := secondRL.ScopeLogs[0]
	require.NotNil(t, secondRLFirstSL)
	require.NotNil(t, secondRLFirstSL.Scope)
	assert.Equal(t, "with_logs", secondRLFirstSL.Scope.Name)
	assert.Equal(t, "another_version", secondRLFirstSL.Scope.Version)
	require.NotNil(t, secondRLFirstSL.Logs)

	require.Equal(t, 2, len(secondRLFirstSL.Logs))
	secondRLFirstSLFirstLog := secondRLFirstSL.Logs[0]
	require.NotNil(t, secondRLFirstSLFirstLog)
	assert.Equal(t, true, secondRLFirstSLFirstLog.Body)
	assert.Equal(t, plog.SeverityNumber(24), *secondRLFirstSLFirstLog.Severity)
	assert.Equal(t, "arbitrary", secondRLFirstSLFirstLog.SeverityText)
	assert.Nil(t, secondRLFirstSLFirstLog.Attributes)

	secondRLFirstScopeLogSecondLog := secondRLFirstSL.Logs[1]
	require.NotNil(t, secondRLFirstScopeLogSecondLog)
	assert.Equal(t, 0.123, secondRLFirstScopeLogSecondLog.Body)
	assert.Equal(t, plog.SeverityNumber(9), *secondRLFirstScopeLogSecondLog.Severity)
	assert.Empty(t, secondRLFirstScopeLogSecondLog.SeverityText)
	assert.Nil(t, secondRLFirstScopeLogSecondLog.Attributes)
}

func TestLoadLogsNotAValidPath(t *testing.T) {
	resourceLogs, err := LoadResourceLogs("notafile")
	require.Error(t, err)
	require.Contains(t, err.Error(), invalidPathErrorMsg())
	require.Nil(t, resourceLogs)
}

func TestLoadLogsInvalidItems(t *testing.T) {
	resourceLogs, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "invalid-resource-logs.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "field notAttributesOrScopeLogs not found in type telemetry.ResourceLog")
	require.Nil(t, resourceLogs)
}

func TestLogEquivalence(t *testing.T) {
	sev := plog.SeverityNumberFatal
	log := func() Log {
		return Log{
			Body: "a_log_body", SeverityText: "a_severity_text",
			Severity: &sev,
			Attributes: &map[string]any{
				"one": "one", "two": "two",
			},
		}
	}

	lOne := log()
	lOneSelf := log()
	assert.True(t, lOne.Equals(lOneSelf))

	lTwo := log()
	assert.True(t, lOne.Equals(lTwo))
	assert.True(t, lTwo.Equals(lOne))

	lTwo.Body = ""
	assert.False(t, lOne.Equals(lTwo))
	assert.False(t, lTwo.Equals(lOne))
	lOne.Body = ""
	assert.True(t, lOne.Equals(lTwo))
	assert.True(t, lTwo.Equals(lOne))

	lTwo.SeverityText = ""
	assert.False(t, lOne.Equals(lTwo))
	assert.False(t, lTwo.Equals(lOne))
	lOne.SeverityText = ""
	assert.True(t, lOne.Equals(lTwo))
	assert.True(t, lTwo.Equals(lOne))

	sevOne := plog.SeverityNumberError
	sevTwo := plog.SeverityNumberError
	lTwo.Severity = &sevOne
	assert.False(t, lOne.Equals(lTwo))
	assert.False(t, lTwo.Equals(lOne))
	lOne.Severity = &sevTwo
	assert.True(t, lOne.Equals(lTwo))
	assert.True(t, lTwo.Equals(lOne))

	(*lTwo.Attributes)["three"] = "three"
	assert.False(t, lOne.Equals(lTwo))
	assert.False(t, lTwo.Equals(lOne))
	(*lOne.Attributes)["three"] = "three"
	assert.True(t, lOne.Equals(lTwo))
	assert.True(t, lTwo.Equals(lOne))
}

func TestLogRelaxedEquivalence(t *testing.T) {
	sevOne := plog.SeverityNumberDebug
	lacksBodyAndSeverityText := Log{
		Attributes: &map[string]any{
			"one": "one", "two": "two",
		}, Severity: &sevOne,
	}

	sevTwo := plog.SeverityNumberDebug
	completeLog := Log{
		Body: "a_log", SeverityText: "a_description",
		Attributes: &map[string]any{
			"one": "one", "two": "two",
		}, Severity: &sevTwo,
	}

	require.True(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.False(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))

	(*lacksBodyAndSeverityText.Attributes)["three"] = "three"
	require.False(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.False(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))
	(*completeLog.Attributes)["three"] = "three"
	require.True(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.False(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))

	sev := plog.SeverityNumberTrace
	lacksBodyAndSeverityText.Severity = &sev
	require.False(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.False(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))
	completeLog.Severity = &sev
	require.True(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.False(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))

	completeLog.Body = nil
	completeLog.SeverityText = ""
	require.True(t, lacksBodyAndSeverityText.RelaxedEquals(completeLog))
	require.True(t, completeLog.RelaxedEquals(lacksBodyAndSeverityText))
}

func TestLogAttributeRelaxedEquivalence(t *testing.T) {
	sev := plog.SeverityNumberInfo
	lackingAttributes := Log{
		Body: "a_log", SeverityText: "severity_text",
		Severity: &sev,
	}

	emptyAttributes := Log{
		Body: "a_log", SeverityText: "severity_text",
		Severity: &sev, Attributes: &map[string]any{},
	}

	completeLog := Log{
		Body: "a_log", SeverityText: "severity_text",
		Severity: &sev, Attributes: &map[string]any{
			"one": "one", "two": "two",
		},
	}

	require.True(t, lackingAttributes.RelaxedEquals(completeLog))
	require.False(t, emptyAttributes.RelaxedEquals(completeLog))
}

func TestLogHashFunctionConsistency(t *testing.T) {
	sev := plog.SeverityNumberInfo
	log := Log{
		Body: "some log", SeverityText: "some severity_texti",
		Severity: &sev, Attributes: &map[string]any{
			"attributeOne": "1", "attributeTwo": "two",
		},
	}
	for i := 0; i < 100; i++ {
		require.Equal(t, "2f160310dd1c038e106ca2a3564ca561", log.Hash())
	}
}

func TestFlattenResourceLogsByResourceIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_one": nil, "attribute_two": 123.456}}
	resourceLogs := ResourceLogs{
		ResourceLogs: []ResourceLog{
			{Resource: resource},
			{Resource: resource},
			{Resource: resource},
		},
	}
	expectedResourceLogs := ResourceLogs{ResourceLogs: []ResourceLog{{Resource: resource}}}
	require.Equal(t, expectedResourceLogs, FlattenResourceLogs(resourceLogs))
}

func TestFlattenResourceLogsByScopeLogsIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_three": true, "attribute_four": 23456}}
	sm := ScopeLogs{Scope: InstrumentationScope{
		Name: "an instrumentation library", Version: "an instrumentation library version",
	}, Logs: []Log{}}
	resourceLogs := ResourceLogs{
		ResourceLogs: []ResourceLog{
			{Resource: resource, ScopeLogs: []ScopeLogs{}},
			{Resource: resource, ScopeLogs: []ScopeLogs{sm}},
			{Resource: resource, ScopeLogs: []ScopeLogs{sm, sm}},
			{Resource: resource, ScopeLogs: []ScopeLogs{sm, sm, sm}},
		},
	}
	expectedResourceLogs := ResourceLogs{
		ResourceLogs: []ResourceLog{
			{Resource: resource, ScopeLogs: []ScopeLogs{sm}},
		},
	}
	require.Equal(t, expectedResourceLogs, FlattenResourceLogs(resourceLogs))
}

func TestFlattenResourceLogsByLogsIdentity(t *testing.T) {
	sevOne := plog.SeverityNumberTrace
	sevTwo := plog.SeverityNumberTrace2
	sevThree := plog.SeverityNumberTrace3
	resource := Resource{Attributes: &map[string]any{}}
	logs := []Log{
		{Body: "a log", SeverityText: "a severity_text", Severity: &sevOne},
		{Body: "another log", SeverityText: "another severity_text", Severity: &sevTwo},
		{Body: "yet another log", SeverityText: "yet anothe severity_text", Severity: &sevThree},
	}
	sm := ScopeLogs{Logs: logs}
	smRepeated := ScopeLogs{Logs: append(logs, logs...)}
	smRepeatedTwice := ScopeLogs{Logs: append(logs, append(logs, logs...)...)}
	smWithoutLogs := ScopeLogs{}
	resourceLogs := ResourceLogs{
		ResourceLogs: []ResourceLog{
			{Resource: resource, ScopeLogs: []ScopeLogs{}},
			{Resource: resource, ScopeLogs: []ScopeLogs{sm}},
			{Resource: resource, ScopeLogs: []ScopeLogs{smRepeated}},
			{Resource: resource, ScopeLogs: []ScopeLogs{smRepeatedTwice}},
			{Resource: resource, ScopeLogs: []ScopeLogs{smWithoutLogs}},
		},
	}
	expectedResourceLogs := ResourceLogs{
		ResourceLogs: []ResourceLog{
			{Resource: resource, ScopeLogs: []ScopeLogs{sm}},
		},
	}
	require.Equal(t, expectedResourceLogs, FlattenResourceLogs(resourceLogs))
}

func TestFlattenResourceLogsConsistency(t *testing.T) {
	resourceLogs, err := PDataToResourceLogs(PDataLogs())
	require.NoError(t, err)
	require.NotNil(t, resourceLogs)
	require.Equal(t, resourceLogs, FlattenResourceLogs(resourceLogs))
	var rms []ResourceLogs
	for i := 0; i < 50; i++ {
		rms = append(rms, resourceLogs)
	}
	for i := 0; i < 50; i++ {
		require.Equal(t, resourceLogs, FlattenResourceLogs(rms...))
	}
}

func TestLogContainsAllSelfCheck(t *testing.T) {
	resourceLogs := loadedResourceLogs(t)
	containsAll, err := resourceLogs.ContainsAll(resourceLogs)
	require.True(t, containsAll, err)
	require.NoError(t, err)
}

func TestLogContainsAllNoBijection(t *testing.T) {
	received := loadedResourceLogs(t)

	expected, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "expected-logs.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	containsAll, err := received.ContainsAll(*expected)
	require.True(t, containsAll, err)
	require.NoError(t, err)

	// Since expectedLogs specify no severities, they will never find matches with logs w/ them.
	containsAll, err = expected.ContainsAll(received)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(),
		"Missing Logs: [body: true\nseverity: 24\nseverity_text: arbitrary\n body: 0.123\nseverity: 9\n]",
	)
}

func TestLogContainsAllSeverityTextNeverReceived(t *testing.T) {
	received := loadedResourceLogs(t)
	expected, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "never-received-logs.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedLogs.yaml details a Log with a value that isn't in resourceLogs.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Logs: [body: true\nseverity_text: never received\n]")
}

func TestLogContainsAllInstrumentationScopeNeverReceived(t *testing.T) {
	received := loadedResourceLogs(t)
	expected, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "never-received-instrumentation-scope.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedLogs.yaml details an InstrumentationScope that isn't in resourceLogs.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing InstrumentationScopes: [name: unmatched_instrumentation_scope\n]")
}

func TestLogContainsAllResourceNeverReceived(t *testing.T) {
	received := loadedResourceLogs(t)
	expected, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "never-received-resource.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedLogs.yaml details a Resource that isn't in resourceLogs.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing resources: [attributes:\n  not: matched\n]")
}

func TestLogContainsAllWithMissingAndEmptyAttributes(t *testing.T) {
	received, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "attribute-value-resource-logs.yaml"))
	require.NoError(t, err)
	require.NotNil(t, received)

	unspecified, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "unspecified-attributes-allowed.yaml"))
	require.NoError(t, err)
	require.NotNil(t, unspecified)

	empty, err := LoadResourceLogs(filepath.Join(".", "testdata", "logs", "empty-attributes-required.yaml"))
	require.NoError(t, err)
	require.NotNil(t, empty)

	containsAll, err := received.ContainsAll(*unspecified)
	require.True(t, containsAll)
	require.NoError(t, err)

	containsAll, err = received.ContainsAll(*empty)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Logs: [body: a string body\nattributes: {}\nseverity: 1\nseverity_text: info\n]")
}
