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
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ConfFiles returns every *.conf file under configDir as slash-separated paths
// relative to it, sorted. Both adapters walk the tree rather than reading one
// level, so a case can lay out a conf root with apps in it.
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
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sort.Strings(rel)
	return rel, nil
}

// ConfDest maps a conf path relative to the case's conf/ onto its location
// under an agent's etc/, returned relative to etc/.
//
// A bare filename is system-local config, which is where a case's .conf files
// have always gone. A nested path is mirrored verbatim, which is how a case
// expresses an app: conf/apps/my_app/local/inputs.conf lands at
// etc/apps/my_app/local/inputs.conf, so the agent discovers it as an installed
// app rather than as system config.
func ConfDest(rel string) string {
	if !strings.Contains(rel, "/") {
		return filepath.Join("system", "local", rel)
	}
	return filepath.FromSlash(rel)
}
