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

// The code is copied from
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/c07d1e622c59ed013e6a02f96a5cc556513263da/receiver/receivercreator/rules.go
// with minimal changes. Once the discovery receiver upstreamed, this code can be reused.

package discoveryreceiver

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/builtin"
	"github.com/expr-lang/expr/vm"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
)

// Rule wraps expr rule for later evaluation.
type Rule struct {
	program *vm.Program
	text    string
}

// UnmarshalText will unmarshal a field from text
func (r *Rule) UnmarshalText(text []byte) error {
	rule, err := newRule(string(text))
	if err != nil {
		return fmt.Errorf(`invalid rule "%s": %w`, text, err)
	}
	*r = rule
	return err
}

// MarshalText marshals the rule to text
func (r Rule) MarshalText() (text []byte, err error) {
	return []byte(r.String()), nil
}

// ruleRe is used to verify the rule starts type check.
var ruleRe = regexp.MustCompile(
	fmt.Sprintf(`^type\s*==\s*(%q|%q|%q|%q|%q|%q|%q)`, observer.PodType, observer.K8sServiceType, observer.PortType, observer.HostPortType, observer.ContainerType, observer.K8sNodeType, observer.PodContainerType),
)

// newRule creates a new rule instance.
func newRule(ruleStr string) (Rule, error) {
	if ruleStr == "" {
		return Rule{}, errors.New("rule cannot be empty")
	}
	if !ruleRe.MatchString(ruleStr) {
		// TODO: Try validating against bytecode instead.
		return Rule{}, errors.New("rule must specify type")
	}

	// TODO: Maybe use https://godoc.org/github.com/expr-lang/expr#Env in type checking
	// depending on type == specified.
	v, err := expr.Compile(
		ruleStr,
		// expr v1.14.1 introduced a `type` builtin whose implementation we relocate to `typeOf`
		// to avoid collision
		expr.DisableBuiltin("type"),
		expr.Function("typeOf", func(params ...any) (any, error) {
			return builtin.Type(params[0]), nil
		}, new(func(any) string)),
	)
	if err != nil {
		return Rule{}, err
	}
	return Rule{text: ruleStr, program: v}, nil
}

func mustNewRule(ruleStr string) Rule {
	rule, err := newRule(ruleStr)
	if err != nil {
		panic(err)
	}
	return rule
}

func (r *Rule) String() string {
	return r.text
}

// eval the rule against the given endpoint.
func (r *Rule) eval(env observer.EndpointEnv) (bool, error) {
	res, err := expr.Run(r.program, env)
	if err != nil {
		return false, err
	}
	if ret, ok := res.(bool); ok {
		return ret, nil
	}
	return false, errors.New("rule did not return a boolean")
}
