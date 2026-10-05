// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parity

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Case is one parity test loaded from a test.yaml: shell setup/script hooks and
// the fields to compare. The agent configs live beside it in the case directory,
// not in here.
type Case struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Stage       string   `yaml:"stage"`
	Setup       string   `yaml:"setup"`  // shell run before the agent starts
	Script      string   `yaml:"script"` // shell run after the agent starts
	OS          []string `yaml:"os"`
	// Expected is last because its trailing booleans pack better there.
	Expected Expected `yaml:"expected"` // the fields this case compares
}

// Expected selects the Record fields a case is defined on. It is the case's only
// filter, applied twice: generating a golden keeps these fields of the oracle's
// events and drops the rest, and validating compares the candidate on these same
// fields. The values live in the golden, never here, so the two cannot disagree.
//
// Every field a case does not select is ignored, which is how volatile keys (the
// sandbox path in source, an auto-assigned sourcetype) stay out of a comparison.
type Expected struct {
	// Fields selects Record.Fields keys by exact name or glob, e.g. "punct" or
	// "date_*". Globs use path.Match syntax.
	Fields     []string `yaml:"fields"`
	Raw        bool     `yaml:"raw"`
	Host       bool     `yaml:"host"`
	Source     bool     `yaml:"source"`
	Sourcetype bool     `yaml:"sourcetype"`
	Index      bool     `yaml:"index"`
}

// selectsNothing reports whether the filter would compare no fields at all,
// which makes a case vacuous.
func (e Expected) selectsNothing() bool {
	return !e.Raw && !e.Host && !e.Source && !e.Sourcetype && !e.Index && len(e.Fields) == 0
}

// LoadCase reads a test.yaml from path. Unknown keys are an error: the filter is
// all booleans, so a typo would silently stop comparing a field instead of
// failing.
func LoadCase(path string) (*Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var c Case
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if c.Expected.selectsNothing() {
		return nil, fmt.Errorf("%s: expected selects no fields, so the case compares nothing", path)
	}
	return &c, nil
}

// Tokens are the interpolation variables a case's conf, agent config templates,
// and shell hooks may use. They are resolved per run because the sandbox, HEC
// port, and per-agent index are dynamic.
type Tokens struct {
	BaseDir     string // per-run sandbox working directory
	HECEndpoint string // Backend HEC endpoint, e.g. https://127.0.0.1:32769
	HECToken    string // Backend HEC token
	Index       string // Splunk index this agent forwards to
}

// tokenRef matches a ${NAME} reference. Only the braced form is a token, so
// shell constructs ($1, $(date)) and the collector's own ${env:...} references
// pass through untouched, as does any bare occurrence of a token name in event
// text or a config key.
var tokenRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// apply substitutes the ${NAME} references a case may use. Splunk .conf files
// have no expansion of their own, so the framework does it for every config it
// renders. Unknown names are left as they are.
func (t Tokens) apply(s string) string {
	vals := map[string]string{
		"BASE_DIR":     t.BaseDir,
		"HEC_ENDPOINT": t.HECEndpoint,
		"HEC_TOKEN":    t.HECToken,
		"INDEX":        t.Index,
	}
	return tokenRef.ReplaceAllStringFunc(s, func(ref string) string {
		if v, ok := vals[ref[2:len(ref)-1]]; ok {
			return v
		}
		return ref
	})
}
