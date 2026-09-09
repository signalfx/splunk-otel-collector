// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"regexp"
	"sort"
	"strings"

	"gopkg.in/ini.v1"
)

type PropType int

const (
	Source = iota
	Host
	SourceType
	Default
)

var fieldAliasRegex = regexp.MustCompile(`(\w+) as (\w+)`)

type Prop struct {
	Name                  string
	TimePrefix            string
	TimeFormat            string
	DatetimeConfig        string
	Category              string
	SourceType            string
	FieldAliases          []FieldAlias
	Transforms            []PropsTransforms
	MaxTimestampLookAhead int
	NoBinaryCheck         bool
	ShouldLineMerge       bool
}

type FieldAlias struct {
	Name string
	From string
	To   string
}

type PropsTransforms struct {
	Class  string
	Stanza []string
}

func (p *Prop) Type() PropType {
	switch {
	case strings.HasPrefix(p.Name, "source::"):
		return Source
	case strings.HasPrefix(p.Name, "host::"):
		return Host
	case p.Name == "default":
		return Default
	default:
		return SourceType
	}
}

// MergeProps merges multiple slices of props, with later slices taking
// precedence. Stanzas are keyed by name; the last definition wins.
// The merged result is re-ordered by specificity.
func MergeProps(layers [][]Prop) []Prop {
	seen := make(map[string]int)
	var result []Prop
	for _, layer := range layers {
		for _, p := range layer {
			if idx, ok := seen[p.Name]; ok {
				result[idx] = p
			} else {
				seen[p.Name] = len(result)
				result = append(result, p)
			}
		}
	}
	orderProps(result)
	return result
}

func ReadProps(payload []byte) ([]Prop, error) {
	f, err := ini.Load(payload)
	if err != nil {
		return nil, err
	}
	result := make([]Prop, len(f.Sections())-1)
	s := 0
	for _, section := range f.Sections() {
		if section.Name() == ini.DefaultSection {
			continue // disregard default section. We need a stanza per prop.
		}
		maxTimestampLookAhead := 0
		if section.Key("MAX_TIMESTAMP_LOOKAHEAD").String() != "" {
			maxTimestampLookAhead, err = section.Key("MAX_TIMESTAMP_LOOKAHEAD").Int()
			if err != nil {
				return nil, err
			}
		}

		noBinaryCheck := false
		if section.Key("NO_BINARY_CHECK").String() != "" {
			noBinaryCheck, err = section.Key("NO_BINARY_CHECK").Bool()
			if err != nil {
				return nil, err
			}
		}
		shouldLineMerge := false
		if section.Key("SHOULD_LINEMERGE").String() != "" {
			shouldLineMerge, err = section.Key("SHOULD_LINEMERGE").Bool()
			if err != nil {
				return nil, err
			}
		}
		p := Prop{
			Name:                  section.Name(),
			TimePrefix:            section.Key("TIME_PREFIX").String(),
			TimeFormat:            section.Key("TIME_FORMAT").String(),
			MaxTimestampLookAhead: maxTimestampLookAhead,
			DatetimeConfig:        section.Key("DATETIME_CONFIG").String(),
			NoBinaryCheck:         noBinaryCheck,
			Category:              section.Key("category").String(),
			SourceType:            section.Key("sourcetype").String(),
			FieldAliases:          readFieldAliases(section),
			Transforms:            readPropsTransforms(section),
			ShouldLineMerge:       shouldLineMerge,
		}

		result[s] = p
		s++
	}

	orderProps(result)
	return result, nil
}

func readPropsTransforms(section *ini.Section) []PropsTransforms {
	var result []PropsTransforms
	for _, key := range section.Keys() {
		if strings.HasPrefix(key.Name(), "TRANSFORMS-") {
			values := strings.Split(key.Value(), ",")
			transformStanzas := make([]string, len(values))
			for i, value := range values {
				transformStanzas[i] = strings.TrimSpace(value)
			}
			result = append(result, PropsTransforms{
				Class:  key.Name(),
				Stanza: transformStanzas,
			})
		}
	}
	return result
}

func readFieldAliases(section *ini.Section) []FieldAlias {
	var result []FieldAlias
	for _, key := range section.Keys() {
		if strings.HasPrefix(key.Name(), "FIELDALIAS-") {
			from, to := parseFieldAliasExpr(key.Value())
			result = append(result, FieldAlias{
				Name: key.Name(),
				From: from,
				To:   to,
			})
		}
	}
	return result
}

// parseFieldAliasExpr reads an expression such as `sender as src_user`
// and return from and to fields
func parseFieldAliasExpr(expr string) (string, string) {
	captures := fieldAliasRegex.FindStringSubmatch(expr)
	// TODO handle mismatches
	if len(captures) != 3 {
		return "", ""
	}
	return captures[1], captures[2]
}

// orderProps orders props from more specific to more general.
// For settings that are specified in multiple categories of matching [<spec>]
// stanzas, [host::<host>] settings override [<sourcetype>] settings.
// Additionally, [source::<source>] settings override both [host::<host>]
// and [<sourcetype>] settings.
func orderProps(props []Prop) {
	sort.Slice(props, func(i, j int) bool {
		ti := props[i].Type()
		tj := props[j].Type()
		if ti == tj {
			return props[i].Name < props[j].Name
		}
		return ti < tj
	})
}
