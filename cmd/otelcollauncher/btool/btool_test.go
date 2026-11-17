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
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

func TestListMergesAllAppsAndReportsWinningSources(t *testing.T) {
	etcDir := t.TempDir()
	writeConf := func(relative, contents string) {
		t.Helper()
		writeFile(t, filepath.Join(etcDir, relative), contents)
	}
	const stanza = "monitor:///tmp/btool-smoke.log"
	writeConf("system/default/inputs.conf", "[default]\nhost = inherited\n["+stanza+"]\nvalue = system-default\n")
	writeConf("apps/B/default/inputs.conf", "["+stanza+"]\nvalue = b-default\n")
	writeConf("apps/B/local/inputs.conf", "["+stanza+"]\nvalue = b-local\n")
	writeConf("apps/A/local/inputs.conf", "["+stanza+"]\nvalue = a-local\n[monitor:///tmp/other.log]\nvalue = hidden-by-prefix\n")
	writeConf("system/local/inputs.conf", "[system-only]\nvalue = system-local\n")
	if err := os.MkdirAll(filepath.Join(etcDir, "apps/C/default"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := New().Dispatch(t.Context(), "btool", []string{"--dir=" + etcDir, "inputs", "list", "monitor:///tmp/btool"}, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	want := "[" + stanza + "]\nhost = inherited\nvalue = a-local\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected btool output:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(stdout.String(), "[monitor:///tmp/other.log]") {
		t.Fatal("stanza prefix did not filter unrelated stanza")
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

func TestUTF32ConfFilesFailInsteadOfBeingMisread(t *testing.T) {
	utf32le := string(utf32Bytes("[install]\nstate = disabled\n", []byte{0xff, 0xfe, 0, 0}, binary.LittleEndian))
	utf32be := string(utf32Bytes("[monitor:///tmp/a.log]\nindex = x\n", []byte{0, 0, 0xfe, 0xff}, binary.BigEndian))

	t.Run("conf layer", func(t *testing.T) {
		etcDir := t.TempDir()
		path := filepath.Join(etcDir, "system", "local", "inputs.conf")
		writeFile(t, path, utf32be)
		var stdout, stderr bytes.Buffer
		code := New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
		if code != cli.ExitFailure || stdout.Len() != 0 {
			t.Fatalf("exit code %d, stdout: %s", code, stdout.String())
		}
		if got := stderr.String(); !strings.Contains(got, path) || !strings.Contains(got, "UTF-32") {
			t.Fatalf("stderr does not name the file and encoding: %s", got)
		}
	})

	t.Run("app.conf", func(t *testing.T) {
		etcDir := t.TempDir()
		writeFile(t, filepath.Join(etcDir, "apps", "a", "default", "app.conf"), utf32le)
		writeFile(t, filepath.Join(etcDir, "apps", "a", "default", "inputs.conf"), "[monitor:///tmp/a.log]\nindex = x\n")
		var stdout, stderr bytes.Buffer
		code := New().Dispatch(t.Context(), "btool", []string{"inputs", "list", "--dir=" + etcDir}, cli.IO{Out: &stdout, Err: &stderr})
		if code != cli.ExitFailure || !strings.Contains(stderr.String(), "UTF-32") {
			t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
		}
	})
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runList(t *testing.T, args ...string) (string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := New().Dispatch(t.Context(), "btool", args, cli.IO{Out: &stdout, Err: &stderr})
	if code != cli.ExitOK {
		t.Logf("stderr: %s", stderr.String())
	}
	return stdout.String(), code
}

func TestSplunkEtcEnvironment(t *testing.T) {
	etcDir := t.TempDir()
	writeFile(t, filepath.Join(etcDir, "system", "local", "server.conf"), "[general]\nk = from-etc\n")
	other := t.TempDir()
	writeFile(t, filepath.Join(other, "etc", "system", "local", "server.conf"), "[general]\nk = from-home\n")

	t.Setenv("SPLUNK_HOME", other)
	t.Setenv("SPLUNK_ETC", etcDir)
	if out, code := runList(t, "server", "list"); code != cli.ExitOK || out != "[general]\nk = from-etc\n" {
		t.Fatalf("SPLUNK_ETC: code %d, output %q", code, out)
	}
	if out, code := runList(t, "--dir="+filepath.Join(other, "etc"), "server", "list"); code != cli.ExitOK || out != "[general]\nk = from-home\n" {
		t.Fatalf("--dir should win: code %d, output %q", code, out)
	}
	t.Setenv("SPLUNK_ETC", filepath.Join(etcDir, "missing"))
	if out, code := runList(t, "server", "list"); code != cli.ExitOK || out != "[general]\nk = from-home\n" {
		t.Fatalf("invalid SPLUNK_ETC should fall back: code %d, output %q", code, out)
	}
}

func TestParseArgsNativeSpellings(t *testing.T) {
	for arg, want := range map[string]bool{
		"true": true, "TRUE": true, "yes": true, "1": true, "t": true, "": true, "garbage": true,
		"false": false, "False": false, "no": false, "0": false, "f": false,
	} {
		options, _, err := parseArgs([]string{"--expand-stanzas=" + arg, "inputs", "list"})
		if err != nil || options.expand != want {
			t.Errorf("--expand-stanzas=%q: expand=%v err=%v, want %v", arg, options.expand, err, want)
		}
	}
	options, _, err := parseArgs([]string{"--dir=relative/etc", "inputs", "list", "prefix", "ignored"})
	if err != nil || !filepath.IsAbs(options.dir) || options.prefix != "prefix" {
		t.Fatalf("unexpected options %+v, err %v", options, err)
	}
	if _, _, err := parseArgs([]string{"inputs", "list", "a", "b", "c"}); err == nil {
		t.Fatal("five positionals should be rejected")
	}
}

func TestListFollowsSymlinkedApps(t *testing.T) {
	etcDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "real")
	writeFile(t, filepath.Join(target, "default", "server.conf"), "[general]\nk = linked\n")
	if err := os.MkdirAll(filepath.Join(etcDir, "apps"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(etcDir, "apps", "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if out, code := runList(t, "--dir="+etcDir, "server", "list"); code != cli.ExitOK || out != "[general]\nk = linked\n" {
		t.Fatalf("code %d, output %q", code, out)
	}
}

func TestDirOverridePreservesEnvironmentExpansion(t *testing.T) {
	home := t.TempDir()
	envEtcDir := t.TempDir()
	readDir := t.TempDir()
	appDir := filepath.Join(readDir, "apps", "A")

	for _, environment := range []struct {
		name        string
		home        string
		etcDir      string
		expandedEtc string
	}{
		{name: "relocated", home: home, etcDir: envEtcDir, expandedEtc: envEtcDir},
		{name: "home-default", home: home, expandedEtc: filepath.Join(home, "etc")},
		{name: "standalone", expandedEtc: readDir},
	} {
		t.Run(environment.name, func(t *testing.T) {
			t.Setenv("SPLUNK_HOME", environment.home)
			t.Setenv("SPLUNK_ETC", environment.etcDir)
			expandedHome := environment.home
			if expandedHome == "" {
				expandedHome = filepath.Dir(readDir)
			}
			for _, tc := range []struct {
				name  string
				input string
				want  string
			}{
				{name: "etc", input: "monitor://$SPLUNK_ETC/logs", want: "monitor://" + environment.expandedEtc + "/logs"},
				{name: "home-etc", input: "monitor://$SPLUNK_HOME/etc/apps", want: "monitor://" + environment.expandedEtc + "/apps"},
				{name: "home", input: "monitor://$SPLUNK_HOME/var/log", want: "monitor://" + expandedHome + "/var/log"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					writeFile(t, filepath.Join(appDir, "local", "inputs.conf"), "["+tc.input+"]\nk = v\n")
					want := "[" + tc.want + "]\nk = v\n"
					if out, code := runList(t, "--dir="+readDir, "inputs", "list"); code != cli.ExitOK || out != want {
						t.Fatalf("code %d, output %q, want %q", code, out, want)
					}
				})
			}
		})
	}
}
