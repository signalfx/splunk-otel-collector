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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// GoldenFile is the checked-in reference for a case, one per case directory.
const GoldenFile = "golden.json"

// ObservedFile is the checked-in reference for a case's observe hook, in the
// case directory beside the golden. Only a case with an observe hook has one.
// It is plain text rather than part of the golden because the golden is an array
// of events and an observation is not an event.
const ObservedFile = "observed.txt"

// LoadObserved reads a case's observed reference.
func LoadObserved(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// WriteObserved writes the oracle's observation, newline-terminated so it diffs
// as one line in review.
func WriteObserved(file, observation string) error {
	return os.WriteFile(file, []byte(observation+"\n"), 0o600)
}

// project returns a copy of rec holding only the fields the filter selects;
// every other field is zeroed so Record's omitempty tags drop it from the
// golden. A capture arrives with source, sourcetype and index populated whether
// or not a case selects them, so without this a golden would pin the sandbox
// path and the oracle's index.
func project(rec Record, e Expected) Record {
	out := Record{Time: rec.Time}
	if e.Raw {
		out.Raw = rec.Raw
	}
	if e.Host {
		out.Host = rec.Host
	}
	if e.Source {
		out.Source = rec.Source
	}
	if e.Sourcetype {
		out.Sourcetype = rec.Sourcetype
	}
	if e.Index {
		out.Index = rec.Index
	}
	for key, v := range rec.Fields {
		if !e.selectsField(key) {
			continue
		}
		if out.Fields == nil {
			out.Fields = map[string]string{}
		}
		out.Fields[key] = v
	}
	return out
}

// selectsField reports whether key matches any of the filter's Fields patterns.
// A malformed pattern matches nothing rather than erroring, so a bad glob shows
// up as a missing field in the golden diff.
func (e Expected) selectsField(key string) bool {
	for _, pattern := range e.Fields {
		if ok, err := path.Match(pattern, key); err == nil && ok {
			return true
		}
	}
	return false
}

// LoadGolden reads a golden file (a JSON array of projected Records).
func LoadGolden(file string) ([]Record, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var recs []Record
	if err := json.Unmarshal(b, &recs); err != nil {
		return nil, fmt.Errorf("parse golden %s: %w", file, err)
	}
	return recs, nil
}

// WriteGolden writes recs reduced to the fields the filter selects, as an
// indented JSON array with a trailing newline so regenerated goldens diff
// cleanly in review. Reducing here rather than at the call site means a golden
// cannot be written holding fields its case does not compare.
func WriteGolden(file string, recs []Record, e Expected) error {
	reduced := make([]Record, len(recs))
	for i, r := range recs {
		reduced[i] = project(r, e)
	}
	b, err := json.MarshalIndent(reduced, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(file, b, 0o600)
}
