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

// expandSA takes an unmarshalled Smart Agent config struct and returns a config
// with any SA #from directives translated into their Otel equivalent.
func expandSA(orig interface{}, wd string) (map[interface{}]interface{}, []string, error) {
	var vaultPaths []string
	expanded, _, err := expand(orig, wd, yamlPath{
		// Prevent these three top-level SA config keys from getting translated into
		// their configsource equivalent. We need monitors expanded/inlined so we can
		// translate them, apiURL is used to get the realm, and globalDimensions is used
		// to create a metricstransform processor.
		forcePaths: []string{"/monitors", "/apiUrl", "/globalDimensions"},
	}, &vaultPaths)
	return expanded.(map[interface{}]interface{}), vaultPaths, err
}

func expand(in interface{}, wd string, yp yamlPath, vaultPaths *[]string) (interface{}, bool, error) {
	switch t := in.(type) {
	case []interface{}:
		return expandSlice(t, wd, yp, vaultPaths)
	case map[interface{}]interface{}:
		return expandMap(t, wd, yp, vaultPaths)
	default:
		return in, false, nil
	}
}

func expandSlice(l []interface{}, wd string, yp yamlPath, vaultPaths *[]string) (interface{}, bool, error) {
	var out []interface{}
	for i, v := range l {
		next, flatten, err := expand(v, wd, yp.index(i), vaultPaths)
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

func expandMap(m map[interface{}]interface{}, wd string, yp yamlPath, vaultPaths *[]string) (interface{}, bool, error) {
	d, isDirective, err := parseDirective(m, wd)
	if err != nil {
		return nil, false, err
	}
	if isDirective {
		rendered, err := d.render(yp.forceExpand(), vaultPaths)
		if err != nil {
			return nil, false, err
		}
		return rendered, d.flatten, nil
	}
	out := map[interface{}]interface{}{}
	for k, v := range m {
		expanded, flatten, _ := expand(v, wd, yp.key(k.(string)), vaultPaths)
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

// yamlPath keeps track of the current yaml path and uses forcePaths
// to determine whether to force expand a part of the config.
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
