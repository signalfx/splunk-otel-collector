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
	"strings"

	"gopkg.in/yaml.v3"
)

// Case is one parity test loaded from a test.yaml. It is the ported shape of
// the 1spl case: a Splunk .conf fragment, shell setup/script hooks, and the
// expected event. Fields are deliberately close to 1spl so the corpus ports
// with minimal edits.
type Case struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Stage       string   `yaml:"stage"`
	Conf        string   `yaml:"conf"`     // inputs.conf fragment
	Setup       string   `yaml:"setup"`    // shell run before the agent starts
	Script      string   `yaml:"script"`   // shell run after the agent starts
	Expected    Expected `yaml:"expected"` // reference event
	OS          []string `yaml:"os"`
}

// Expected is the reference event a case asserts, authored in Splunk-event
// terms: the same fields a search returns and that Record holds. Only set fields
// are asserted; empty fields are ignored, which is how volatile keys stay out of
// the comparison.
type Expected struct {
	Fields     map[string]string `yaml:"fields"`
	Raw        string            `yaml:"raw"`
	Host       string            `yaml:"host"`
	Source     string            `yaml:"source"`
	Sourcetype string            `yaml:"sourcetype"`
	Index      string            `yaml:"index"`
}

// LoadCase reads a test.yaml from path.
func LoadCase(path string) (*Case, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Case
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
}

// Tokens are the interpolation variables a case's conf, agent config templates,
// and shell hooks may use. They are resolved per run because the sandbox, HEC
// port, and per-agent index are dynamic.
type Tokens struct {
	BaseDir     string // per-run sandbox working directory
	AgentDir    string // agent install root
	HECEndpoint string // Backend HEC endpoint, e.g. https://127.0.0.1:32769
	HECToken    string // Backend HEC token
	Index       string // Splunk index this agent forwards to
}

func (t Tokens) apply(s string) string {
	r := strings.NewReplacer(
		"BASE_DIR", t.BaseDir,
		"AGENT_DIR", t.AgentDir,
		"HEC_ENDPOINT", t.HECEndpoint,
		"HEC_TOKEN", t.HECToken,
		"INDEX", t.Index,
	)
	return r.Replace(s)
}

// AsReference turns the authored Expected into a single-record reference
// capture. Empty fields stay empty and are ignored by the validator.
func (e Expected) AsReference() Record {
	return Record{
		Raw:        e.Raw,
		Host:       e.Host,
		Source:     e.Source,
		Sourcetype: e.Sourcetype,
		Index:      e.Index,
		Fields:     e.Fields,
	}
}
