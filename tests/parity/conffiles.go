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
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// ConfFiles returns every *.conf file under dir as slash-separated paths
// relative to it, sorted. It is file discovery only: where a given path belongs
// in an agent's own tree is the Adapter's decision, made in Prepare.
func ConfFiles(configDir string) ([]string, error) {
	var rel []string
	err := filepath.WalkDir(configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".conf") {
			return nil
		}
		r, relErr := filepath.Rel(configDir, path)
		if relErr != nil {
			return relErr
		}
		rel = append(rel, filepath.ToSlash(r))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(rel)
	return rel, nil
}
