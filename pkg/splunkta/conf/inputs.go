// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"encoding/xml"

	"gopkg.in/ini.v1"
)

const (
	appName        = "tarunner"
	xmlDeclaration = `<?xml version="1.0" encoding="UTF-8"?>
`
)

type Input struct {
	ServerHost    string        `xml:"server_host"`
	ServerURI     string        `xml:"server_uri"`
	SessionKey    string        `xml:"session_key"`
	CheckpointDir string        `xml:"checkpoint_dir"`
	AppDir        string        `xml:"-"`
	Configuration Configuration `xml:"configuration"`
}

// IsDisabled reports whether the stanza has disabled=1 or disabled=true.
func (s *Stanza) IsDisabled() bool {
	p := s.Params.Get("disabled")
	return p != nil && (p.Value == "1" || p.Value == "true")
}

func ReadInput(payload []byte, appDir string) ([]Input, error) {
	f, err := ini.Load(payload)
	if err != nil {
		return nil, err
	}
	result := make([]Input, len(f.Sections())-1)
	s := 0
	for _, section := range f.Sections() {
		if section.Name() == ini.DefaultSection {
			continue // disregard default section. We need a stanza per input.
		}
		i := Input{
			AppDir: appDir,
			Configuration: Configuration{
				Stanza: Stanza{
					Name:   section.Name(),
					App:    appName,
					Params: make([]Param, len(section.Keys())),
				},
			},
		}

		for keyIndex, key := range section.Keys() {
			i.Configuration.Stanza.Params[keyIndex] = Param{
				Name:  key.Name(),
				Value: key.Value(),
			}
		}

		result[s] = i
		s++
	}

	return result, nil
}

// MergeInputs merges layered inputs; later layers take precedence per param key.
func MergeInputs(layers [][]Input) []Input {
	seen := make(map[string]int)
	var result []Input
	for _, layer := range layers {
		for _, input := range layer {
			name := input.Configuration.Stanza.Name
			if idx, ok := seen[name]; ok {
				result[idx] = mergeInput(result[idx], input)
			} else {
				seen[name] = len(result)
				result = append(result, input)
			}
		}
	}
	return result
}

func mergeInput(base, override Input) Input {
	merged := base
	params := make(map[string]int, len(base.Configuration.Stanza.Params))
	mergedParams := append([]Param{}, base.Configuration.Stanza.Params...)
	for i, p := range mergedParams {
		params[p.Name] = i
	}
	for _, p := range override.Configuration.Stanza.Params {
		if idx, ok := params[p.Name]; ok {
			mergedParams[idx] = p
		} else {
			params[p.Name] = len(mergedParams)
			mergedParams = append(mergedParams, p)
		}
	}
	merged.Configuration.Stanza.Params = mergedParams
	return merged
}

func (i *Input) ToXML() ([]byte, error) {
	b, err := xml.MarshalIndent(i, "", "  ")
	return append([]byte(xmlDeclaration), b...), err
}
