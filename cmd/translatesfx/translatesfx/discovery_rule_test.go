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

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeToString(t *testing.T) {
	expression := "container_image matches \"redis\" && port == 6379"
	tree, err := parser.Parse(expression)
	require.NoError(t, err)
	assert.Equal(t, expression, treeToString(tree.Node))
}

func TestDiscoveryRuleToRCRule_TargetAndName(t *testing.T) {
	saRule := `target == "hostport" && name == "etcd"`
	rcRule, err := discoveryRuleToRCRule(saRule, nil)
	require.NoError(t, err)
	assert.Equal(t, `type == "hostport" && process_name == "etcd"`, rcRule)
}

func TestDiscoveryRuleToRCRule_HasPort(t *testing.T) {
	saRule := `target == "hostport" && has_port == true`
	rcRule, err := discoveryRuleToRCRule(saRule, nil)
	require.NoError(t, err)
	program, err := expr.Compile(rcRule)
	require.NoError(t, err)
	output, err := expr.Run(program, map[string]interface{}{
		"port": nil,
	})
	require.NoError(t, err)
	assert.False(t, output.(bool))
	output, err = expr.Run(program, map[string]interface{}{
		"type": "hostport",
		"port": 1234,
	})
	require.NoError(t, err)
	assert.True(t, output.(bool))
}

func TestDiscoveryRuleToRCRule_NoTranslation(t *testing.T) {
	saRule := `host == "127.0.0.1"`
	_, err := discoveryRuleToRCRule(saRule, nil)
	require.EqualError(t, err, `no translation for identifier: "host"`)
}

func TestDiscoveryRuleToRCRule_CamelCase(t *testing.T) {
	camel, err := discoveryRuleToRCRule(`target == "hostport" && portType == "TCP"`, nil)
	require.NoError(t, err)
	snake, err := discoveryRuleToRCRule(`target == "hostport" && port_type == "TCP"`, nil)
	require.NoError(t, err)
	assert.Equal(t, snake, camel)
}

func TestHasType(t *testing.T) {
	tree, err := parser.Parse(`type == "hostport" && process_name == "redis-server"`)
	require.NoError(t, err)
	assert.True(t, hasType(tree.Node))

	tree, err = parser.Parse(`process_name == "redis-server"`)
	require.NoError(t, err)
	assert.False(t, hasType(tree.Node))
}

func TestGuessType(t *testing.T) {
	typ, _ := guessType([]interface{}{map[interface{}]interface{}{
		"type": "host",
	}})
	assert.Equal(t, "hostport", typ)
}
