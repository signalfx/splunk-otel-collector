// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import "gopkg.in/ini.v1"

type Transform struct {
	Name   string
	Regex  string
	Format string
}

// MergeTransforms merges multiple slices of transforms, with later slices
// taking precedence. Stanzas are keyed by name; the last definition wins.
func MergeTransforms(layers [][]Transform) []Transform {
	seen := make(map[string]int)
	var result []Transform
	for _, layer := range layers {
		for _, t := range layer {
			if idx, ok := seen[t.Name]; ok {
				result[idx] = t
			} else {
				seen[t.Name] = len(result)
				result = append(result, t)
			}
		}
	}
	return result
}

func ReadTransforms(payload []byte) ([]Transform, error) {
	f, err := ini.Load(payload)
	if err != nil {
		return nil, err
	}
	result := make([]Transform, len(f.Sections())-1)
	s := 0
	for _, section := range f.Sections() {
		if section.Name() == ini.DefaultSection {
			continue // disregard default section. We need a stanza per transform.
		}
		t := Transform{
			Name:   section.Name(),
			Regex:  section.Key("REGEX").String(),
			Format: section.Key("FORMAT").String(),
		}

		result[s] = t
		s++
	}

	return result, nil
}
