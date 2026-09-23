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

import "testing"

// TestNormalizingValidator checks that the normalizer blanks exactly the
// configured fields on both sides before delegating, so index/source/sourcetype
// differences that exist by construction do not read as parity mismatches while
// raw/host still do.
func TestNormalizingValidator(t *testing.T) {
	// oracle and candidate agree on raw+host but differ on index/source/
	// sourcetype the way a real UF-vs-collector run does.
	oracle := Record{
		Raw:        "initial text",
		Host:       "myhost",
		Source:     "/tmp/parity-uf-abc/foo.txt",
		Sourcetype: "foo-too_small",
		Index:      "parity_uf",
	}
	candidate := Record{
		Raw:        "initial text",
		Host:       "myhost",
		Source:     "/tmp/parity-uc-xyz/foo.txt",
		Sourcetype: "httpevent",
		Index:      "parity_uc",
	}

	tests := []struct {
		name       string
		reference  []Record
		candidate  []Record
		wantFields []string // mismatch fields expected, order-independent
		normalizer Normalizer
		wantMatch  bool
	}{
		{
			name:       "parity normalizer ignores index/source/sourcetype",
			normalizer: ParityNormalizer(),
			reference:  []Record{oracle},
			candidate:  []Record{candidate},
			wantMatch:  true,
		},
		{
			name:       "raw mismatch still reported after normalization",
			normalizer: ParityNormalizer(),
			reference:  []Record{oracle},
			candidate:  []Record{{Raw: "different", Host: "myhost", Index: "parity_uc"}},
			wantMatch:  false,
			wantFields: []string{"raw"},
		},
		{
			name:       "host mismatch still reported after normalization",
			normalizer: ParityNormalizer(),
			reference:  []Record{oracle},
			candidate:  []Record{{Raw: "initial text", Host: "other", Index: "parity_uc"}},
			wantMatch:  false,
			wantFields: []string{"host"},
		},
		{
			name:       "without normalization the divergent fields fail",
			normalizer: Normalizer{},
			reference:  []Record{oracle},
			candidate:  []Record{candidate},
			wantMatch:  false,
			wantFields: []string{"source", "sourcetype", "index"},
		},
		{
			name:       "count mismatch surfaces from inner validator",
			normalizer: ParityNormalizer(),
			reference:  []Record{oracle, oracle},
			candidate:  []Record{candidate},
			wantMatch:  false,
			wantFields: []string{"count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NormalizingValidator{Normalizer: tt.normalizer}
			res := v.Validate(tt.reference, tt.candidate)
			if res.Match != tt.wantMatch {
				t.Fatalf("Match = %v, want %v (mismatches: %+v)", res.Match, tt.wantMatch, res.Mismatches)
			}
			got := map[string]bool{}
			for _, m := range res.Mismatches {
				got[m.Field] = true
			}
			if len(got) != len(tt.wantFields) {
				t.Fatalf("mismatch fields = %v, want %v", got, tt.wantFields)
			}
			for _, f := range tt.wantFields {
				if !got[f] {
					t.Errorf("missing expected mismatch on field %q; got %+v", f, res.Mismatches)
				}
			}
		})
	}
}

// TestNormalizerApplyDoesNotMutate confirms applyAll returns copies and leaves
// the caller's records untouched, so a validator can be run repeatedly.
func TestNormalizerApplyDoesNotMutate(t *testing.T) {
	in := []Record{{Index: "parity_uf", Source: "/x/foo.txt", Sourcetype: "st", Raw: "r", Host: "h"}}
	out := ParityNormalizer().applyAll(in)

	if in[0].Index == "" || in[0].Source == "" || in[0].Sourcetype == "" {
		t.Errorf("input was mutated: %+v", in[0])
	}
	if out[0].Index != "" || out[0].Source != "" || out[0].Sourcetype != "" {
		t.Errorf("normalized record kept a blanked field: %+v", out[0])
	}
	if out[0].Raw != "r" || out[0].Host != "h" {
		t.Errorf("normalizer touched raw/host: %+v", out[0])
	}
}
