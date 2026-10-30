// Copyright Splunk Inc.
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

package btool

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

// debugPathColumn collapses the padding after a source path in --debug
// output, which depends on the length of the fixture's location.
var debugPathColumn = regexp.MustCompile(`(?m)^(@ETC@\S*) +`)

// TestParity compares the output of the command with the output recorded from
// a real `splunk btool` for each case in testdata/parity/cases. The expected
// files are produced by testdata/parity/record.sh. Error text is not compared,
// only the exit code, since the wording and output stream differ.
func TestParity(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("recorded outputs use forward slashes")
	}
	casesDir, err := filepath.Abs(filepath.Join("testdata", "parity", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no parity cases found")
	}

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			caseDir := filepath.Join(casesDir, entry.Name())
			etcDir := filepath.Join(caseDir, "etc")
			args := readLines(t, filepath.Join(caseDir, "args"))
			expected := readLines(t, filepath.Join(caseDir, "expected"))
			wantCode, err := strconv.Atoi(strings.TrimPrefix(expected[len(expected)-1], "EXIT="))
			if err != nil {
				t.Fatalf("malformed expected file: %v", err)
			}

			// The recordings were made with Splunk installed in /opt/splunk.
			t.Setenv("SPLUNK_HOME", "/opt/splunk")
			t.Setenv("SPLUNK_DB", "")

			var stdout, stderr bytes.Buffer
			code := New().Dispatch(t.Context(), "btool", append([]string{"--dir=" + etcDir}, args...), cli.IO{Out: &stdout, Err: &stderr})
			if code != wantCode {
				t.Fatalf("exit code %d, want %d; stderr: %s", code, wantCode, stderr.String())
			}
			if wantCode != 0 {
				return
			}

			got := debugPathColumn.ReplaceAllString(strings.ReplaceAll(stdout.String(), etcDir, "@ETC@"), "$1 ")
			want := strings.Join(expected[:len(expected)-1], "\n")
			if want != "" {
				want += "\n"
			}
			if got != want {
				t.Fatalf("output differs from splunk btool\n--- got\n%s--- want\n%s", got, want)
			}
		})
	}
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
}
