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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
)

const (
	otelRuleType         = "type"
	otelRuleCommand      = "command"
	otelRuleProcessName  = "process_name"
	otelRulePort         = "port"
	otelRuleHasPort      = "(port != nil)"
	otelRuleIsIPv6       = "is_ipv6"
	otelRuleTransport    = "transport"
	otelRulePodName      = "pod.name"
	otelRulePodNamespace = "pod.namespace"
	otelRulePodUID       = "pod.uid"
)

// maps SA identifiers to Otel identifiers for hostport rules
var hostportIDs = map[string]string{
	"target":        otelRuleType,
	"command":       otelRuleCommand,
	"has_port":      otelRuleHasPort,
	"hasPort":       otelRuleHasPort,
	"ip_address":    "",
	"ipAddress":     "",
	"is_ipv6":       otelRuleIsIPv6,
	"isIpv6":        otelRuleIsIPv6,
	"network_port":  otelRulePort,
	"networkPort":   "",
	"discovered_by": "",
	"discoveredBy":  "",
	"host":          "",
	"id":            "",
	"name":          otelRuleProcessName,
	"port":          otelRulePort,
	"port_type":     otelRuleTransport,
	"portType":      otelRuleTransport,
}

// maps SA identifiers to Otel identifiers for k8s (type=port) rules
var k8sIDs = map[string]string{
	"target":                 otelRuleType,
	"has_port":               otelRuleHasPort,
	"ip_address":             "",
	"container_name":         "",
	"kubernetes_annotations": "",
	"network_port":           otelRulePort,
	"node_addresses":         "",
	"node_metadata":          "",
	"node_spec":              "",
	"node_status":            "",
	"pod_metadata":           "",
	"pod_spec":               "",
	"private_port":           "",
	"public_port":            otelRulePort,
	"alternate_port":         "",
	"container_command":      "",
	"container_id":           "",
	"container_image":        "",
	"container_labels":       "",
	"container_names":        "",
	"container_state":        "",
	"discovered_by":          "",
	"discoveredBy":           "",
	"host":                   "",
	"id":                     "",
	"name":                   otelRulePodName,
	"port":                   otelRulePort,
	"port_type":              otelRuleTransport,
	"portType":               otelRuleTransport,
	"orchestrator":           "",
	"port_labels":            "",
	"container_spec_name":    "",
	"kubernetes_node":        "",
	"kubernetes_node_uid":    "",
	"kubernetesPodName":      otelRulePodName,
	"kubernetes_pod_name":    otelRulePodName,
	"kubernetesNamespace":    otelRulePodNamespace,
	"kubernetes_namespace":   otelRulePodNamespace,
	"kubernetesPodUid":       otelRulePodUID,
	"kubernetes_pod_uid":     otelRulePodUID,
}

func discoveryRuleToRCRule(dr string, observers []any) (rcRule string, err error) {
	dr = strings.ReplaceAll(dr, "=~", " matches ")

	tree, err := parser.Parse(dr)
	if err != nil {
		return "", fmt.Errorf("discoveryRuleToRCRule failed: error parsing discovery rule: %w", err)
	}

	target, targetFound := updateTarget(tree.Node)

	dr = treeToString(tree.Node)

	if !targetFound {
		target, err = guessRuleType(observers)
		if err != nil {
			return "", fmt.Errorf("discoveryRuleToRCRule failed: target not found: %w", err)
		}
		dr = fmt.Sprintf("target == %q && %s", target, dr)
	}

	tree, err = parser.Parse(dr)
	if err != nil {
		return "", err
	}

	err = translateTree(tree.Node, targetToIDMap(target))
	if err != nil {
		return "", err
	}

	return treeToString(tree.Node), nil
}

func targetToIDMap(target string) map[string]string {
	switch target {
	case "hostport":
		return hostportIDs
	case "port":
		return k8sIDs
	}
	return nil
}

// updateTarget finds the target == "foo" part of the tree if it exists and
// updates "foo" to the corresponding target (type) for otel.
func updateTarget(node ast.Node) (string, bool) {
	if n, ok := node.(*ast.BinaryNode); ok {
		if idn, ok := n.Left.(*ast.IdentifierNode); ok && idn.Value == "target" {
			if target, ok := n.Right.(*ast.StringNode); ok {
				target.Value = saTargetToOtel(target.Value)
				return target.Value, true
			}
		} else {
			if target, found := updateTarget(n.Left); found {
				return target, found
			}
			if target, found := updateTarget(n.Right); found {
				return target, found
			}
		}
	}
	return "", false
}

func saTargetToOtel(target string) string {
	if target == "pod" {
		return "port"
	}
	return target
}

func translateTree(node ast.Node, idMap map[string]string) error {
	switch n := node.(type) {
	case *ast.BinaryNode:
		if err := translateTree(n.Left, idMap); err != nil {
			return err
		}
		if err := translateTree(n.Right, idMap); err != nil {
			return err
		}
	case *ast.IdentifierNode:
		if otelID, ok := idMap[n.Value]; ok {
			if otelID == "" {
				return fmt.Errorf("translateTree: no translation for identifier: %q", n.Value)
			}
			n.Value = otelID
		}
	}
	return nil
}

// guessRuleType is for rules that don't have the required e.g. 'type == "foo"'.
// We guess the type is "hostport" if there is one `host` observer and "port"
// if there is one `k8s-api` observer.
func guessRuleType(observers []any) (string, error) {
	if observers == nil {
		return "", errors.New("guessRuleType: no observers")
	}
	if len(observers) != 1 {
		return "", fmt.Errorf("guessRuleType: too many observers: %v", observers)
	}
	obs := observers[0].(map[any]any)
	switch obs["type"] {
	case "host":
		return "hostport", nil
	case "k8s-api":
		return "port", nil
	}
	return "", fmt.Errorf("guessRuleType: unable to guess type from observers: %v", observers)
}

func treeToString(node ast.Node) string {
	switch n := node.(type) {
	case *ast.BinaryNode:
		return fmt.Sprintf("%s %s %s", treeToString(n.Left), n.Operator, treeToString(n.Right))
	case *ast.IdentifierNode:
		return n.Value
	case *ast.IntegerNode:
		return strconv.Itoa(n.Value)
	case *ast.StringNode:
		return fmt.Sprintf("%q", n.Value)
	case *ast.UnaryNode:
		return fmt.Sprintf("%s(%s)", n.Operator, treeToString(n.Node))
	case *ast.BoolNode:
		return strconv.FormatBool(n.Value)
	default:
		panic(fmt.Sprintf("unhandled node type: %v", n))
	}
}
