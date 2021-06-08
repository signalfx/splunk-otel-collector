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
	"log"
)

// expandSA takes an unmarshalled Smart Agent config struct and returns a config
// with any SA #from directives translated into their Otel equivalent.
func expandSA(in interface{}, wd string) (interface{}, bool) {
	switch t := in.(type) {
	case []interface{}:
		return expandSlice(t, wd)
	case map[interface{}]interface{}:
		return expandMap(t, wd)
	default:
		return in, false
	}
}

func expandSlice(l []interface{}, wd string) (interface{}, bool) {
	var out []interface{}
	for _, v := range l {
		next, flatten := expandSA(v, wd)
		if flatten {
			if a, ok := next.([]interface{}); ok {
				out = append(out, a...)
			}
		} else {
			out = append(out, next)
		}
	}
	return out, false
}

func expandMap(m map[interface{}]interface{}, wd string) (interface{}, bool) {
	d, isDirective, err := parseDirective(m, wd)
	if err != nil {
		log.Fatalf("parseDirective failed: %v: error %v", m, err)
	}
	if isDirective {
		rendered, err := d.render()
		// TODO return the error
		if err == nil {
			return rendered, d.flatten
		}
		log.Fatalf("render failed: %v: error: %v", m, err)
	}
	out := map[interface{}]interface{}{}
	for k, v := range m {
		expanded, flatten := expandSA(v, wd)
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
	return out, false
}
