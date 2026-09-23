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

// Normalizer canonicalizes a Record before a parity comparison, blanking the
// fields that differ between the oracle and the candidate by construction
// rather than by a real defect. A blanked field is not asserted (SubsetValidator
// skips empty reference fields), so the comparison is left with the fields the
// two agents are expected to reproduce identically.
//
// The knobs are deliberately narrow: each names a field that today's oracle
// (UF) sets in a way the candidate does not match yet, and each is documented
// with why. Tightening any of these back into the comparison (e.g. mapping the
// collector's com.splunk.source / com.splunk.sourcetype so source and sourcetype
// become real parity assertions) is a follow-up, not new config surface here.
type Normalizer struct {
	// IgnoreIndex drops index. The oracle and candidate forward to their own
	// indexes (parity_uf / parity_uc) so a run needs no clean-between step; the
	// index is never expected to match.
	IgnoreIndex bool
	// IgnoreSource drops source. UF sets source to the monitored file path,
	// which also carries the per-agent sandbox prefix; the collector does not
	// map com.splunk.source yet, so its source is not the file path.
	IgnoreSource bool
	// IgnoreSourcetype drops sourcetype. UF auto-assigns it from the file; the
	// collector's splunk_hec events default to httpevent.
	IgnoreSourcetype bool
}

// ParityNormalizer is the default normalization for a UF-vs-candidate run: it
// excludes the index, source, and sourcetype the two agents assign differently,
// leaving raw and host (and any asserted custom fields) as the parity signal.
func ParityNormalizer() Normalizer {
	return Normalizer{IgnoreIndex: true, IgnoreSource: true, IgnoreSourcetype: true}
}

func (n Normalizer) apply(r Record) Record {
	if n.IgnoreIndex {
		r.Index = ""
	}
	if n.IgnoreSource {
		r.Source = ""
	}
	if n.IgnoreSourcetype {
		r.Sourcetype = ""
	}
	return r
}

func (n Normalizer) applyAll(recs []Record) []Record {
	out := make([]Record, len(recs))
	for i, r := range recs {
		out[i] = n.apply(r)
	}
	return out
}

// NormalizingValidator applies a Normalizer to both the reference and the
// candidate, then delegates to Inner. It turns the same field-level comparison
// used against an authored Expected into a direct oracle-vs-candidate parity
// check: the oracle's real capture is the reference, minus the fields the two
// agents legitimately assign differently. Inner defaults to SubsetValidator.
type NormalizingValidator struct {
	Inner      Validator
	Normalizer Normalizer
}

func (v NormalizingValidator) Validate(reference, candidate []Record) Result {
	inner := v.Inner
	if inner == nil {
		inner = SubsetValidator{}
	}
	return inner.Validate(v.Normalizer.applyAll(reference), v.Normalizer.applyAll(candidate))
}
