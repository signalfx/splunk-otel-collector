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
	"os"
	"path/filepath"
	"testing"
)

func TestSubsetValidator(t *testing.T) {
	tests := []struct {
		name           string
		reference      []Record
		candidate      []Record
		wantMismatches []string // Field values of expected mismatches, order-independent
		wantMatch      bool
	}{
		{
			name:      "empty reference matches anything",
			reference: []Record{{}},
			candidate: []Record{{Raw: "whatever", Host: "h", Source: "s"}},
			wantMatch: true,
		},
		{
			name:      "set fields must match",
			reference: []Record{{Raw: "hi", Host: "myhost"}},
			candidate: []Record{{Raw: "hi", Host: "myhost", Source: "ignored"}},
			wantMatch: true,
		},
		{
			name:           "field mismatch reported",
			reference:      []Record{{Host: "myhost"}},
			candidate:      []Record{{Host: "other"}},
			wantMatch:      false,
			wantMismatches: []string{"host"},
		},
		{
			name:           "custom fields compared",
			reference:      []Record{{Fields: map[string]string{"index": "a", "punct": "..."}}},
			candidate:      []Record{{Fields: map[string]string{"index": "b"}}},
			wantMatch:      false,
			wantMismatches: []string{"field:index", "field:punct"},
		},
		{
			name:           "count mismatch reported",
			reference:      []Record{{Raw: "one"}, {Raw: "two"}},
			candidate:      []Record{{Raw: "one"}},
			wantMatch:      false,
			wantMismatches: []string{"count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := SubsetValidator{}.Validate(tt.reference, tt.candidate)
			if res.Match != tt.wantMatch {
				t.Fatalf("Match = %v, want %v (mismatches: %+v)", res.Match, tt.wantMatch, res.Mismatches)
			}
			got := map[string]bool{}
			for _, m := range res.Mismatches {
				got[m.Field] = true
			}
			if len(res.Mismatches) != len(tt.wantMismatches) {
				t.Fatalf("got %d mismatches %+v, want fields %v", len(res.Mismatches), res.Mismatches, tt.wantMismatches)
			}
			for _, f := range tt.wantMismatches {
				if !got[f] {
					t.Errorf("missing expected mismatch on field %q, got %+v", f, res.Mismatches)
				}
			}
		})
	}
}

func TestTokensApply(t *testing.T) {
	tokens := Tokens{
		BaseDir:     "/tmp/run",
		AgentDir:    "/opt/agent",
		HECEndpoint: "https://127.0.0.1:8088",
		HECToken:    "tok",
		Index:       "parity_uc",
	}
	in := "dir=BASE_DIR agent=AGENT_DIR ep=HEC_ENDPOINT token=HEC_TOKEN idx=INDEX"
	want := "dir=/tmp/run agent=/opt/agent ep=https://127.0.0.1:8088 token=tok idx=parity_uc"
	if got := tokens.apply(in); got != want {
		t.Errorf("apply() = %q, want %q", got, want)
	}
}

func TestLoadCase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	content := `name: "Set host"
description: custom host
stage: alpha
conf: |
  [monitor:///BASE_DIR/foo.txt]
  host=myhost
setup: |
  echo hi > foo.txt
expected:
  raw: "hi"
  host: myhost
os: ["darwin", "linux"]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCase(path)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}
	if c.Name != "Set host" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.Expected.Host != "myhost" || c.Expected.Raw != "hi" {
		t.Errorf("Expected = %+v", c.Expected)
	}
	if len(c.OS) != 2 || c.OS[0] != "darwin" {
		t.Errorf("OS = %v", c.OS)
	}
}

func TestLoadCaseError(t *testing.T) {
	if _, err := LoadCase(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestExpectedAsReference(t *testing.T) {
	e := Expected{
		Raw:    "hi",
		Host:   "myhost",
		Fields: map[string]string{"index": "a"},
	}
	r := e.AsReference()
	if r.Raw != "hi" || r.Host != "myhost" || r.Fields["index"] != "a" {
		t.Errorf("AsReference = %+v", r)
	}
	if r.Source != "" || r.Sourcetype != "" {
		t.Errorf("unset fields should stay empty: %+v", r)
	}
}

func TestSanitize(t *testing.T) {
	tests := map[string]string{
		"Set host":   "Set-host",
		"a/b:c":      "a-b-c",
		"keep123ABC": "keep123ABC",
		"":           "",
	}
	for in, want := range tests {
		if got := sanitize(in); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}
