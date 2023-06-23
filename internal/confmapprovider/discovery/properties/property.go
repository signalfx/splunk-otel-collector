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

package properties

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// Discovery properties are the method of configuring individual components for discovery mode.
// They are available as commandline --set options and provide equivalent environment variables.
// They are always of the format:
// splunk.discovery.receivers.<receiver-type(/name)>.config.<field>(<::subfield>)*=value
// splunk.discovery.extensions.<observer-type(/name)>.config.<field>(<::subfield>)*=value
// with corresponding env var:
// SPLUNK_DISCOVERY_RECEIVERS_receiver_x2d_type_x2f_receiver_x2d_name_CONFIG_field_x3a__x3a_subfield=value
// SPLUNK_DISCOVERY_EXTENSIONS_observer_x2d_type_x2f_observer_x2d_name_CONFIG_field_x3a__x3a_subfield=value

// Parsing properties requires lookaheads (backtracking), which isn't possible in re2. Using participle we
// can define a simple lexer and grammar to establish the Property type as an ast.
var lex = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Dot", Pattern: `\.`},
	{Name: "ForwardSlash", Pattern: `/`},
	{Name: "Whitespace", Pattern: `\s+`},
	{Name: "String", Pattern: `[^./]*`},
})

var parser = participle.MustBuild[Property](
	participle.Lexer(lex),
	participle.Elide("Whitespace"),
	participle.UseLookahead(participle.MaxLookahead),
)

// Property is the ast for a parsed property.
// TODO: support rules and resource_attributes instead of just embedded config
type Property struct {
	stringMap     map[string]any
	ComponentType string      `parser:"'splunk' Dot 'discovery' Dot @('receivers' | 'extensions') Dot"`
	Component     ComponentID `parser:"@@ Dot"`
	Type          string      `parser:"((@'config' Dot)|@('enabled'))"`
	Key           string      `parser:"@(String|Dot|ForwardSlash)*"`
	Val           string
	Input         string
}

type ComponentID struct {
	Type string `parser:"@~(ForwardSlash | (Dot (?= ('config'|'enabled'))))+"`
	Name string `parser:"(ForwardSlash @(~(Dot (?= ('config'|'enabled')))+)*)?"`
}

func NewProperty(property, val string) (*Property, error) {
	p, err := parser.ParseString("splunk.discovery", property)
	if err != nil {
		return nil, fmt.Errorf("invalid property %q (parsing error): %w", property, err)
	}
	p.Val = val

	var subStringMap map[string]any
	switch p.Type {
	case "enabled":
		bVal, e := strconv.ParseBool(p.Val)
		if e != nil {
			return nil, fmt.Errorf("failed parsing %q bool: %w", property, e)
		}
		p.Val = fmt.Sprintf("%t", bVal)
		subStringMap = map[string]any{"enabled": p.Val}
	case "config":
		var dst map[string]any
		cfgItem := []byte(fmt.Sprintf("%s: %s", p.Key, val))
		if err = yaml.Unmarshal(cfgItem, &dst); err != nil {
			return nil, fmt.Errorf("failed unmarshaling property %q: %w", p.Key, err)
		}
		subStringMap = map[string]any{"config": confmap.NewFromStringMap(dst).ToStringMap()}
	}
	p.stringMap = map[string]any{
		p.ComponentType: map[string]any{
			component.NewIDWithName(component.Type(p.Component.Type), p.Component.Name).String(): subStringMap,
		},
	}
	p.Input = property
	return p, nil
}

// ToEnvVar will output the equivalent env var property for informational purposes.
func (p *Property) ToEnvVar() string {
	envVar := envVarPrefixS
	envVar = fmt.Sprintf("%s%s_", envVar, strings.ToUpper(p.ComponentType))
	envVar = fmt.Sprintf("%s%s", envVar, wordify(p.Component.Type))
	// preserve input of `component.type/` with no name
	if p.Component.Name != "" || strings.HasPrefix(p.Input, fmt.Sprintf("splunk.discovery.%s.%s/.", p.ComponentType, p.Component.Type)) {
		envVar = fmt.Sprintf("%s%s", envVar, wordify(fmt.Sprintf("/%s", p.Component.Name)))
	}
	envVar = fmt.Sprintf("%s_%s", envVar, strings.ToUpper(p.Type))
	if p.Type == "config" {
		envVar = fmt.Sprintf("%s_%s", envVar, wordify(p.Key))
	}
	return envVar
}

