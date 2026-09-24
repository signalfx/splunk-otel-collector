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
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

type fakeSystemctl struct {
	outputErr    error
	runErr       error
	outputResult []byte
	calls        [][]string
	runCode      int
}

func (f *fakeSystemctl) output(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string{}, args...))
	return f.outputResult, f.outputErr
}

func (f *fakeSystemctl) run(_ context.Context, _ cli.IO, args ...string) (int, error) {
	f.calls = append(f.calls, append([]string{}, args...))
	return f.runCode, f.runErr
}

func TestLoadState(t *testing.T) {
	tests := map[string]struct {
		out  string
		want string
	}{
		"loaded":               {out: "LoadState=loaded\n", want: "loaded"},
		"not found":            {out: "LoadState=not-found\n", want: "not-found"},
		"masked":               {out: "LoadState=masked\n", want: "masked"},
		"trailing CRLF":        {out: "LoadState=loaded\r\n", want: "loaded"},
		"no trailing newline":  {out: "LoadState=loaded", want: "loaded"},
		"value containing '='": {out: "LoadState=loaded=weird\n", want: "loaded=weird"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeSystemctl{outputResult: []byte(tt.out)}
			p := &systemdProxy{runner: fake, unit: serviceUnitName}

			got, err := p.loadState(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			require.Len(t, fake.calls, 1)
			assert.Equal(t, []string{"show", serviceUnitName, "--property=LoadState", "--no-pager"}, fake.calls[0])
		})
	}
}

func TestExecSystemctlOutputIncludesStderr(t *testing.T) {
	r := execSystemctl{binary: "sh"}

	_, err := r.output(context.Background(), "-c", "echo 'Failed to connect to bus: No such file or directory' >&2; exit 1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to connect to bus")
}

func TestSystemdProxyBooted(t *testing.T) {
	t.Run("runtime dir present", func(t *testing.T) {
		dir := t.TempDir()
		p := &systemdProxy{runtimeDir: dir}
		assert.True(t, p.booted())
	})

	t.Run("runtime dir absent", func(t *testing.T) {
		p := &systemdProxy{runtimeDir: filepath.Join(t.TempDir(), "does-not-exist")}
		assert.False(t, p.booted())
	})
}

func TestSystemdProxyManagesUnit(t *testing.T) {
	// ActiveState and UnitFileState are not even requested from systemctl:
	// LoadState alone decides this predicate, so a stopped-but-installed
	// unit (still "loaded") must still route through systemd, and only the
	// LoadState value itself varies across these cases.
	tests := map[string]struct {
		loadState string
		want      bool
	}{
		"loaded":    {loadState: "loaded", want: true},
		"not-found": {loadState: "not-found", want: false},
		"masked":    {loadState: "masked", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			fake := &fakeSystemctl{outputResult: []byte("LoadState=" + tt.loadState + "\n")}
			p := &systemdProxy{runner: fake, runtimeDir: dir, unit: serviceUnitName}

			managed, err := p.managesUnit(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tt.want, managed)
			assert.Len(t, fake.calls, 1)
		})
	}

	t.Run("not booted short-circuits without invoking the runner", func(t *testing.T) {
		fake := &fakeSystemctl{outputResult: []byte("LoadState=loaded\n")}
		p := &systemdProxy{runner: fake, runtimeDir: filepath.Join(t.TempDir(), "absent"), unit: serviceUnitName}

		managed, err := p.managesUnit(context.Background())
		require.NoError(t, err)
		assert.False(t, managed)
		assert.Empty(t, fake.calls, "systemctl must not be invoked when systemd is not booted")
	})
}

func TestSystemdProxyManagesUnitReturnsErrorWhenShowFails(t *testing.T) {
	fake := &fakeSystemctl{outputErr: errors.New("systemctl: connection refused")}
	p := &systemdProxy{runner: fake, runtimeDir: t.TempDir(), unit: serviceUnitName}

	managed, err := p.managesUnit(context.Background())
	require.Error(t, err)
	assert.False(t, managed)
}

func TestSystemdProxyDispatch(t *testing.T) {
	tests := map[string]struct {
		verb     verb
		wantArgs []string
	}{
		"start":   {verb: verbStart, wantArgs: []string{"start", serviceUnitName}},
		"stop":    {verb: verbStop, wantArgs: []string{"stop", serviceUnitName}},
		"restart": {verb: verbRestart, wantArgs: []string{"restart", serviceUnitName}},
		"status":  {verb: verbStatus, wantArgs: []string{"status", serviceUnitName, "--no-pager"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeSystemctl{runCode: 0}
			p := &systemdProxy{runner: fake, unit: serviceUnitName}
			streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

			code := p.dispatch(context.Background(), tt.verb, nil, streams)

			assert.Equal(t, cli.ExitOK, code)
			require.Len(t, fake.calls, 1)
			assert.Equal(t, tt.wantArgs, fake.calls[0])
		})
	}
}

func TestSystemdProxyDispatchPropagatesExitCode(t *testing.T) {
	fake := &fakeSystemctl{runCode: 3}
	p := &systemdProxy{runner: fake, unit: serviceUnitName}
	streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	code := p.dispatch(context.Background(), verbStatus, nil, streams)

	assert.Equal(t, 3, code)
}

func TestSystemdProxyDispatchStreamsWarningForIgnoredArgs(t *testing.T) {
	fake := &fakeSystemctl{runCode: 0}
	p := &systemdProxy{runner: fake, unit: serviceUnitName}
	errBuf := &bytes.Buffer{}
	streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

	code := p.dispatch(context.Background(), verbStart, []string{"--config", "/x"}, streams)

	assert.Equal(t, cli.ExitOK, code)
	assert.Contains(t, errBuf.String(), "ignoring arguments")
}
