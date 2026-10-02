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
	"reflect"
	"testing"
	"time"
)

func TestProject(t *testing.T) {
	ts := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	full := Record{
		Time:       ts,
		Raw:        "hi",
		Host:       "myhost",
		Source:     "/tmp/foo.txt",
		Sourcetype: "too_small",
		Index:      "parity",
		Fields: map[string]string{
			"punct": "...", "linecount": "1",
			"date_hour": "12", "date_wday": "wednesday",
		},
	}

	tests := []struct {
		want   Record
		name   string
		filter Expected
	}{
		{
			name:   "empty filter keeps nothing but time",
			filter: Expected{},
			want:   Record{Time: ts},
		},
		{
			name:   "only the selected fields",
			filter: Expected{Raw: true, Host: true},
			want:   Record{Time: ts, Raw: "hi", Host: "myhost"},
		},
		{
			name:   "all typed fields",
			filter: Expected{Raw: true, Host: true, Source: true, Sourcetype: true, Index: true},
			want: Record{
				Time: ts, Raw: "hi", Host: "myhost",
				Source: "/tmp/foo.txt", Sourcetype: "too_small", Index: "parity",
			},
		},
		{
			name:   "fields selected by exact name",
			filter: Expected{Fields: []string{"punct"}},
			want:   Record{Time: ts, Fields: map[string]string{"punct": "..."}},
		},
		{
			name:   "fields selected by glob",
			filter: Expected{Fields: []string{"date_*"}},
			want:   Record{Time: ts, Fields: map[string]string{"date_hour": "12", "date_wday": "wednesday"}},
		},
		{
			name:   "glob matching no key yields nothing",
			filter: Expected{Fields: []string{"missing_*"}},
			want:   Record{Time: ts},
		},
		{
			name:   "malformed glob matches nothing rather than erroring",
			filter: Expected{Fields: []string{"[unclosed"}},
			want:   Record{Time: ts},
		},
		{
			name:   "patterns union",
			filter: Expected{Fields: []string{"punct", "linecount"}},
			want:   Record{Time: ts, Fields: map[string]string{"punct": "...", "linecount": "1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := project(full, tt.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("project() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestWriteGoldenMatchesValidator pins the invariant the whole design leans on:
// the fields the filter writes into a golden are exactly the fields the validator
// then compares, so one filter governs both generation and validation. Record's
// omitempty tags alone would not achieve this, because a capture arrives with
// source, sourcetype and index populated whether or not the case selects them.
func TestWriteGoldenMatchesValidator(t *testing.T) {
	filter := Expected{Raw: true, Host: true, Fields: []string{"punct"}}
	captured := Record{
		Raw: "initial text", Host: "myhost",
		Source: "/tmp/foo.txt", Sourcetype: "too_small", Index: "parity",
		Fields: map[string]string{"punct": "...", "linecount": "1"},
	}

	file := filepath.Join(t.TempDir(), GoldenFile)
	if err := WriteGolden(file, []Record{captured}, filter); err != nil {
		t.Fatalf("WriteGolden: %v", err)
	}
	golden, err := LoadGolden(file)
	if err != nil {
		t.Fatalf("LoadGolden: %v", err)
	}
	want := []Record{{Raw: "initial text", Host: "myhost", Fields: map[string]string{"punct": "..."}}}
	if !reflect.DeepEqual(golden, want) {
		t.Fatalf("golden = %+v, want only the selected fields %+v", golden, want)
	}

	// A candidate differing on every unselected field still matches: the collector
	// sets its own sourcetype and monitors a different sandbox path, so those
	// really do diverge from UF in practice.
	candidate := Record{
		Raw: "initial text", Host: "myhost",
		Source: "/other/path", Sourcetype: "otel", Index: "parity",
		Fields: map[string]string{"punct": "...", "linecount": "99"},
	}
	if res := (SubsetValidator{}).Validate(golden, []Record{candidate}); !res.Match {
		t.Errorf("unselected fields should be ignored, got %+v", res.Mismatches)
	}

	// A candidate differing on a selected field does not.
	candidate.Fields["punct"] = "!!!"
	if res := (SubsetValidator{}).Validate(golden, []Record{candidate}); res.Match {
		t.Error("selected punct differs but validation passed")
	}
}

func TestGoldenRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), GoldenFile)
	// Time must not survive the round trip: Record.Time is json:"-".
	recs := []Record{{Time: time.Now(), Raw: "hi", Host: "myhost", Fields: map[string]string{"punct": "..."}}}
	filter := Expected{Raw: true, Host: true, Fields: []string{"punct"}}
	if err := WriteGolden(path, recs, filter); err != nil {
		t.Fatalf("WriteGolden: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "fields": {
      "punct": "..."
    },
    "raw": "hi",
    "host": "myhost"
  }
]
`
	if string(b) != want {
		t.Errorf("golden file =\n%s\nwant\n%s", b, want)
	}

	got, err := LoadGolden(path)
	if err != nil {
		t.Fatalf("LoadGolden: %v", err)
	}
	wantRecs := []Record{{Raw: "hi", Host: "myhost", Fields: map[string]string{"punct": "..."}}}
	if !reflect.DeepEqual(got, wantRecs) {
		t.Errorf("LoadGolden() = %+v, want %+v", got, wantRecs)
	}
}

func TestLoadGoldenMissing(t *testing.T) {
	if _, err := LoadGolden(filepath.Join(t.TempDir(), GoldenFile)); err == nil {
		t.Error("expected error for missing golden")
	}
}

func TestLoadGoldenMalformed(t *testing.T) {
	path := filepath.Join(t.TempDir(), GoldenFile)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGolden(path); err == nil {
		t.Error("expected parse error for malformed golden")
	}
}
