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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestExpandSA_Map(t *testing.T) {
	yml := `myMap: {"#from": "testdata/map.yaml", "default": "foo"}`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _, _ := expand(v, "", yamlPath{})
	require.NoError(t, err)
	expected := `myMap:
  baz: glarch
  foo: bar
`
	assert.Equal(t, expected, toYaml(t, out))
}

func TestExpandSA_List(t *testing.T) {
	yml := `myList: [{"#from": "testdata/map.yaml", "default": "foo"}]`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _, _ := expand(v, "", yamlPath{})
	expandedYaml, err := yaml.Marshal(out)
	require.NoError(t, err)
	expected := `myList:
- baz: glarch
  foo: bar
`
	assert.Equal(t, expected, string(expandedYaml))
}

func TestExpandSA_FlattenSlice(t *testing.T) {
	yml := `list:
  - name: foo
  - {"#from": "testdata/list1.yaml", flatten: true}
  - {"#from": "testdata/list2.yaml", flatten: true}
`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	out, _, _ := expand(v, "", yamlPath{})
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

func TestExpandSA_FlattenMap(t *testing.T) {
	yml := `map:
  message: hello
  xxx: {"#from": "testdata/map.yaml", flatten: true}
`
	var v interface{}
	err := yaml.UnmarshalStrict([]byte(yml), &v)
	require.NoError(t, err)
	expanded, _, _ := expand(v, "", yamlPath{})
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

func TestExpandSA_Complex(t *testing.T) {
	yml, err := ioutil.ReadFile("testdata/sa-complex.yaml")
	require.NoError(t, err)
	var v interface{}
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	expanded, _, err := expandSA(v, "")
	require.NoError(t, err)
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

func TestMultiMonitors(t *testing.T) {
	yml, err := ioutil.ReadFile("testdata/sa-multimonitors.yaml")
	require.NoError(t, err)
	var v interface{}
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	expanded, _, err := expandSA(v, "")
	require.NoError(t, err)
	m := expanded.(map[interface{}]interface{})
	assert.Equal(t, 2, len(m["monitors"].([]interface{})))
}

func TestYamlPath(t *testing.T) {
	yp0 := yamlPath{
		forcePaths: []string{"/foo"},
	}
	yp1 := yp0.key("aaa")
	assert.Equal(t, "/aaa", yp1.curr)
	yp2 := yp1.index(0)
	assert.Equal(t, "/aaa/0", yp2.curr)
	yp3 := yp2.key("xyz")
	assert.Equal(t, "/aaa/0/xyz", yp3.curr)
	assert.False(t, yp1.forceExpand())

	yp1 = yp0.key("foo")
	assert.Equal(t, "/foo", yp1.curr)
	assert.True(t, yp1.forceExpand())
	yp2 = yp1.key("bar")
	assert.True(t, yp2.forceExpand())
}

func toYaml(t *testing.T, v interface{}) string {
	expandedYaml, err := yaml.Marshal(v)
	require.NoError(t, err)
	return string(expandedYaml)
}
