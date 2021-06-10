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
	"fmt"
	"strings"
)

// expand takes an unmarshalled Smart Agent config struct and returns a config
// with any SA #from directives translated into their Otel equivalent.
func expandSA(orig interface{}, wd string) (interface{}, bool, error) {
	return expand(orig, wd, yamlPath{
		// Prevent these two top-level SA config keys from getting translated into their
		// configsource equivalent. We need monitors expanded/inlined so we can
		// translate them, and apiURL is used to get the realm, then not used in the
		// final otel config.
		forcePaths: []string{"/monitors", "/apiUrl"},
	})
}

func expand(in interface{}, wd string, yp yamlPath) (interface{}, bool, error) {
	switch t := in.(type) {
	case []interface{}:
		return expandSlice(t, wd, yp)
	case map[interface{}]interface{}:
		return expandMap(t, wd, yp)
	default:
		return in, false, nil
	}
}

func expandSlice(l []interface{}, wd string, yp yamlPath) (interface{}, bool, error) {
	var out []interface{}
	for i, v := range l {
		next, flatten, err := expand(v, wd, yp.index(i))
		if err != nil {
			return nil, false, err
		}
		if flatten {
			if a, ok := next.([]interface{}); ok {
				out = append(out, a...)
			}
		} else {
			out = append(out, next)
		}
	}
	return out, false, nil
}

func expandMap(m map[interface{}]interface{}, wd string, yp yamlPath) (interface{}, bool, error) {
	d, isDirective, err := parseDirective(m, wd)
	if err != nil {
		return nil, false, err
	}
	if isDirective {
		rendered, err := d.render(yp.forceExpand())
		if err != nil {
			return nil, false, err
		}
		return rendered, d.flatten, nil
	}
	out := map[interface{}]interface{}{}
	for k, v := range m {
		expanded, flatten, _ := expand(v, wd, yp.key(k.(string)))
		if flatten {
			if flattened, ok := expanded.(map[interface{}]interface{}); ok {
				for fk, fv := range flattened {
					out[fk] = fv
				}
			}
		} else {
			out[k] = expanded
		}
	}
	return out, false, nil
}

// yamlPath keeps track of the current yaml path and stores forcePaths so
// callers can check whether the current or parent path is marked as
// force-expand.
type yamlPath struct {
	curr       string
	forcePaths []string
}

func (p yamlPath) index(i int) yamlPath {
	p.curr += fmt.Sprintf("/%d", i)
	return p
}

func (p yamlPath) key(k string) yamlPath {
	p.curr += fmt.Sprintf("/%s", k)
	return p
}

func (p yamlPath) forceExpand() bool {
	for _, forcePath := range p.forcePaths {
		// not bulletproof, but enough for our tiny usecase
		if strings.HasPrefix(p.curr, forcePath) {
			return true
		}
	}
	return false
}
