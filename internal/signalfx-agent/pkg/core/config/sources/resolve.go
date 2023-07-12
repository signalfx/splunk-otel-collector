package sources

import (
	"encoding/json"
	"fmt"
	"strings"

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
		return nil, "", nil, fmt.Errorf(
			"could not resolve path %s: %w", spec.From.String(), err)
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
				return nil, fmt.Errorf("deserialization error at path %s: %w", path, err)
			}
		}

		out = append(out, v)
	}
	return out, nil
}
