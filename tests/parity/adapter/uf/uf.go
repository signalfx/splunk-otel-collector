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

// Package uf is the parity Adapter for the Splunk Universal Forwarder, the
// parity oracle. It installs a case's .conf files into the UF's
// etc/system/local layout (restoring prior state on cleanup) and drives the
// agent through bin/splunk start/stop.
package uf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvInstallDir sets the UF install location. There is no built-in default: a
// run without it set has no UF to drive, so the test skips.
const EnvInstallDir = "PARITY_UF_DIR"

// Adapter drives one UF install through a case.
type Adapter struct {
	installDir string
	restores   []func() error
}

// New returns a UF adapter. installDir may be empty, in which case
// PARITY_UF_DIR is used. If that is also unset, InstallDir is empty and the
// caller is expected to skip.
func New(installDir string) *Adapter {
	if installDir == "" {
		installDir = os.Getenv(EnvInstallDir)
	}
	return &Adapter{installDir: installDir}
}

func (a *Adapter) Name() string       { return "UF" }
func (a *Adapter) InstallDir() string { return a.installDir }

// Prepare copies every .conf file from configDir into etc/system/local,
// recording a restore for each so the install is left as it was found.
func (a *Adapter) Prepare(configDir string) error {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}
	localDir := filepath.Join(a.installDir, "etc", "system", "local")
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(configDir, e.Name()))
		if err != nil {
			return err
		}
		if err := a.writeWithRestore(filepath.Join(localDir, e.Name()), contents); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) Start(ctx context.Context) error {
	// --accept-license/--answer-yes/--no-prompt are idempotent on an already
	// initialized install and required on a first start.
	return a.run(ctx, "start", "--accept-license", "--answer-yes", "--no-prompt")
}

func (a *Adapter) Stop(ctx context.Context) error {
	return a.run(ctx, "stop")
}

func (a *Adapter) Cleanup() error {
	var firstErr error
	// Restore in reverse order so overlapping writes unwind cleanly.
	for i := len(a.restores) - 1; i >= 0; i-- {
		if err := a.restores[i](); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	a.restores = nil
	return firstErr
}

func (a *Adapter) run(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, a.executable(), args...) //nolint:gosec // G204: the UF binary path and args are test-controlled inputs
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("splunk %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return nil
}

func (a *Adapter) executable() string {
	name := "splunk"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(a.installDir, "bin", name)
}

// writeWithRestore writes contents to path and records a cleanup that restores
// the prior contents (or removes the file if it did not exist). Ported from the
// 1spl writeFileWithRestore.
func (a *Adapter) writeWithRestore(path string, contents []byte) error {
	info, statErr := os.Stat(path)
	existed := statErr == nil
	var original []byte
	if existed {
		var err error
		if original, err = os.ReadFile(path); err != nil {
			return err
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return err
	}

	a.restores = append(a.restores, func() error {
		if existed {
			return os.WriteFile(path, original, info.Mode())
		}
		return os.Remove(path)
	})
	return nil
}
