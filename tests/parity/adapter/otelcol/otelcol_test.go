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
	"strings"
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
	// Neither a yaml nor a .conf tree is an error.
	if err := New("bin").Prepare(t.TempDir(), nil); err == nil {
		t.Error("expected error when no config source resolves")
	}
	// A staged file that is not on disk is an error.
	if err := New("bin").Prepare(t.TempDir(), []string{"inputs.conf"}); err == nil {
		t.Error("expected error for missing config dir")
	}
}

// TestPrepareMultipleConfigs covers the core behavior: every *.yaml/*.yml file
// becomes a --config (sorted). A file that is neither a collector config nor a
// .conf is installed into the sandbox home instead, and contributes no --config
// of its own.
func TestPrepareMultipleConfigs(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{ConfigFile, "override.yml", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	a := New("bin")
	if err := a.Prepare(dir, []string{ConfigFile, "override.yml", "notes.txt"}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	want := []string{
		"--config", filepath.Join(dir, ConfigFile), // config.yaml sorts before override.yml
		"--config", filepath.Join(dir, "override.yml"),
	}
	got := a.args()
	if len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestPrepareConfDrivenCase: a case that ships only .conf files is configured
// from the sandbox tree alone, addressed with forward slashes so the URI is
// usable on Windows too.
func TestPrepareConfDrivenCase(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inputs.conf"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	a := New("bin")
	if err := a.Prepare(dir, []string{"inputs.conf"}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	want := "splunkconf://" + filepath.ToSlash(filepath.Join(dir, splunkHomeDir, "etc"))
	if got := a.configs; len(got) != 1 || got[0] != want {
		t.Errorf("configs = %v, want [%q]", got, want)
	}
}

// TestFeatureGates: enabled gates follow the --config flags as a single
// comma-separated flag, each prefixed with +.
func TestFeatureGates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ConfigFile), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := New("bin").EnableFeatureGates("enableTARunner", "other")
	if err := a.Prepare(dir, []string{ConfigFile}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	want := []string{
		"--config", filepath.Join(dir, ConfigFile),
		"--feature-gates=+enableTARunner,+other",
	}
	got := a.args()
	if len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestSplunkHome: a case's .conf files land in etc/system/local of a sandbox
// tree the process gets as SPLUNK_HOME. They are never a --config source
// themselves; the provider reaches them through the tree's URI.
func TestSplunkHome(t *testing.T) {
	dir := t.TempDir()
	confs := []string{"inputs.conf", "outputs.conf"}
	for _, f := range append([]string{ConfigFile}, confs...) {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	a := New("bin")
	if err := a.Prepare(dir, append([]string{ConfigFile}, confs...)); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	home := filepath.Join(dir, splunkHomeDir)
	for _, name := range confs {
		if _, err := os.Stat(filepath.Join(home, "etc", "system", "local", name)); err != nil {
			t.Errorf("%s not installed: %v", name, err)
		}
	}
	want := []string{
		"--config", filepath.Join(dir, ConfigFile),
		"--config", "splunkconf://" + filepath.ToSlash(filepath.Join(home, "etc")),
	}
	if got := a.args(); len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("args[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	}
	if env := a.env(); len(env) == 0 || env[len(env)-1] != "SPLUNK_HOME="+home {
		t.Errorf("env does not end with SPLUNK_HOME=%s: %v", home, env)
	}
}

// TestNoConfNoSplunkHome: with no .conf files the child inherits the test's
// environment unchanged.
func TestNoConfNoSplunkHome(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ConfigFile), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := New("bin")
	if err := a.Prepare(dir, []string{ConfigFile}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if env := a.env(); env != nil {
		t.Errorf("env = %v, want nil", env)
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
	if err := a.Prepare(dir, []string{ConfigFile}); err != nil {
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

// A bare filename is system config; a nested path is mirrored, so an app in the
// case's conf/ is an installed app in the sandbox $SPLUNK_HOME too.
func TestInstallConfLayout(t *testing.T) {
	dir := t.TempDir()
	rels := []string{"inputs.conf", "apps/my_app/local/inputs.conf"}
	for _, rel := range rels {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	a := New("bin")
	if err := a.installHome(dir, rels, nil); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		filepath.Join("etc", "system", "local", "inputs.conf"),
		filepath.Join("etc", "apps", "my_app", "local", "inputs.conf"),
	} {
		if _, err := os.Stat(filepath.Join(a.splunkHome, want)); err != nil {
			t.Errorf("%s not installed: %v", want, err)
		}
	}
}

// A non-config file lands at the path the case named it, relative to
// $SPLUNK_HOME rather than to etc/, and is executable: splunkd runs a scripted
// input only from $SPLUNK_HOME/bin/scripts or a bundle's bin.
func TestInstallHomePayload(t *testing.T) {
	dir := t.TempDir()
	rels := []string{"bin/scripts/hello.sh", "etc/apps/my_app/bin/run.sh"}
	for _, rel := range rels {
		stage(t, dir, rel)
	}

	a := New("bin")
	if err := a.installHome(dir, nil, rels); err != nil {
		t.Fatal(err)
	}
	for _, rel := range rels {
		info, err := os.Stat(filepath.Join(a.splunkHome, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("%s not installed: %v", rel, err)
			continue
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Errorf("%s mode = %v, want the execute bit set", rel, info.Mode().Perm())
		}
	}
}

// splunkd derives $SPLUNK_DB from $SPLUNK_HOME and an input that checkpoints
// there refuses to start when it does not exist, so the sandbox home has it.
func TestInstallHomeCreatesSplunkDB(t *testing.T) {
	dir := t.TempDir()
	stage(t, dir, "inputs.conf")

	a := New("bin")
	if err := a.installHome(dir, []string{"inputs.conf"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(a.splunkHome, "var", "lib", "splunk")); err != nil {
		t.Errorf("var/lib/splunk not created: %v", err)
	}
}

// A payload file needs the sandbox home to live in, but configures nothing on
// its own, so it must not turn a collector-config case into a .conf one.
func TestPreparePayloadDoesNotAddConfRootURI(t *testing.T) {
	dir := t.TempDir()
	stage(t, dir, "config.yaml")
	stage(t, dir, "bin/scripts/hello.sh")

	a := New("bin")
	if err := a.Prepare(dir, []string{"bin/scripts/hello.sh", "config.yaml"}); err != nil {
		t.Fatal(err)
	}
	for _, c := range a.configs {
		if strings.HasPrefix(c, "splunkconf://") {
			t.Errorf("configs = %v, want no splunkconf:// source", a.configs)
		}
	}
	if a.splunkHome == "" {
		t.Error("splunkHome is empty, want the payload installed into a sandbox home")
	}
}

func stage(t *testing.T, dir, rel string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}