// ToStringMap() will return a map[string]any equivalent to the property's root-level confmap.ToStringMap()
func (p *Property) ToStringMap() map[string]any {
	if p != nil {
		return p.stringMap
	}
	return nil
}

const (
	envVarPrefixS = "SPLUNK_DISCOVERY_"
)

var (
	envVarPrefixRE = regexp.MustCompile(fmt.Sprintf("^%s", envVarPrefixS))
	envVarHexRE    = regexp.MustCompile("_x[0-9a-fA-F]+_")
)

func NewPropertyFromEnvVar(envVar, val string) (*Property, bool, error) {
	if !envVarPrefixRE.MatchString(envVar) {
		return nil, false, nil
	}
	evp, err := NewEnvVarProperty(envVar, val)
	if err != nil {
		return nil, true, fmt.Errorf("invalid env var property (parsing error): %w", err)
	}

	cid, err := unwordify(evp.Component.Type)
	if err != nil {
		return nil, true, fmt.Errorf("failed parsing env var property component id type: %w", err)
	}

	if evp.Component.Name != "" {
		cidName, e := unwordify(evp.Component.Name)
		if e != nil {
			return nil, true, fmt.Errorf("failed parsing env var property component id name: %w", err)
		}
		cid = fmt.Sprintf("%s/%s", cid, cidName)
	}

	key, err := unwordify(evp.Key)
	if err != nil {
		return nil, true, fmt.Errorf("failed parsing env var property key: %w", err)
	}

	pType := strings.ToLower(evp.Type)
	property := fmt.Sprintf("splunk.discovery.%s.%s.%s", strings.ToLower(evp.ComponentType), cid, pType)
	if pType == "config" {
		property = fmt.Sprintf("%s.%s", property, key)
	}

	prop, err := NewProperty(property, val)
	return prop, true, err
}

// wordify takes an arbitrary string (utf8) and will hex encode any rune not in \w, escaping with `_x<hex-encoded-rune>_`.
func wordify(text string) string {
	var wordified []rune
	for _, c := range text {
		// encoded all non-word characters to hex
		if c != '_' && c < '0' || (c > '9') && (c < 'A') || (c > 'Z') && (c < 'a') || (c > 'z') {
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(c))
			hexEncoded := make([]byte, len(b)*2) // hex.EncodedLen
			hex.Encode(hexEncoded, b)

			// strip all leading '0' unless evenness at stake
			for len(hexEncoded) > 0 && hexEncoded[0] == '0' {
				if len(hexEncoded) > 1 {
					if hexEncoded[1] != '0' && len(hexEncoded)%2 == 0 {
						break
					}
				}
				hexEncoded = hexEncoded[1:]
			}
			for _, r := range fmt.Sprintf("_x%s_", hexEncoded) {
				wordified = append(wordified, r)
			}
		} else {
			wordified = append(wordified, c)
		}
	}
	return string(wordified)
}

// unwordify takes any string, expanding `_x<.>_` content as hex-decoded utf8 strings
func unwordify(text string) (string, error) {
	var err error
	unwordified := envVarHexRE.ReplaceAllStringFunc(text, func(s string) string {
		s = s[2 : len(s)-1]
		decoded, e := hex.DecodeString(s)
		if e != nil {
			err = multierr.Combine(err, fmt.Errorf("%q: %w", s, e))
			return ""
		}
		// left pad if too short for uint32 conversion
		for len(decoded) < 4 {
			decoded = append([]byte{0}, decoded...)
		}
		r := int32(binary.BigEndian.Uint32(decoded))
		return fmt.Sprintf("%c", r)
	})
	if err != nil {
		return "", fmt.Errorf("failed parsing env var hex-encoded content: %w", err)
	}
	return unwordified, nil

}
