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

package uf

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/parity"
)

var _ parity.Adapter = (*Adapter)(nil)

func TestNew(t *testing.T) {
	if got := New("/opt/uf").InstallDir(); got != "/opt/uf" {
		t.Errorf("explicit dir = %q", got)
	}

	t.Setenv(EnvInstallDir, "/from/env")
	if got := New("").InstallDir(); got != "/from/env" {
		t.Errorf("env dir = %q", got)
	}

	// Explicit dir wins over the environment.
	if got := New("/explicit").InstallDir(); got != "/explicit" {
		t.Errorf("explicit over env = %q", got)
	}
}

func TestName(t *testing.T) {
	if got := New("/opt/uf").Name(); got != "UF" {
		t.Errorf("Name = %q", got)
	}
}

func TestExecutable(t *testing.T) {
	got := New("/opt/uf").executable()
	want := filepath.Join("/opt/uf", "bin", "splunk")
	if runtime.GOOS == "windows" {
		want += ".exe"
	}
	if got != want {
		t.Errorf("executable = %q, want %q", got, want)
	}
}

// TestPrepareCopiesConfsAndCleanup covers the whole install/restore cycle:
// Prepare copies only .conf files into etc/system/local, and Cleanup removes
// files that did not exist before while restoring the original contents of one
// that did.
func TestPrepareCopiesConfsAndCleanup(t *testing.T) {
	install := t.TempDir()
	localDir := filepath.Join(install, "etc", "system", "local")
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// A pre-existing inputs.conf that Prepare will overwrite and Cleanup restore.
	if err := os.WriteFile(filepath.Join(localDir, "inputs.conf"), []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	configDir := t.TempDir()
	writeFile(t, filepath.Join(configDir, "inputs.conf"), "new-inputs")
	writeFile(t, filepath.Join(configDir, "outputs.conf"), "new-outputs")
	writeFile(t, filepath.Join(configDir, "notes.txt"), "ignored") // not a .conf
	if err := os.Mkdir(filepath.Join(configDir, "sub.conf"), 0o700); err != nil {
		t.Fatal(err) // a dir named *.conf must be skipped
	}

	a := New(install)
	if err := a.Prepare(configDir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	if got := readFile(t, filepath.Join(localDir, "inputs.conf")); got != "new-inputs" {
		t.Errorf("inputs.conf = %q", got)
	}
	if got := readFile(t, filepath.Join(localDir, "outputs.conf")); got != "new-outputs" {
		t.Errorf("outputs.conf = %q", got)
	}
	if _, err := os.Stat(filepath.Join(localDir, "notes.txt")); !os.IsNotExist(err) {
		t.Error("non-conf file should not be copied")
	}
	if _, err := os.Stat(filepath.Join(localDir, "sub.conf")); !os.IsNotExist(err) {
		t.Error("directory named *.conf should be skipped")
	}

	if err := a.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	// Pre-existing file restored to its original contents.
	if got := readFile(t, filepath.Join(localDir, "inputs.conf")); got != "original" {
		t.Errorf("inputs.conf after cleanup = %q, want %q", got, "original")
	}
	// Newly created file removed.
	if _, err := os.Stat(filepath.Join(localDir, "outputs.conf")); !os.IsNotExist(err) {
		t.Error("outputs.conf should be removed on cleanup")
	}
	// Cleanup is idempotent: a second call is a no-op.
	if err := a.Cleanup(); err != nil {
		t.Errorf("second Cleanup: %v", err)
	}
}

func TestPrepareMissingConfigDir(t *testing.T) {
	a := New(t.TempDir())
	if err := a.Prepare(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for missing config dir")
	}
}

// TestStartStop puts a shell-script stub at bin/splunk and drives it through
// Start and Stop, so the exec path (run) is exercised without a real UF. POSIX
// only, since the stub is a shell script.
func TestStartStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub is POSIX only")
	}
	install := t.TempDir()
	binDir := filepath.Join(install, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Echoes its args (so `start ...` and `stop` both exit 0), used for both.
	if err := os.WriteFile(filepath.Join(binDir, "splunk"), []byte("#!/bin/sh\necho \"$@\"\n"), 0o700); err != nil { //nolint:gosec // test stub
		t.Fatal(err)
	}

	a := New(install)
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := a.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestRunError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub is POSIX only")
	}
	install := t.TempDir()
	binDir := filepath.Join(install, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "splunk"), []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil { //nolint:gosec // test stub
		t.Fatal(err)
	}
	if err := New(install).Start(context.Background()); err == nil {
		t.Error("expected error when splunk exits non-zero")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
