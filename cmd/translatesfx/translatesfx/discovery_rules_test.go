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
	output, err := expr.Run(program, map[string]any{
		"port": nil,
	})
	require.NoError(t, err)
	assert.False(t, output.(bool))
	output, err = expr.Run(program, map[string]any{
		"type": "hostport",
		"port": 1234,
	})
	require.NoError(t, err)
	assert.True(t, output.(bool))
}

func TestDiscoveryRuleToRCRule_NoObservers(t *testing.T) {
	saRule := `host == "127.0.0.1"`
	_, err := discoveryRuleToRCRule(saRule, nil)
	require.EqualError(t, err, `discoveryRuleToRCRule failed: target not found: guessRuleType: no observers`)
}

func TestDiscoveryRuleToRCRule_CamelCase(t *testing.T) {
	camel, err := discoveryRuleToRCRule(`target == "hostport" && portType == "TCP"`, nil)
	require.NoError(t, err)
	snake, err := discoveryRuleToRCRule(`target == "hostport" && port_type == "TCP"`, nil)
	require.NoError(t, err)
	assert.Equal(t, snake, camel)
}

func TestDiscoveryRuleToRCRule_K8s(t *testing.T) {
	otel, err := discoveryRuleToRCRule(`target == "pod" && kubernetes_pod_name == "redis" && kubernetes_namespace == "default" && kubernetes_pod_uid == "abc123"`, nil)
	require.NoError(t, err)
	assert.Equal(t, `type == "port" && pod.name == "redis" && pod.namespace == "default" && pod.uid == "abc123"`, otel)
}

func TestGuessType(t *testing.T) {
	typ, _ := guessRuleType([]any{map[any]any{
		"type": "host",
	}})
	assert.Equal(t, "hostport", typ)
}

func TestUpdateTarget_HostPort(t *testing.T) {
	const orig = `target == "hostport" && name == "redis"`
	tree, err := parser.Parse(orig)
	require.NoError(t, err)
	_, found := updateTarget(tree.Node)
	assert.True(t, found)
	assert.Equal(t, orig, treeToString(tree.Node))
}

func TestUpdateTarget_Pod(t *testing.T) {
	tree, err := parser.Parse(`name == "redis" && target == "pod"`)
	require.NoError(t, err)
	_, found := updateTarget(tree.Node)
	assert.True(t, found)
	assert.Equal(t, `name == "redis" && target == "port"`, treeToString(tree.Node))
}

func TestUpdateTarget_MissingTarget(t *testing.T) {
	tree, err := parser.Parse(`name == "redis"`)
	require.NoError(t, err)
	_, found := updateTarget(tree.Node)
	assert.False(t, found)
}
