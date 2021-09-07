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
	otelIDCommand     = "command"
	otelIDProcessName = "process_name"
	otelIDPort        = "port"
	otelIDHasPort     = "(port != nil)"
	otelIDIsIPv6      = "is_ipv6"
	otelIDTransport   = "transport"
	otelIDType        = "type"
)

// maps SA identifiers to Otel identifiers
var identifierMap = map[string]string{
	"command":       otelIDCommand,
	"has_port":      otelIDHasPort,
	"hasPort":       otelIDHasPort,
	"ip_address":    "",
	"ipAddress":     "",
	"is_ipv6":       otelIDIsIPv6,
	"isIpv6":        otelIDIsIPv6,
	"network_port":  "",
	"networkPort":   "",
	"discovered_by": "",
	"discoveredBy":  "",
	"host":          "",
	"id":            "",
	"name":          otelIDProcessName,
	"port":          otelIDPort,
	"port_type":     otelIDTransport,
	"portType":      otelIDTransport,
	"target":        otelIDType,
}

func discoveryRuleToRCRule(dr string, observers []interface{}) (rcRule string, err error) {
	dr = strings.ReplaceAll(dr, "=~", " matches ")
	tree, err := parser.Parse(dr)
	if err != nil {
		return "", err
	}
	err = translateTree(tree.Node)
	if err != nil {
		return "", err
	}

	rcRule = treeToString(tree.Node)
	if !hasType(tree.Node) {
		var typ string
		typ, err = guessType(observers)
		rcRule = fmt.Sprintf("type == %q && %s", typ, rcRule)
	}
	return rcRule, err
}

func translateTree(node ast.Node) error {
	switch n := node.(type) {
	case *ast.BinaryNode:
		if err := translateTree(n.Left); err != nil {
			return err
		}
		if err := translateTree(n.Right); err != nil {
			return err
		}
	case *ast.MatchesNode:
		if err := translateTree(n.Left); err != nil {
			return err
		}
		if err := translateTree(n.Right); err != nil {
			return err
		}
	case *ast.IdentifierNode:
		if otelID, ok := identifierMap[n.Value]; ok {
			if otelID == "" {
				return fmt.Errorf("no translation for identifier: %q", n.Value)
			}
			n.Value = otelID
		}
	}
	return nil
}

func hasType(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.BinaryNode:
		if hasType(n.Left) || hasType(n.Right) {
			return true
		}
	case *ast.IdentifierNode:
		if n.Value == "type" {
			return true
		}
	}
	return false
}

func guessType(observers []interface{}) (string, error) {
	if observers == nil {
		return "", errors.New("unable to guess type; no observers")
	}
	if len(observers) != 1 {
		return "", fmt.Errorf("unable to guess type; too many observers: %v", observers)
	}
	obs := observers[0].(map[interface{}]interface{})
	if obs["type"] == "host" {
		return "hostport", nil
	}
	return "", fmt.Errorf("unable to guess type from observers: %v", observers)
}

func treeToString(node ast.Node) string {
	switch n := node.(type) {
	case *ast.BinaryNode:
		return fmt.Sprintf("%s %s %s", treeToString(n.Left), n.Operator, treeToString(n.Right))
	case *ast.MatchesNode:
		return fmt.Sprintf("%s matches %s", treeToString(n.Left), treeToString(n.Right))
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
