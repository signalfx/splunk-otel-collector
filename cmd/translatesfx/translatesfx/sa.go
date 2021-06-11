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
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
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
	d, err := parseDirective(m)
	if err != nil {
		log.Fatalf("parseDirective failed: %v: error %v", m, err)
	}
	if d.isDirective {
		replacement, err := processDirective(d, wd)
		if err == nil {
			return replacement, d.flatten
		}
		log.Fatalf("processDirective failed: %v: error: %v", m, err)
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

func processDirective(d directive, wd string) (interface{}, error) {
	expanded, err := expandFromSource(d, wd)
	if err != nil {
		return nil, err
	}
	return expanded, err
}

func expandFromSource(d directive, wd string) (interface{}, error) {
	switch d.fromType {
	case directiveSourceFile:
		return expandFiles(d, wd)
	case directiveSourceEnv:
		return expandEnv(d)
	case directiveSourceUnknown:
		return nil, fmt.Errorf("#from fromType type unknown: %v", d.fromType)
	default:
		return nil, fmt.Errorf("#from fromType type not supported by translatesfx at this time: %v", d.fromType)
	}
}

func expandEnv(d directive) (interface{}, error) {
	return fmt.Sprintf("${%s}", d.fromPath), nil
}

func directiveSource(from string) string {
	idx := strings.Index(from, ":")
	if idx == -1 {
		return ""
	}
	return from[:idx]
}

func expandFiles(d directive, wd string) (interface{}, error) {
	fromFullpath := d.fromPath
	if d.fromPath[:1] != string(os.PathSeparator) {
		fromFullpath = filepath.Join(wd, d.fromPath)
	}
	paths, err := filepath.Glob(fromFullpath)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		if d.defaultV != "" {
			return d.defaultV, nil
		}

		if !d.optional {
			return nil, fmt.Errorf("#from files not found and directive not marked optional: %v", d)
		}
	}

	var items []interface{}
	for _, path := range paths {
		unmarshaled, err := unmarshal(path)
		if err != nil {
			return nil, err
		}
		items = append(items, unmarshaled)
	}
	return merge(items)
}

func unmarshal(path string) (interface{}, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var replacement interface{}
	err = yaml.UnmarshalStrict(bytes, &replacement)
	if err != nil {
		return nil, err
	}
	return replacement, nil
}

func merge(items []interface{}) (interface{}, error) {
	switch len(items) {
	case 0:
		return nil, nil
	case 1:
		return items[0], nil
	}
	switch items[0].(type) {
	case []interface{}:
		return mergeSlices(items)
	case map[interface{}]interface{}:
		return mergeMaps(items)
	}
	return nil, fmt.Errorf("unable to merge: %v", items)
}

func mergeSlices(items []interface{}) (interface{}, error) {
	var out []interface{}
	for _, item := range items {
		l, ok := item.([]interface{})
		if !ok {
			return nil, fmt.Errorf("mergeSlices: type coersion failed for item %v in items %v", item, items)
		}
		out = append(out, l...)
	}
	return out, nil
}

func mergeMaps(items []interface{}) (interface{}, error) {
	out := map[interface{}]interface{}{}
	for _, item := range items {
		m, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("mergeMaps: type coersion failed for item %v in items %v", item, items)
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out, nil
}
