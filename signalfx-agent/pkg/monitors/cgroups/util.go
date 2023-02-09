// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cgroups

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func withOpenFile(path string, handler func(fd *os.File)) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer fd.Close()
	handler(fd)

	return nil
}

func parseSingleInt(fileReader io.Reader) (int64, error) {
	bytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return 0, err
	}

	out, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func walkControllerHierarchy(root string, cb func(cgroupName string, files []string)) {
	dirs := []string{root}

	var currentDir string
	for len(dirs) > 0 {
		currentDir, dirs = dirs[0], dirs[1:]
		f, err := os.Open(currentDir)
		if err != nil {
			continue
		}
		defer f.Close()

		contents, err := f.Readdir(0)
		if err != nil {
			continue
		}

		var files []string

		for i := range contents {
			name := contents[i].Name()
			if name == "." || name == ".." {
				continue
			}
			if contents[i].IsDir() {
				dirs = append(dirs, filepath.Join(currentDir, name))
			} else {
				files = append(files, filepath.Join(currentDir, name))
			}
		}

		cgroupName, err := filepath.Rel(root, currentDir)
		if err != nil {
			continue
		}

		cb("/"+strings.TrimPrefix(strings.TrimPrefix(cgroupName, "."), "/"), files)
	}
}
