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
	"sort"
	"strings"
	"syscall"
	"time"
)

// DefaultBin is the built binary used when PARITY_OTELCOL_BIN is unset. It is
// relative to the test working directory (tests/parity), pointing at the repo
// root's bin/otelcol produced by `make otelcol`.
const DefaultBin = "../../bin/otelcol"

// EnvBin overrides the otelcol binary location.
const EnvBin = "PARITY_OTELCOL_BIN"

// ConfigFile is the conventional filename for a case's primary collector
// config, supplied as an AgentRun.ConfigFiles entry. The adapter is not limited
// to it: it passes every *.yaml/*.yml file the runner writes into configDir as
// its own --config.
const ConfigFile = "config.yaml"

// splunkHomeDir is the sandbox $SPLUNK_HOME created inside the run's config
// directory.
const splunkHomeDir = "splunkhome"

// payloadMode is the mode a non-config case file is installed with. It carries
// the execute bit because a payload file is typically a script, and splunkd will
// not run one it cannot execute.
const payloadMode = 0o755

// Adapter drives one otelcol binary through a case.
type Adapter struct {
	cmd        *exec.Cmd
	bin        string
	splunkHome string   // sandbox $SPLUNK_HOME, set by Prepare when the case has .conf files
	gates      []string // feature gates to enable
	configs    []string // resolved --config sources, set by Prepare
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

// EnableFeatureGates turns on collector feature gates, passed as
// --feature-gates=+<id>.
func (a *Adapter) EnableFeatureGates(ids ...string) *Adapter {
	a.gates = append(a.gates, ids...)
	return a
}

func (a *Adapter) Name() string { return "otelcol" }

// InstallDir is the directory holding the binary.
func (a *Adapter) InstallDir() string { return filepath.Dir(a.bin) }

// Prepare collects every *.yaml/*.yml file the runner wrote into configDir as a
// --config source, sorted for a deterministic merge order. Everything else goes
// into a sandbox $SPLUNK_HOME (see installHome), and a case that ships a .conf
// file is configured from that tree through the splunkconf provider. At least
// one source must resolve.
//
// Only a .conf file adds the provider URI. A payload file needs the sandbox home
// to live in, but on its own it configures nothing, so a case shipping one
// alongside a collector config is still run from that config alone.
func (a *Adapter) Prepare(configDir string, files []string) error {
	var yamls, confs, payloads []string
	for _, rel := range files {
		switch filepath.Ext(rel) {
		case ".yaml", ".yml":
			yamls = append(yamls, filepath.Join(configDir, filepath.FromSlash(rel)))
		case ".conf":
			confs = append(confs, rel)
		default:
			payloads = append(payloads, rel)
		}
	}
	sort.Strings(yamls)
	if err := a.installHome(configDir, confs, payloads); err != nil {
		return err
	}
	a.configs = yamls
	if len(confs) > 0 {
		a.configs = append(a.configs, a.confRootURI())
	}
	if len(a.configs) == 0 {
		return fmt.Errorf("no collector config: want a *.yaml file or a .conf tree in %s", configDir)
	}
	return nil
}

// confRootURI addresses the sandbox conf tree for the splunkconf provider. The
// path is forward-slashed because a URI authority rejects the backslashes of a
// native Windows path.
func (a *Adapter) confRootURI() string {
	return "splunkconf://" + filepath.ToSlash(filepath.Join(a.splunkHome, "etc"))
}

// installHome lays out the sandbox $SPLUNK_HOME the case is run from, which
// Start then puts in the environment.
//
// A .conf file goes into etc/: a bare filename is system config, a nested path
// is mirrored, so an app in the case's conf/ is an installed app here too. Every
// other file lands at the path the case named it, relative to $SPLUNK_HOME
// itself, because that is the only way to reach the directories outside etc/
// that splunkd reads. A scripted input is the case in point: splunkd runs a
// script only from $SPLUNK_HOME/bin/scripts or a bundle's bin, so bin/scripts/x
// in a case installs there rather than under etc/.
func (a *Adapter) installHome(configDir string, confs, payloads []string) error {
	if len(confs) == 0 && len(payloads) == 0 {
		return nil
	}
	home := filepath.Join(configDir, splunkHomeDir)
	etcDir := filepath.Join(home, "etc")
	for _, rel := range confs {
		dest := filepath.Join(etcDir, "system", "local", rel)
		if strings.Contains(rel, "/") {
			dest = filepath.Join(etcDir, filepath.FromSlash(rel))
		}
		if err := copyInto(configDir, rel, dest, 0o600); err != nil {
			return err
		}
	}
	for _, rel := range payloads {
		if err := copyInto(configDir, rel, filepath.Join(home, filepath.FromSlash(rel)), payloadMode); err != nil {
			return err
		}
	}
	// A Splunk install holds its data under var/lib/splunk, and an input that
	// checkpoints there refuses to start when $SPLUNK_DB does not resolve to an
	// existing directory. The sandbox is a $SPLUNK_HOME like any other, so it
	// has the directory too.
	if err := os.MkdirAll(filepath.Join(home, "var", "lib", "splunk"), 0o755); err != nil {
		return err
	}
	a.splunkHome = home
	return nil
}

// copyInto copies the staged file rel into dest with mode. The contents are
// written restrictively and then chmod'ed, rather than written with mode
// directly, so an execute bit never passes through os.WriteFile.
func copyInto(configDir, rel, dest string, mode os.FileMode) error {
	contents, err := os.ReadFile(filepath.Join(configDir, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dest, contents, 0o600); err != nil {
		return err
	}
	return os.Chmod(dest, mode)
}

// args expands the resolved sources into repeated --config flags, followed by
// the enabled feature gates.
func (a *Adapter) args() []string {
	args := make([]string, 0, 2*len(a.configs)+1)
	for _, c := range a.configs {
		args = append(args, "--config", c)
	}
	if len(a.gates) > 0 {
		enabled := make([]string, len(a.gates))
		for i, g := range a.gates {
			enabled[i] = "+" + g
		}
		args = append(args, "--feature-gates="+strings.Join(enabled, ","))
	}
	return args
}

// env is the test's own environment plus $SPLUNK_HOME when a tree was
// installed; nil means the child inherits it unchanged.
func (a *Adapter) env() []string {
	if a.splunkHome == "" {
		return nil
	}
	return append(os.Environ(), "SPLUNK_HOME="+a.splunkHome)
}

func (a *Adapter) Start(ctx context.Context) error {
	if _, err := os.Stat(a.bin); err != nil {
		return fmt.Errorf("otelcol binary %s not found (build it with `make otelcol`, or set %s): %w", a.bin, EnvBin, err)
	}
	cmd := exec.CommandContext(ctx, a.bin, a.args()...) //nolint:gosec // G204: binary path and config sources are test-controlled inputs
	cmd.Env = a.env()
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
