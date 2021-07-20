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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDirective_Empty(t *testing.T) {
	_, isDirective, err := parseDirective(map[interface{}]interface{}{}, "")
	require.NoError(t, err)
	require.False(t, isDirective)
}

func TestParseFrom_Simple(t *testing.T) {
	source, target, err := parseFrom("/some/path")
	require.NoError(t, err)
	assert.Equal(t, directiveSourceFile, source)
	assert.Equal(t, "/some/path", target)
}

func TestParseFrom_Env(t *testing.T) {
	source, target, err := parseFrom("env:FOO")
	require.NoError(t, err)
	assert.Equal(t, directiveSourceEnv, source)
	assert.Equal(t, "FOO", target)
}

func TestParseFrom_Unknown(t *testing.T) {
	_, _, err := parseFrom("xyz:FOO")
	require.Error(t, err)
}

func TestParseDirective(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from":    "/some/path",
		"flatten":  true,
		"default":  "abc123",
		"optional": true,
	}, "")
	require.NoError(t, err)
	assert.Equal(t, "/some/path", d.fromPath)
	assert.True(t, d.flatten)
	assert.Equal(t, "abc123", d.defaultV)
	assert.True(t, d.optional)
}

func TestRender_ConfigSource(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/token",
	}, "")
	require.NoError(t, err)
	v, err := d.render(false)
	require.NoError(t, err)
	assert.Equal(t, "${include:testdata/token}", v.(string))
}

func TestRender_FileExpansion_Simple(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from":   "testdata/token",
		"default": "foo", // specify a default to force file expansion
	}, "")
	require.NoError(t, err)
	v, err := d.render(false)
	require.NoError(t, err)
	assert.Equal(t, "abc123", v.(string))
}

func TestRender_FileExpansion_Map(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from":   "testdata/map.yaml",
		"default": "foo", // specify a default to force file expansion
	}, "")
	require.NoError(t, err)
	v, err := d.render(false)
	require.NoError(t, err)
	assert.Equal(t, map[interface{}]interface{}{
		"foo": "bar",
		"baz": "glarch",
	}, v)
}

func TestRender_FileExpansion_Wildcard(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from": "testdata/cfgs/map*.yaml", // asterisk forces file expansion
	}, "")
	require.NoError(t, err)
	v, err := d.render(false)
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestRender_FileExpansion_MissingNonOptionalFile(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from":   "testdata/missing",
		"default": "foo", // specify a default to force file expansion
	}, "")
	require.NoError(t, err)
	rendered, err := d.render(false)
	require.NoError(t, err)
	assert.Equal(t, "foo", rendered)
}

func TestDirective_ZK(t *testing.T) {
	d, _, err := parseDirective(map[interface{}]interface{}{
		"#from": "zk:/foo/bar",
	}, "")
	require.NoError(t, err)
	rendered, err := d.render(false)
	require.NoError(t, err)
	assert.Equal(t, "${zookeeper:/foo/bar}", rendered)
}

func TestDirectiveSource(t *testing.T) {
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

func TestHandleFileDirective(t *testing.T) {
	d := directive{
		fromPath: "testdata/token",
		fromType: directiveSourceFile,
		flatten:  false,
		optional: false,
	}
	expanded, err := d.handleFileType(false)
	require.NoError(t, err)
	assert.Equal(t, "${include:testdata/token}", expanded)
}
