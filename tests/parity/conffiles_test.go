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

package parity_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/parity"
)

func TestConfFiles(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range []string{
		"inputs.conf",
		"outputs.conf",
		"apps/my_app/local/inputs.conf",
		"apps/my_app/default/app.conf",
		"notes.txt", // not a .conf, must be ignored
	} {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := parity.ConfFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"apps/my_app/default/app.conf",
		"apps/my_app/local/inputs.conf",
		"inputs.conf",
		"outputs.conf",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

// The runner always creates the config dir before Prepare, so an absent one
// means something upstream is broken and must not read as "no conf files".
func TestConfFilesMissingDir(t *testing.T) {
	if _, err := parity.ConfFiles(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("expected an error for a missing dir")
	}
}
