// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"errors"
	"sort"

	"gopkg.in/ini.v1"
)

// ErrNoHTTPOut is returned when outputs.conf contains no [httpout] stanza.
var ErrNoHTTPOut = errors.New("no [httpout] stanza found in outputs.conf")

// ErrNoOutputStanzas is returned when outputs.conf contains no output stanzas.
var ErrNoOutputStanzas = errors.New("no output stanzas found in outputs.conf")

// ConfMap is a parsed .conf file: stanza name -> key -> value.
type ConfMap map[string]map[string]string

// Output holds the settings from outputs.conf stanza.
type Output struct {
	Configuration Configuration
}

func ParseConf(payload []byte) (ConfMap, error) {
	f, err := ini.Load(payload)
	if err != nil {
		return nil, err
	}
	result := make(ConfMap)
	for _, section := range f.Sections() {
		name := section.Name()
		if name == ini.DefaultSection {
			continue
		}
		keys := make(map[string]string)
		for _, key := range section.Keys() {
			keys[key.Name()] = key.Value()
		}
		result[name] = keys
	}
	return result, nil
}

func ParseAndMergeConf(payloads [][]byte) (ConfMap, error) {
	var layers []ConfMap
	for _, b := range payloads {
		parsed, err := ParseConf(b)
		if err != nil {
			return nil, err
		}
		layers = append(layers, parsed)
	}
	return MergeConf(layers), nil
}

func MergeConf(layers []ConfMap) ConfMap {
	merged := make(ConfMap)
	for _, layer := range layers {
		for stanza, keys := range layer {
			if merged[stanza] == nil {
				merged[stanza] = make(map[string]string)
			}
			for k, v := range keys {
				merged[stanza][k] = v
			}
		}
	}
	return merged
}

// ReadOutputGroups parses an outputs.conf payload and returns every output
// stanza.
func ReadOutputGroups(payload []byte) ([]Output, error) {
	parsed, err := ParseConf(payload)
	if err != nil {
		return nil, err
	}
	return OutputGroups(parsed)
}

// OutputGroups converts merged outputs.conf stanzas to Output values.
func OutputGroups(merged ConfMap) ([]Output, error) {
	if len(merged) == 0 {
		return nil, ErrNoOutputStanzas
	}

	names := make([]string, 0, len(merged))
	for name := range merged {
		names = append(names, name)
	}
	sort.Strings(names)

	outputs := make([]Output, 0, len(names))
	for _, name := range names {
		outputs = append(outputs, readOutput(name, merged[name]))
	}
	return outputs, nil
}

// HTTPOut extracts the [httpout] stanza from a merged outputs.conf map.
func HTTPOut(merged ConfMap) (*Output, error) {
	keys, ok := merged["httpout"]
	if !ok {
		return nil, ErrNoHTTPOut
	}
	output := readOutput("httpout", keys)
	return &output, nil
}

func readOutput(name string, keys map[string]string) Output {
	return Output{
		Configuration: Configuration{
			Stanza: Stanza{
				Name:   name,
				App:    appName,
				Params: readOutputParams(keys),
			},
		},
	}
}

func readOutputParams(keys map[string]string) Params {
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	sort.Strings(names)

	params := make(Params, len(names))
	for i, name := range names {
		params[i] = Param{
			Name:  name,
			Value: keys[name],
		}
	}
	return params
}
