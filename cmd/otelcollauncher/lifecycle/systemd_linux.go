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

//go:build linux

package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const (
	serviceUnitName   = "splunk-otel-collector.service"
	systemdRuntimeDir = "/run/systemd/system"
)

type systemctlRunner interface {
	output(ctx context.Context, args ...string) ([]byte, error)
	run(ctx context.Context, streams cli.IO, args ...string) (int, error)
}

type execSystemctl struct {
	binary string
}

func (r execSystemctl) output(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.binary, args...) //nolint:gosec // G204: r.binary is "systemctl"; args are launcher-constructed, not user input
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s %s: %w", r.binary, strings.Join(args, " "), err)
	}
	return out, nil
}

func (r execSystemctl) run(ctx context.Context, streams cli.IO, args ...string) (int, error) {
	cmd := exec.CommandContext(ctx, r.binary, args...) //nolint:gosec // G204: r.binary is "systemctl"; args are launcher-constructed, not user input
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.Err
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), nil
	default:
		return cli.ExitFailure, fmt.Errorf("failed to run %s %s: %w", r.binary, strings.Join(args, " "), err)
	}
}

type systemdProxy struct {
	runner     systemctlRunner
	runtimeDir string
	unit       string
}

func newSystemdProxy() *systemdProxy {
	return &systemdProxy{
		runner:     execSystemctl{binary: "systemctl"},
		runtimeDir: systemdRuntimeDir,
		unit:       serviceUnitName,
	}
}

func (p *systemdProxy) booted() bool {
	_, err := os.Stat(p.runtimeDir)
	return err == nil
}

func (p *systemdProxy) loadState(ctx context.Context) (string, error) {
	out, err := p.runner.output(ctx, "show", p.unit, "--property=LoadState", "--no-pager")
	if err != nil {
		return "", err
	}
	line, _, _ := strings.Cut(string(out), "\n")
	_, value, _ := strings.Cut(strings.TrimRight(line, "\r"), "=")
	return value, nil
}

func (p *systemdProxy) managesUnit(ctx context.Context) (bool, error) {
	if !p.booted() {
		return false, nil
	}
	state, err := p.loadState(ctx)
	if err != nil {
		return false, err
	}
	return state == "loaded", nil
}

func (p *systemdProxy) dispatch(ctx context.Context, v verb, args []string, streams cli.IO) int {
	if len(args) > 0 {
		fmt.Fprintf(streams.Err, "warning: ignoring arguments %q; systemd manages splunk-otel-collector\n", args)
	}
	var systemctlArgs []string
	switch v {
	case verbStart, verbStop, verbRestart:
		systemctlArgs = []string{string(v), p.unit}
	case verbStatus:
		systemctlArgs = []string{"status", p.unit, "--no-pager"}
	default:
		fmt.Fprintf(streams.Err, "unknown lifecycle command %q\n", string(v))
		return cli.ExitUsage
	}

	code, err := p.runner.run(ctx, streams, systemctlArgs...)
	if err != nil {
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	return code
}
