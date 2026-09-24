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

import "strconv"

// SubsetValidator is the default comparison. For each reference record it
// asserts only the fields the reference actually sets; a zero field means "not
// asserted" and is skipped. Volatile keys (indextime, stream/ACK ids) never
// appear in a reference, so they are ignored for free.
//
// Records are compared position by position. A count mismatch is reported as
// a top-level mismatch so drops and dupes surface even when the overlapping
// records match.
type SubsetValidator struct{}

func (SubsetValidator) Validate(reference, candidate []Record) Result {
	var res Result

	if len(reference) != len(candidate) {
		res.Mismatches = append(res.Mismatches, Mismatch{
			Record:   -1,
			Field:    "count",
			Expected: strconv.Itoa(len(reference)),
			Actual:   strconv.Itoa(len(candidate)),
		})
	}

	n := min(len(reference), len(candidate))
	for i := 0; i < n; i++ {
		ref, cand := reference[i], candidate[i]
		add := func(field, exp, act string) {
			if exp != "" && exp != act {
				res.Mismatches = append(res.Mismatches, Mismatch{Record: i, Field: field, Expected: exp, Actual: act})
			}
		}
		add("raw", ref.Raw, cand.Raw)
		add("host", ref.Host, cand.Host)
		add("source", ref.Source, cand.Source)
		add("sourcetype", ref.Sourcetype, cand.Sourcetype)
		add("index", ref.Index, cand.Index)
		for k, exp := range ref.Fields {
			add("field:"+k, exp, cand.Fields[k])
		}
	}

	res.Match = len(res.Mismatches) == 0
	return res
}
