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
	"strings"
)

// GoldenFile is the checked-in reference for a case, one per case directory.
const GoldenFile = "golden.json"

// fieldPrefix marks an assert entry that targets a Record.Fields key, e.g.
// "field:punct" asserts Fields["punct"].
const fieldPrefix = "field:"

// Project returns a copy of rec holding only the fields named in assert; every
// other field is zeroed. It is how a case's filter is applied: the golden saves
// the projection of the oracle run, and replay compares the projection of the
// candidate run, so only the asserted fields ever take part.
func Project(rec Record, assert []string) Record {
	out := Record{Time: rec.Time}
	for _, f := range assert {
		if key, ok := strings.CutPrefix(f, fieldPrefix); ok {
			if v, present := rec.Fields[key]; present {
				if out.Fields == nil {
					out.Fields = map[string]string{}
				}
				out.Fields[key] = v
			}
			continue
		}
		switch f {
		case "raw":
			out.Raw = rec.Raw
		case "host":
			out.Host = rec.Host
		case "source":
			out.Source = rec.Source
		case "sourcetype":
			out.Sourcetype = rec.Sourcetype
		case "index":
			out.Index = rec.Index
		}
	}
	return out
}

// ProjectAll projects every record in recs.
func ProjectAll(recs []Record, assert []string) []Record {
	out := make([]Record, len(recs))
	for i, r := range recs {
		out[i] = Project(r, assert)
	}
	return out
}

// LoadGolden reads a golden file (a JSON array of projected Records).
func LoadGolden(path string) ([]Record, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var recs []Record
	if err := json.Unmarshal(b, &recs); err != nil {
		return nil, fmt.Errorf("parse golden %s: %w", path, err)
	}
	return recs, nil
}

// WriteGolden writes recs as an indented JSON array with a trailing newline, so
// regenerated goldens diff cleanly in review.
func WriteGolden(path string, recs []Record) error {
	b, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o600)
}
