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
	d, err := parseDirective(map[interface{}]interface{}{})
	require.NoError(t, err)
	assert.False(t, d.isDirective)
}

func TestParseDirective(t *testing.T) {
	d, err := parseDirective(map[interface{}]interface{}{
		"#from":    "/some/path",
		"flatten":  true,
		"default":  "abc123",
		"optional": true,
	})
	require.NoError(t, err)
	require.True(t, d.isDirective)
	assert.Equal(t, "/some/path", d.fromPath)
	assert.True(t, d.flatten)
	assert.Equal(t, "abc123", d.defaultV)
	assert.True(t, d.optional)
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
