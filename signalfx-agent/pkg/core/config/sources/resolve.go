// Copyright  Splunk, Inc.
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

package sources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/yalp/jsonpath"

	yaml "gopkg.in/yaml.v2"
)

type resolveFunc func(v RawDynamicValueSpec) ([]interface{}, string, *dynamicValueSpec, error)

// The resolver is what aggregates together multiple source caches and converts
// raw dynamic value specs (e.g. the {"#from": ...} values) to the actual
// value.
type resolver struct {
	sources map[string]*configSourceCacher
}

func newResolver(sources map[string]*configSourceCacher) *resolver {
	return &resolver{
		sources: sources,
	}
}

func (r *resolver) Resolve(raw RawDynamicValueSpec) ([]interface{}, string, *dynamicValueSpec, error) {
	spec, err := parseRawSpec(raw)
	if err != nil {
		return nil, "", nil, err
	}

	sourceName := spec.From.SourceName()
	source, ok := r.sources[sourceName]
	if !ok {
		return nil, "", nil, fmt.Errorf("source '%s' is not configured", sourceName)
	}

	contentMap, err := source.Get(spec.From.Path(), spec.Optional)
	if err != nil {
		return nil, "", nil, errors.WithMessage(err,
			"could not resolve path "+spec.From.String())
	}

	var value []interface{}
	if len(contentMap) == 0 && spec.Default != nil {
		value = []interface{}{
			spec.Default,
		}
	} else {
		value, err = convertFileBytesToValues(contentMap, spec.Raw, spec.JSONPath)
	}

	return value, spec.From.Path(), spec, err
}

func convertFileBytesToValues(content map[string][]byte, raw bool, jsonPath string) ([]interface{}, error) {
	var out []interface{}
	for path := range content {
		if len(strings.TrimSpace(string(content[path]))) == 0 {
			continue
		}

		var v interface{}
		switch {
		case raw:
			v = string(content[path])
		case jsonPath != "":
			var data interface{}
			err := json.Unmarshal(content[path], &data)
			if err != nil {
				return nil, fmt.Errorf("could not parse value as JSON for jsonPath: %v", err)
			}

			v, err = jsonpath.Read(data, jsonPath)
			if err != nil {
				return nil, fmt.Errorf("could not resolve jsonPath value: %v", err)
			}
		default:
			err := yaml.Unmarshal(content[path], &v)
			if err != nil {
				return nil, errors.WithMessage(err, "deserialization error at path "+path)
			}
		}

		out = append(out, v)
	}
	return out, nil
}
