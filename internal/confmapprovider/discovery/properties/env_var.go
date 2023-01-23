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
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// SPLUNK_DISCOVERY_RECEIVERS_receiver_x2d_type_x2f_receiver_x2d_name_CONFIG_field_x3a_x3a_subfield=val
// SPLUNK_DISCOVERY_EXTENSIONS_observer_x2d_type_x2f_observer_x2d_name_CONFIG_field_x3a_x3a_subfield=val

var envVarLex = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Underscore", Pattern: `_`},
	{Name: "String", Pattern: `[^_]+`},
})

var envVarParser = participle.MustBuild[EnvVarProperty](
	participle.Lexer(envVarLex),
	participle.UseLookahead(participle.MaxLookahead),
)

type EnvVarProperty struct {
	ComponentType string            `parser:"'SPLUNK' Underscore 'DISCOVERY' Underscore @('RECEIVERS' | 'EXTENSIONS') Underscore"`
	Component     EnvVarComponentID `parser:"@@"`
	Key           string            `parser:"Underscore 'CONFIG' Underscore @(String|Underscore)+"`
	Val           string
}

type EnvVarComponentID struct {
	Type string `parser:"@~(Underscore (?= 'CONFIG'))+"`
	// _x2f_ -> '/'
	Name string `parser:"(Underscore 'x2f' Underscore @(~(?= Underscore (?= 'CONFIG'))+|''))?"`
}

func NewEnvVarProperty(property, val string) (*EnvVarProperty, error) {
	p, err := envVarParser.ParseString("SPLUNK_DISCOVERY", property)
	if err != nil {
		return nil, fmt.Errorf("invalid property env var (parsing error): %w", err)
	}
	p.Val = val
	return p, nil
}
