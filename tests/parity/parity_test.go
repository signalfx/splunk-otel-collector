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
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// fakeBackend is an in-memory Backend: it returns a fixed set of records from
// Search so the runner's capture loop can be exercised without Docker.
type fakeBackend struct {
	hec       HEC
	searchErr error
	lastSPL   string
	records   []Record
}

func (f *fakeBackend) Start(context.Context) error { return nil }
func (f *fakeBackend) Stop(context.Context) error  { return nil }
func (f *fakeBackend) HEC() HEC                    { return f.hec }

func (f *fakeBackend) Search(_ context.Context, spl string) ([]Record, error) {
	f.lastSPL = spl
	return f.records, f.searchErr
}

func (f *fakeBackend) Clean(context.Context, string) error { return nil }

// fakeAdapter records the lifecycle calls the runner makes and captures the
// interpolated inputs.conf it is handed, so tests can assert both.
type fakeAdapter struct {
	prepareErr error
	startErr   error
	name       string
	dir        string
	inputsConf string
	prepared   bool
	started    bool
	stopped    bool
	cleaned    bool
}

func (a *fakeAdapter) Name() string       { return a.name }
func (a *fakeAdapter) InstallDir() string { return a.dir }

func (a *fakeAdapter) Prepare(configDir string) error {
	a.prepared = true
	if b, err := os.ReadFile(filepath.Join(configDir, "inputs.conf")); err == nil {
		a.inputsConf = string(b)
	}
	return a.prepareErr
}

func (a *fakeAdapter) Start(context.Context) error { a.started = true; return a.startErr }
func (a *fakeAdapter) Stop(context.Context) error  { a.stopped = true; return nil }
func (a *fakeAdapter) Cleanup() error              { a.cleaned = true; return nil }

func fastOpts() RunOptions {
	return RunOptions{Quiescence: time.Millisecond, Timeout: 2 * time.Second, MinEvents: 1}
}

func TestRunAgent(t *testing.T) {
	rec := Record{Raw: "hi", Host: "myhost", Index: "parity_uc"}
	backend := &fakeBackend{hec: HEC{Endpoint: "https://splunk:8088", Token: "tok"}, records: []Record{rec}}
	a := &fakeAdapter{name: "fake", dir: t.TempDir()}
	c := &Case{
		Name:   "Set host",
		Setup:  "echo BASE_DIR > setup-marker",
		Script: "echo done >> setup-marker",
	}
	run := AgentRun{
		Adapter:     a,
		ConfigFiles: map[string]string{"inputs.conf": "index=INDEX\nhost=myhost"},
		Index:       "parity_uc",
		Search:      "search index=INDEX",
	}

	got, err := RunAgent(context.Background(), c, run, backend, fastOpts())
	if err != nil {
		t.Fatalf("RunAgent: %v", err)
	}
	if len(got) != 1 || got[0].Raw != rec.Raw || got[0].Host != rec.Host || got[0].Index != rec.Index {
		t.Errorf("records = %+v, want %+v", got, []Record{rec})
	}
	if a.inputsConf != "index=parity_uc\nhost=myhost" {
		t.Errorf("inputs.conf not interpolated: %q", a.inputsConf)
	}
	if backend.lastSPL != "search index=parity_uc" {
		t.Errorf("search SPL not interpolated: %q", backend.lastSPL)
	}
	if !a.prepared || !a.started || !a.stopped || !a.cleaned {
		t.Errorf("lifecycle incomplete: %+v", a)
	}
}

func TestRunAgentDefaultSearch(t *testing.T) {
	backend := &fakeBackend{records: []Record{{Raw: "x"}}}
	a := &fakeAdapter{name: "fake", dir: t.TempDir()}
	run := AgentRun{Adapter: a, Index: "parity_uf"} // no Search set
	if _, err := RunAgent(context.Background(), &Case{Name: "c"}, run, backend, fastOpts()); err != nil {
		t.Fatalf("RunAgent: %v", err)
	}
	if backend.lastSPL != "search index=parity_uf" {
		t.Errorf("default search = %q", backend.lastSPL)
	}
}

