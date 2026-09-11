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

package gnmireceiver

import (
	"testing"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInterfaceLeafPath(leaf string) []string {
	return []string{"interfaces", "interface", "state", leaf}
}

func loadTestSchema(t *testing.T) *yangSchema {
	t.Helper()
	schema, err := loadYangSchema([]string{"testdata/yang"})
	require.NoError(t, err)
	return schema
}

func TestLoadYangSchemaNoFilesReturnsEmptySchema(t *testing.T) {
	schema, err := loadYangSchema(nil)
	require.NoError(t, err)
	_, ok := schema.lookup(testInterfaceLeafPath("oper-status"))
	assert.False(t, ok)
}

func TestLoadYangSchemaMissingFileFails(t *testing.T) {
	_, err := loadYangSchema([]string{"testdata/yang/does-not-exist.yang"})
	require.Error(t, err)
}

func TestLoadYangSchemaSyntaxErrorFails(t *testing.T) {
	_, err := loadYangSchema([]string{"testdata/yang-broken"})
	require.Error(t, err)
}

func TestLoadYangSchemaResolvesCounter64ToSum(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup([]string{"interfaces", "interface", "state", "counters", "in-octets"})
	require.True(t, ok)
	assert.Equal(t, metricTypeSum, rm.Type)
	assert.Equal(t, "bytes", rm.Unit)
	assert.Equal(t, valueKindInt, rm.kind)
}

func TestLoadYangSchemaResolvesCounter32ToSum(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup([]string{"interfaces", "interface", "state", "counters", "in-errors"})
	require.True(t, ok)
	assert.Equal(t, metricTypeSum, rm.Type)
	assert.Equal(t, valueKindInt, rm.kind)
}

func TestLoadYangSchemaResolvesPlainUint64ToGauge(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup([]string{"interfaces", "interface", "state", "counters", "in-pkts"})
	require.True(t, ok)
	assert.Equal(t, metricTypeGauge, rm.Type)
	assert.Equal(t, valueKindInt, rm.kind)
}

func TestLoadYangSchemaResolvesDecimal64ToFloatGauge(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup([]string{"interfaces", "interface", "state", "counters", "temperature"})
	require.True(t, ok)
	assert.Equal(t, metricTypeGauge, rm.Type)
	assert.Equal(t, valueKindFloat, rm.kind)
}

func TestLoadYangSchemaResolvesEnum(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup(testInterfaceLeafPath("oper-status"))
	require.True(t, ok)
	assert.Equal(t, metricTypeGauge, rm.Type)
	assert.Equal(t, valueKindString, rm.kind)
	assert.Equal(t, []string{"DOWN", "TESTING", "UP"}, rm.EnumValues)
}

func TestLoadYangSchemaResolvesIdentityrefTransitively(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup(testInterfaceLeafPath("negotiated-speed"))
	require.True(t, ok)
	assert.Equal(t, metricTypeGauge, rm.Type)
	assert.Equal(t, valueKindString, rm.kind)
	assert.ElementsMatch(t, []string{"SPEED_10MB", "SPEED_100MB", "SPEED_1GB"}, rm.EnumValues)
}

func TestLoadYangSchemaResolvesLeafrefByFollowingTarget(t *testing.T) {
	schema := loadTestSchema(t)
	direct, ok := schema.lookup(testInterfaceLeafPath("oper-status"))
	require.True(t, ok)

	viaRef, ok := schema.lookup(testInterfaceLeafPath("oper-status-ref"))
	require.True(t, ok)
	assert.Equal(t, direct.MetricConfig, viaRef.MetricConfig)
	assert.Equal(t, direct.kind, viaRef.kind)
}

func TestLoadYangSchemaUnionBailsToConfig(t *testing.T) {
	schema := loadTestSchema(t)
	_, ok := schema.lookup([]string{"interfaces", "interface", "state", "counters", "raw-id"})
	assert.False(t, ok)
}

func TestLoadYangSchemaChoiceCaseIsTransparent(t *testing.T) {
	schema := loadTestSchema(t)

	rm, ok := schema.lookup(testInterfaceLeafPath("speed"))
	require.True(t, ok, "leaf under a choice/case pair should resolve at its gNMI-visible path, skipping the choice/case names")
	assert.Equal(t, valueKindInt, rm.kind)

	_, ok = schema.lookup(testInterfaceLeafPath("auto-negotiate"))
	require.True(t, ok)
}

func TestLoadYangSchemaLookupStripsModulePrefixes(t *testing.T) {
	schema := loadTestSchema(t)
	rm, ok := schema.lookup([]string{"ti:interfaces", "interface", "ti:state", "ti:counters", "in-octets"})
	require.True(t, ok)
	assert.Equal(t, metricTypeSum, rm.Type)
}

func TestLoadYangSchemaUnknownLeafIsUnresolved(t *testing.T) {
	schema := loadTestSchema(t)
	_, ok := schema.lookup(testInterfaceLeafPath("does-not-exist"))
	assert.False(t, ok)
}

func TestIsCounterType(t *testing.T) {
	tests := []struct {
		typ  *yang.YangType
		name string
		want bool
	}{
		{name: "direct counter64 name", typ: &yang.YangType{Name: "counter64"}, want: true},
		{name: "direct counter32 name", typ: &yang.YangType{Name: "counter32"}, want: true},
		{name: "plain uint64", typ: &yang.YangType{Name: "uint64"}, want: false},
		{
			name: "chained typedef falls back to root name",
			typ:  &yang.YangType{Name: "my-octets", Root: &yang.YangType{Name: "counter64"}},
			want: true,
		},
		{
			name: "chained typedef with unrelated root",
			typ:  &yang.YangType{Name: "my-octets", Root: &yang.YangType{Name: "uint64"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isCounterType(tt.typ))
		})
	}
}

func TestResolveLeafrefTargetRelativePath(t *testing.T) {
	index := map[string]*yang.Entry{
		"interfaces/interface/state/oper-status": {Name: "oper-status"},
	}
	target := resolveLeafrefTarget(
		"../oper-status",
		[]string{"interfaces", "interface", "state", "oper-status-ref"},
		index,
	)
	require.NotNil(t, target)
	assert.Equal(t, "oper-status", target.Name)
}

func TestResolveLeafrefTargetAbsolutePathStripsPrefixes(t *testing.T) {
	index := map[string]*yang.Entry{
		"interfaces/interface/state/oper-status": {Name: "oper-status"},
	}
	target := resolveLeafrefTarget(
		"/ti:interfaces/interface/ti:state/oper-status",
		nil,
		index,
	)
	require.NotNil(t, target)
	assert.Equal(t, "oper-status", target.Name)
}

func TestResolveLeafrefTargetUnresolvedReturnsNil(t *testing.T) {
	index := map[string]*yang.Entry{}
	assert.Nil(t, resolveLeafrefTarget("../nope", []string{"a", "b"}, index))
	assert.Nil(t, resolveLeafrefTarget("", []string{"a", "b"}, index))
}
