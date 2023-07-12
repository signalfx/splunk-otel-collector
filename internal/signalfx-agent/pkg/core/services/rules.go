package services

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// get returns the value of the specified key in the supplied map
func get(args ...interface{}) interface{} {
	if len(args) < 2 {
		panic("get takes at least 2 args")
	}
	inputMap := args[0]
	key := args[1]

	var defVal interface{}
	if len(args) == 3 {
		defVal = args[2]
	}

	mapVal := reflect.ValueOf(inputMap)
	if mapVal.Kind() != reflect.Map {
		panic("first arg to Get must be a map")
	}

	keyVal := reflect.ValueOf(key)

	if val := mapVal.MapIndex(keyVal); val.IsValid() && val.CanInterface() {
		return val.Interface()
	} else if defVal != nil {
		return defVal
	}

	return nil
}

// Kept for backwards-compatibility, they aren't really necessary for newly
// written rules.
var ruleFunctions = map[string]interface{}{
	"Get": get,
	"Contains": func(args ...interface{}) bool {
		val := get(args...)
		return val != nil
	},
	"ToString": func(val interface{}) string {
		return fmt.Sprintf("%v", val)
	},
	"Sprintf": fmt.Sprintf,
	"Getenv":  os.Getenv,
}

func parseRuleText(text string) (*parser.Tree, error) {
	return parser.Parse(text)
}

func preprocessRuleText(text string) string {
	out := strings.ReplaceAll(text, " =~ ", " matches ")
	return strings.ReplaceAll(out, "=~", " matches ")
}

// EvaluateRule executes a govaluate expression against an endpoint
func EvaluateRule(si Endpoint, originalText string, errorOnMissing bool, doValidation bool) (interface{}, error) {
	asMap := utils.DuplicateInterfaceMapKeysAsCamelCase(EndpointAsMap(si))
	missing, err := findMissingIdentifiers(originalText, asMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discovery rule: %v", err)
	}

	// If there are missing vars but errorOnMissing is true, then the
	// processing will continue to expr.Run below, which will generate an error
	// due to the missing identifier.
	if len(missing) > 0 && !errorOnMissing {
		if doValidation {
			log.WithField("discoveryRule", originalText).Warnf("Discovery rule contains unknown variables: %v", missing)
		}
		return nil, nil
	}

	log.WithFields(log.Fields{
		"ruleText": originalText,
		"asMap":    spew.Sdump(asMap),
	}).Debug("Evaluating rule")

	return ExecuteRule(originalText, asMap)
}

func ExecuteRule(originalText string, variables map[string]interface{}) (interface{}, error) {
	env := utils.MergeInterfaceMaps(ruleFunctions, variables)
	ruleProg, err := expr.Compile(preprocessRuleText(originalText), expr.Env(env))
	if err != nil {
		return nil, err
	}

	return expr.Run(ruleProg, env)
}

// DoesServiceMatchRule returns true if service endpoint satisfies the rule
// given
func DoesServiceMatchRule(si Endpoint, ruleText string, doValidation bool) bool {
	ret, err := EvaluateRule(si, ruleText, false, doValidation)
	if err != nil {
		log.WithFields(log.Fields{
			"discoveryRule":   ruleText,
			"serviceInstance": spew.Sdump(si),
			"error":           err,
		}).Error("Could not evaluate discovery rule")
		return false
	}

	if ret == nil {
		return false
	}
	exprVal, ok := ret.(bool)
	if !ok {
		log.WithFields(log.Fields{
			"discoveryRule":   ruleText,
			"serviceInstance": spew.Sdump(si),
		}).Errorf("Discovery rule did not evaluate to a true/false value")
		return false
	}

	return exprVal
}

// ValidateDiscoveryRule takes a discovery rule string and returns false if it
// can be determined to be invalid.  It does not guarantee validity but can be
// used to give upfront feedback to the user if there are syntax errors in the
// rule.
func ValidateDiscoveryRule(originalRule string) error {
	ruleText := preprocessRuleText(originalRule)
	if _, err := parseRuleText(ruleText); err != nil {
		return fmt.Errorf("syntax error in discovery rule '%s': %s", ruleText, err.Error())
	}
	return nil
}

var _ ast.Visitor = (*exprVisitor)(nil)

type exprVisitor struct {
	identifiers []string
}

func (v *exprVisitor) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.identifiers = append(v.identifiers, n.Value)
	}
}

func extractIdentifiers(ruleTree *parser.Tree) []string {
	visitor := &exprVisitor{}
	ast.Walk(&ruleTree.Node, visitor)
	return visitor.identifiers
}

func findMissingIdentifiers(originalText string, endpointParams map[string]interface{}) ([]string, error) {
	ruleText := preprocessRuleText(originalText)
	ruleTree, err := parseRuleText(ruleText)
	if err != nil {
		return nil, fmt.Errorf("could not parse rule: %v", err)
	}

	vars := extractIdentifiers(ruleTree)

	var missing []string
	for _, v := range vars {
		if _, ok := endpointParams[v]; !ok {
			if _, ok = ruleFunctions[v]; !ok {
				missing = append(missing, v)
			}
		}
	}
	return missing, nil
}