func TestRunAgentSetupError(t *testing.T) {
	backend := &fakeBackend{records: []Record{{Raw: "x"}}}
	a := &fakeAdapter{name: "fake", dir: t.TempDir()}
	run := AgentRun{Adapter: a, Index: "i"}
	c := &Case{Name: "c", Setup: "exit 3"}
	if _, err := RunAgent(context.Background(), c, run, backend, fastOpts()); err == nil {
		t.Fatal("expected setup error")
	}
	if a.started {
		t.Error("agent should not start after setup failure")
	}
}

func TestRunAgentPrepareError(t *testing.T) {
	backend := &fakeBackend{records: []Record{{Raw: "x"}}}
	a := &fakeAdapter{name: "fake", dir: t.TempDir(), prepareErr: errors.New("boom")}
	run := AgentRun{Adapter: a, Index: "i"}
	if _, err := RunAgent(context.Background(), &Case{Name: "c"}, run, backend, fastOpts()); err == nil {
		t.Fatal("expected prepare error")
	}
}

func TestRunAgentTimeoutReturnsLast(t *testing.T) {
	backend := &fakeBackend{} // Search always returns no records
	a := &fakeAdapter{name: "fake", dir: t.TempDir()}
	run := AgentRun{Adapter: a, Index: "i"}
	opts := RunOptions{Quiescence: time.Millisecond, Timeout: 1500 * time.Millisecond, MinEvents: 1}
	got, err := RunAgent(context.Background(), &Case{Name: "c"}, run, backend, opts)
	if err != nil {
		t.Fatalf("RunAgent: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want no records on timeout, got %+v", got)
	}
}

func TestRunCaseAgainstExpected(t *testing.T) {
	rec := Record{Raw: "hi", Host: "myhost"}
	backend := &fakeBackend{records: []Record{rec}}
	candidate := AgentRun{Adapter: &fakeAdapter{name: "otelcol", dir: t.TempDir()}, Index: "parity_uc"}
	c := &Case{Name: "c", Expected: Expected{Raw: "hi", Host: "myhost"}}

	// oracle.Adapter nil -> compared against the case's authored Expected.
	res, err := RunCase(context.Background(), c, backend, AgentRun{}, candidate, SubsetValidator{}, fastOpts())
	if err != nil {
		t.Fatalf("RunCase: %v", err)
	}
	if !res.Match {
		t.Errorf("expected match, got mismatches %+v", res.Mismatches)
	}
}

func TestRunCaseOracleVsCandidate(t *testing.T) {
	rec := Record{Raw: "hi", Host: "myhost"}
	backend := &fakeBackend{records: []Record{rec}}
	oracle := AgentRun{Adapter: &fakeAdapter{name: "UF", dir: t.TempDir()}, Index: "parity_uf"}
	candidate := AgentRun{Adapter: &fakeAdapter{name: "otelcol", dir: t.TempDir()}, Index: "parity_uc"}

	res, err := RunCase(context.Background(), &Case{Name: "c"}, backend, oracle, candidate, SubsetValidator{}, fastOpts())
	if err != nil {
		t.Fatalf("RunCase: %v", err)
	}
	if !res.Match {
		t.Errorf("oracle and candidate landed identical records but got mismatches %+v", res.Mismatches)
	}
}

func TestRunCaseCandidateError(t *testing.T) {
	backend := &fakeBackend{records: []Record{{Raw: "x"}}}
	candidate := AgentRun{Adapter: &fakeAdapter{name: "otelcol", dir: t.TempDir(), startErr: errors.New("nope")}, Index: "i"}
	if _, err := RunCase(context.Background(), &Case{Name: "c"}, backend, AgentRun{}, candidate, SubsetValidator{}, fastOpts()); err == nil {
		t.Fatal("expected candidate error")
	}
}

func TestRunOptionsDefaults(t *testing.T) {
	o := RunOptions{}.withDefaults()
	if o.Quiescence != 3*time.Second || o.Timeout != 90*time.Second || o.MinEvents != 1 || o.Shell != "bash" {
		t.Errorf("defaults = %+v", o)
	}
	custom := RunOptions{Quiescence: time.Second, Timeout: time.Minute, MinEvents: 2, Shell: "sh"}.withDefaults()
	if custom.Shell != "sh" || custom.MinEvents != 2 {
		t.Errorf("custom overridden: %+v", custom)
	}
}
