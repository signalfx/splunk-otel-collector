// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestExpand_Map(t *testing.T) {
	yml := `myMap: {"#from": "testdata/map.yaml"}`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _ := expandSA(v, "")
	require.NoError(t, err)
	expected := `myMap:
  baz: glarch
  foo: bar
`
	assert.Equal(t, expected, toYaml(t, out))
}

func TestExpand_List(t *testing.T) {
	yml := `myList: [{"#from": "testdata/map.yaml"}]`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _ := expandSA(v, "")
	expandedYaml, err := yaml.Marshal(out)
	require.NoError(t, err)
	expected := `myList:
- baz: glarch
  foo: bar
`
	assert.Equal(t, expected, string(expandedYaml))
}

func TestExpand_FlattenSlice(t *testing.T) {
	yml := `list:
  - name: foo
  - {"#from": "testdata/list1.yaml", flatten: true}
  - {"#from": "testdata/list2.yaml", flatten: true}
`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _ := expandSA(v, "")
	expandedYaml, err := yaml.Marshal(out)
	require.NoError(t, err)
	expected := `list:
- name: foo
- name: my-db1
- name: my-db2
- name: my-db3
- name: my-db4
`
	assert.Equal(t, expected, string(expandedYaml))
}

func TestExpand_FlattenMap(t *testing.T) {
	yml := `map:
  message: hello
  xxx: {"#from": "testdata/map.yaml", flatten: true}
`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	expanded, _ := expandSA(v, "")
	require.NoError(t, err)
	expected := map[interface{}]interface{}{
		"map": map[interface{}]interface{}{
			"message": "hello",
			"foo":     "bar",
			"baz":     "glarch",
		},
	}
	assert.Equal(t, expected, expanded)
}

func TestExpand_SAComplex(t *testing.T) {
	yml, err := ioutil.ReadFile("testdata/sa-complex.yaml")
	require.NoError(t, err)
	var v interface{}
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	expanded, _ := expandSA(v, "")
	m := expanded.(map[interface{}]interface{})
	assert.Equal(t, "https://api.us1.signalfx.com", m["apiUrl"])
	monitors := m["monitors"].([]interface{})
	cpuFound, loadFound := false, false
	for _, monitor := range monitors {
		monMap := monitor.(map[interface{}]interface{})
		monType := monMap["type"]
		if monType == "cpu" {
			cpuFound = true
		} else if monType == "load" {
			loadFound = true
		}
	}
	assert.True(t, cpuFound)
	assert.True(t, loadFound)
}

func TestProcessDirective_Simple(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/token",
	})
	require.NoError(t, err)
	v, err := processDirective(d, "")
	require.NoError(t, err)
	assert.Equal(t, "abc123", v.(string))
}

func TestProcessDirective_Map(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/map.yaml",
	})
	require.NoError(t, err)
	v, err := processDirective(d, "")
	require.NoError(t, err)
	assert.Equal(t, map[interface{}]interface{}{
		"foo": "bar",
		"baz": "glarch",
	}, v)
}

func TestProcessDirective_Wildcard(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/cfgs/map*.yaml",
	})
	require.NoError(t, err)
	v, err := processDirective(d, "")
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestProcessDirective_MissingNonOptionalFile(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/missing",
	})
	require.NoError(t, err)
	_, err = processDirective(d, "")
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "#from files not found"))
}

func TestProcessDirective_Default(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from":   "testdata/missing",
		"default": "foo",
	})
	require.NoError(t, err)
	v, err := processDirective(d, "")
	require.NoError(t, err)
	require.Equal(t, "foo", v)
}

func TestGetSource(t *testing.T) {
	assert.Equal(t, "env", directiveSource("env:SIGNALFX_ACCESS_TOKEN"))
	assert.Equal(t, "", directiveSource("foo"))
}

func TestMerge_Slices(t *testing.T) {
	merged, err := merge([]interface{}{
		[]interface{}{"aaa"}, []interface{}{"bbb"},
	})
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"aaa", "bbb"}, merged)
}

func TestMerge_Maps(t *testing.T) {
	m1 := map[interface{}]interface{}{"aaa": 111}
	m2 := map[interface{}]interface{}{"bbb": 222}
	merged, err := merge([]interface{}{m1, m2})
	require.NoError(t, err)
	assert.Equal(t, map[interface{}]interface{}{
		"aaa": 111,
		"bbb": 222,
	}, merged)
}

func TestMerge_Empty(t *testing.T) {
	merged, err := merge(nil)
	require.NoError(t, err)
	require.Empty(t, merged)
}

func toYaml(t *testing.T, out interface{}) string {
	expandedYaml, err := yaml.Marshal(out)
	require.NoError(t, err)
	return string(expandedYaml)
}
