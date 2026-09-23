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

// Package otelcol is the parity Adapter for the Splunk OTel Collector, run as a
// candidate. It runs the already-built otelcol binary against the case's
// rendered config.yaml, mirroring how testutils.CollectorProcess drives a local
// binary. The binary must be built first (make otelcol); this adapter does not
// build it.
package otelcol

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// DefaultBin is the built binary used when PARITY_OTELCOL_BIN is unset. It is
// relative to the test working directory (tests/parity), pointing at the repo
// root's bin/otelcol produced by `make otelcol`.
const DefaultBin = "../../bin/otelcol"

// EnvBin overrides the otelcol binary location.
const EnvBin = "PARITY_OTELCOL_BIN"

// ConfigFile is the filename the adapter runs with --config; the case must
// supply an AgentRun.ConfigFiles entry under this name.
const ConfigFile = "config.yaml"

// Adapter drives one otelcol binary through a case.
type Adapter struct {
	cmd        *exec.Cmd
	bin        string
	configPath string
}

// New returns an otelcol adapter. bin may be empty, in which case
// PARITY_OTELCOL_BIN or DefaultBin is used.
func New(bin string) *Adapter {
	if bin == "" {
		bin = os.Getenv(EnvBin)
	}
	if bin == "" {
		bin = DefaultBin
	}
	return &Adapter{bin: bin}
}

func (a *Adapter) Name() string { return "otelcol" }

// InstallDir is the directory holding the binary, exposed as AGENT_DIR.
func (a *Adapter) InstallDir() string { return filepath.Dir(a.bin) }

// Prepare records the rendered config path. The runner has already written the
// ConfigFiles into configDir.
func (a *Adapter) Prepare(configDir string) error {
	p := filepath.Join(configDir, ConfigFile)
	if _, err := os.Stat(p); err != nil {
		return fmt.Errorf("config %s: %w", ConfigFile, err)
	}
	a.configPath = p
	return nil
}

func (a *Adapter) Start(ctx context.Context) error {
	if _, err := os.Stat(a.bin); err != nil {
		return fmt.Errorf("otelcol binary %s not found (build it with `make otelcol`, or set %s): %w", a.bin, EnvBin, err)
	}
	cmd := exec.CommandContext(ctx, a.bin, "--config", a.configPath) //nolint:gosec // G204: binary path and config are test-controlled inputs
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start otelcol: %w", err)
	}
	a.cmd = cmd
	return nil
}

// Stop signals the collector to shut down and waits briefly for it to exit
// before killing it.
func (a *Adapter) Stop(_ context.Context) error {
	if a.cmd == nil || a.cmd.Process == nil {
		return nil
	}
	_ = a.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		_ = a.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		_ = a.cmd.Process.Kill()
		<-done
	}
	a.cmd = nil
	return nil
}

func (a *Adapter) Cleanup() error { return nil }
