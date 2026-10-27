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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

func TestListMergesAllAppsAndReportsWinningSources(t *testing.T) {
	etcDir := t.TempDir()
	writeConf := func(relative, contents string) string {
		t.Helper()
		path := filepath.Join(etcDir, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	const stanza = "monitor:///tmp/btool-smoke.log"
	systemDefault := writeConf("system/default/inputs.conf", "[default]\nhost = inherited\n["+stanza+"]\nvalue = system-default\n")
	writeConf("apps/B/default/inputs.conf", "["+stanza+"]\nvalue = b-default\n")
	writeConf("apps/B/local/inputs.conf", "["+stanza+"]\nvalue = b-local\n")
	aLocal := writeConf("apps/A/local/inputs.conf", "["+stanza+"]\nvalue = a-local\n[monitor:///tmp/other.log]\nvalue = hidden-by-prefix\n")
	writeConf("system/local/inputs.conf", "[system-only]\nvalue = system-local\n")
	if err := os.MkdirAll(filepath.Join(etcDir, "apps/C/default"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := New().Dispatch(t.Context(), "btool", []string{"--debug", "--dir=" + etcDir, "inputs", "list", "monitor:///tmp/btool"}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	pathWidth := max(len(aLocal), len(systemDefault))
	want := fmt.Sprintf("%-*s [%s]\n%-*s host = inherited\n%-*s value = a-local\n", pathWidth, aLocal, stanza, pathWidth, systemDefault, pathWidth, aLocal)
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected btool output:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(stdout.String(), "[monitor:///tmp/other.log]") {
		t.Fatal("stanza prefix did not filter unrelated stanza")
	}

	stdout.Reset()
	stderr.Reset()
	code = New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK || strings.Contains(stdout.String(), "[default]") {
		t.Fatalf("inputs.conf default stanza was displayed: exit code %d, output: %s", code, stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "monitor:///tmp/btool", "--app=B", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK {
		t.Fatalf("app context exit code %d, stderr: %s", code, stderr.String())
	}
	if got, want := stdout.String(), "["+stanza+"]\nvalue = b-local\n"; got != want {
		t.Fatalf("unexpected app context output:\n%s\nwant:\n%s", got, want)
	}

	stdout.Reset()
	stderr.Reset()
	code = New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "--app=C", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK {
		t.Fatalf("other app context exit code %d, stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("empty app context included settings from another app: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "--app=missing", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitFailure || !strings.Contains(stderr.String(), "application does not exist") {
		t.Fatalf("missing app: exit code %d, stderr: %s", code, stderr.String())
	}
}
