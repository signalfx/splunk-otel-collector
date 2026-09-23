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
// list of event fields to assert. Fields are deliberately close to 1spl so the
// corpus ports with minimal edits.
type Case struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Stage       string `yaml:"stage"`
	Conf        string `yaml:"conf"`   // inputs.conf fragment
	Setup       string `yaml:"setup"`  // shell run before the agent starts
	Script      string `yaml:"script"` // shell run after the agent starts
	// Assert names the Record fields this case compares against the golden:
	// "raw", "host", "source", "sourcetype", "index", or "field:<key>" for a
	// Fields entry. It is the case's filter: -update saves only these fields
	// from the oracle run, and replay compares only these. Keeping the golden to
	// the asserted fields avoids checking in large, volatile payloads.
	Assert []string `yaml:"assert"`
	OS     []string `yaml:"os"`
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
