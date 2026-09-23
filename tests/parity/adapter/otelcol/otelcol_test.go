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

package otelcol

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
	if got := New("/some/bin/otelcol").InstallDir(); got != filepath.Dir("/some/bin/otelcol") {
		t.Errorf("explicit bin InstallDir = %q", got)
	}

	t.Setenv(EnvBin, "/env/otelcol")
	if got := New(""); got.bin != "/env/otelcol" {
		t.Errorf("env bin = %q", got.bin)
	}

	os.Unsetenv(EnvBin)
	if got := New(""); got.bin != DefaultBin {
		t.Errorf("default bin = %q, want %q", got.bin, DefaultBin)
	}
}

func TestName(t *testing.T) {
	if got := New("x").Name(); got != "otelcol" {
		t.Errorf("Name = %q", got)
	}
}

func TestPrepareNoSources(t *testing.T) {
	// No yaml files and no extra sources is an error.
	if err := New("bin").Prepare(t.TempDir()); err == nil {
		t.Error("expected error when no config source resolves")
	}
	// A missing config dir is an error.
	if err := New("bin").Prepare(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for missing config dir")
	}
}

// TestPrepareMultipleConfigs covers the core behavior: every *.yaml/*.yml file
// becomes a --config (sorted), non-yaml files are ignored, and extra sources
// from New are appended after the rendered files.
func TestPrepareMultipleConfigs(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{ConfigFile, "override.yml", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	a := New("bin", "splunkhome://SPLUNK_HOME?pipeline=uf")
	if err := a.Prepare(dir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	want := []string{
		"--config", filepath.Join(dir, ConfigFile), // config.yaml sorts before override.yml
		"--config", filepath.Join(dir, "override.yml"),
		"--config", "splunkhome://SPLUNK_HOME?pipeline=uf",
	}
	got := a.configArgs()
	if len(got) != len(want) {
		t.Fatalf("configArgs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("configArgs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestPrepareExtraOnly: an extra source alone is enough, even with no yaml files.
func TestPrepareExtraOnly(t *testing.T) {
	a := New("bin", "splunkhome://SPLUNK_HOME")
	if err := a.Prepare(t.TempDir()); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if want := []string{"--config", "splunkhome://SPLUNK_HOME"}; len(a.configArgs()) != len(want) {
		t.Errorf("configArgs = %v, want %v", a.configArgs(), want)
	}
}

func TestStartBinNotFound(t *testing.T) {
	a := New(filepath.Join(t.TempDir(), "does-not-exist"))
	if err := a.Start(context.Background()); err == nil {
		t.Error("expected error when the binary does not exist")
	}
}

func TestStopNoProcess(t *testing.T) {
	if err := New("bin").Stop(context.Background()); err != nil {
		t.Errorf("Stop with no process = %v, want nil", err)
	}
}

// TestStartStop launches a real long-running stub as the "binary" and confirms
// Stop terminates it. The stub is a shell script, so this runs on POSIX only.
func TestStartStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub is POSIX only")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "otelcol")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexec sleep 30\n"), 0o700); err != nil { //nolint:gosec // test stub
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ConfigFile), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	a := New(bin)
	if err := a.Prepare(dir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if a.cmd == nil || a.cmd.Process == nil {
		t.Fatal("process not started")
	}
	if err := a.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if a.cmd != nil {
		t.Error("cmd should be cleared after Stop")
	}
}

func TestCleanup(t *testing.T) {
	if err := New("x").Cleanup(); err != nil {
		t.Errorf("Cleanup = %v, want nil", err)
	}
}
