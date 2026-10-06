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

//go:build windows

package lifecycle

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

type fakeService struct {
	startErr       error
	controlErr     error
	queryErr       error
	queryResponses []svc.Status
	controlCalls   []svc.Cmd
	queryIdx       int
	startCalls     int
	closed         bool
}

func (f *fakeService) Start(_ ...string) error {
	f.startCalls++
	return f.startErr
}

func (f *fakeService) Control(c svc.Cmd) (svc.Status, error) {
	f.controlCalls = append(f.controlCalls, c)
	if f.controlErr != nil {
		return svc.Status{}, f.controlErr
	}
	return f.currentStatus(), nil
}

func (f *fakeService) Query() (svc.Status, error) {
	if f.queryErr != nil {
		return svc.Status{}, f.queryErr
	}
	return f.nextStatus(), nil
}

func (f *fakeService) currentStatus() svc.Status {
	if len(f.queryResponses) == 0 {
		return svc.Status{}
	}
	idx := f.queryIdx
	if idx >= len(f.queryResponses) {
		idx = len(f.queryResponses) - 1
	}
	return f.queryResponses[idx]
}

func (f *fakeService) nextStatus() svc.Status {
	st := f.currentStatus()
	if f.queryIdx < len(f.queryResponses) {
		f.queryIdx++
	}
	return st
}

func (f *fakeService) Close() error {
	f.closed = true
	return nil
}

func notInstalledErr() error {
	return fmt.Errorf("failed to open service %q: %w", windowsServiceName, windows.ERROR_SERVICE_DOES_NOT_EXIST)
}

func TestServiceNotInstalled(t *testing.T) {
	assert.True(t, serviceNotInstalled(notInstalledErr()))
	assert.False(t, serviceNotInstalled(fmt.Errorf("some other error: %w", windows.ERROR_ACCESS_DENIED)))
	assert.False(t, serviceNotInstalled(nil))
}

func TestStateName(t *testing.T) {
	tests := map[string]struct {
		want  string
		state svc.State
	}{
		"stopped":           {state: svc.Stopped, want: "stopped"},
		"start pending":     {state: svc.StartPending, want: "starting"},
		"stop pending":      {state: svc.StopPending, want: "stopping"},
		"running":           {state: svc.Running, want: "running"},
		"continue pending":  {state: svc.ContinuePending, want: "resuming"},
		"pause pending":     {state: svc.PausePending, want: "pausing"},
		"paused":            {state: svc.Paused, want: "paused"},
		"unknown state 255": {state: svc.State(255), want: "unknown (255)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, stateName(tt.state))
		})
	}
}

func TestWaitForState(t *testing.T) {
	t.Run("reaches state before timeout", func(t *testing.T) {
		calls := 0
		query := func() (svc.Status, error) {
			calls++
			if calls < 3 {
				return svc.Status{State: svc.StartPending}, nil
			}
			return svc.Status{State: svc.Running}, nil
		}

		err := waitForState(query, svc.Running, time.Second, 5*time.Millisecond)
		require.NoError(t, err)
	})

	t.Run("times out", func(t *testing.T) {
		query := func() (svc.Status, error) {
			return svc.Status{State: svc.StartPending}, nil
		}

		err := waitForState(query, svc.Running, 20*time.Millisecond, 5*time.Millisecond)
		assert.Error(t, err)
	})
}

func TestWindowsManagerStart(t *testing.T) {
	t.Run("already running", func(t *testing.T) {
		fake := &fakeService{startErr: windows.ERROR_SERVICE_ALREADY_RUNNING}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			pollInterval: time.Millisecond,
		}
		outBuf := &bytes.Buffer{}
		streams := cli.IO{Out: outBuf, Err: &bytes.Buffer{}}

		code := m.start(streams)

		assert.Equal(t, cli.ExitOK, code)
		assert.Contains(t, outBuf.String(), "already running")
	})

	t.Run("stopped, start succeeds, reaches running", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Running}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			pollInterval: time.Millisecond,
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.start(streams)

		assert.Equal(t, cli.ExitOK, code)
		assert.Equal(t, 1, fake.startCalls)
	})

	t.Run("not installed", func(t *testing.T) {
		m := &windowsManager{
			open: func(string) (serviceController, error) { return nil, notInstalledErr() },
		}
		errBuf := &bytes.Buffer{}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

		code := m.start(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Contains(t, errBuf.String(), "not installed")
	})

	t.Run("open fails for a non-not-installed reason", func(t *testing.T) {
		m := &windowsManager{
			open: func(string) (serviceController, error) {
				return nil, fmt.Errorf("access denied: %w", windows.ERROR_ACCESS_DENIED)
			},
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.start(streams)

		assert.Equal(t, cli.ExitFailure, code)
	})

	t.Run("still starting when the wait times out is reported as failure", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.StartPending}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			startTimeout: 20 * time.Millisecond,
			pollInterval: 2 * time.Millisecond,
		}
		errBuf := &bytes.Buffer{}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

		code := m.start(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Contains(t, errBuf.String(), "may still be starting")
	})

	t.Run("reaches running between the wait timeout and the re-query is reported as success", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.StartPending}, {State: svc.Running}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			startTimeout: 0,
			pollInterval: time.Millisecond,
		}
		outBuf := &bytes.Buffer{}
		streams := cli.IO{Out: outBuf, Err: &bytes.Buffer{}}

		code := m.start(streams)

		assert.Equal(t, cli.ExitOK, code)
		assert.Contains(t, outBuf.String(), "started splunk-otel-collector")
		assert.NotContains(t, outBuf.String(), "may still be starting")
	})

	t.Run("crashes back to stopped after start is reported as failure", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Stopped}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			startTimeout: 20 * time.Millisecond,
			pollInterval: 2 * time.Millisecond,
		}
		errBuf := &bytes.Buffer{}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

		code := m.start(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Contains(t, errBuf.String(), "did not reach the running state")
	})

	t.Run("query error after a failed wait is reported as failure", func(t *testing.T) {
		fake := &fakeService{
			queryErr: errors.New("rpc failure"),
		}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			startTimeout: 20 * time.Millisecond,
			pollInterval: 2 * time.Millisecond,
		}
		errBuf := &bytes.Buffer{}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

		code := m.start(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Contains(t, errBuf.String(), "did not reach the running state")
	})
}

func TestWindowsManagerStop(t *testing.T) {
	t.Run("stops successfully", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Stopped}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			stopTimeout:  time.Second,
			pollInterval: time.Millisecond,
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.stop(streams)

		assert.Equal(t, cli.ExitOK, code)
		require.Len(t, fake.controlCalls, 1)
		assert.Equal(t, svc.Stop, fake.controlCalls[0])
	})

	t.Run("already stopped", func(t *testing.T) {
		fake := &fakeService{controlErr: windows.ERROR_SERVICE_NOT_ACTIVE}
		m := &windowsManager{open: func(string) (serviceController, error) { return fake, nil }}
		outBuf := &bytes.Buffer{}
		streams := cli.IO{Out: outBuf, Err: &bytes.Buffer{}}

		code := m.stop(streams)

		assert.Equal(t, cli.ExitOK, code)
		assert.Contains(t, outBuf.String(), "already stopped")
	})

	t.Run("never reaches stopped within timeout", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.StopPending}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			stopTimeout:  20 * time.Millisecond,
			pollInterval: 5 * time.Millisecond,
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.stop(streams)

		assert.Equal(t, cli.ExitFailure, code)
	})

	t.Run("not installed", func(t *testing.T) {
		m := &windowsManager{open: func(string) (serviceController, error) { return nil, notInstalledErr() }}
		errBuf := &bytes.Buffer{}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: errBuf}

		code := m.stop(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Contains(t, errBuf.String(), "not installed")
	})
}

func TestWindowsManagerStatus(t *testing.T) {
	t.Run("running", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Running}}}
		m := &windowsManager{openQuery: func(string) (serviceController, error) { return fake, nil }}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		assert.Equal(t, cli.ExitOK, m.status(streams))
	})

	t.Run("stopped", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Stopped}}}
		m := &windowsManager{openQuery: func(string) (serviceController, error) { return fake, nil }}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		assert.Equal(t, cli.ExitNotRunning, m.status(streams))
	})

	t.Run("not installed", func(t *testing.T) {
		m := &windowsManager{openQuery: func(string) (serviceController, error) { return nil, notInstalledErr() }}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		assert.Equal(t, cli.ExitNotRunning, m.status(streams))
	})

	t.Run("uses the query-only opener, not the elevated one", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Running}}}
		m := &windowsManager{
			open:      func(string) (serviceController, error) { return nil, errors.New("open must not be called by status") },
			openQuery: func(string) (serviceController, error) { return fake, nil },
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		assert.Equal(t, cli.ExitOK, m.status(streams))
	})
}

func TestWindowsManagerRestart(t *testing.T) {
	t.Run("stops then starts in order", func(t *testing.T) {
		fake := &fakeService{queryResponses: []svc.Status{{State: svc.Stopped}, {State: svc.Running}}}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			stopTimeout:  time.Second,
			pollInterval: time.Millisecond,
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.restart(streams)

		assert.Equal(t, cli.ExitOK, code)
		require.Len(t, fake.controlCalls, 1)
		assert.Equal(t, 1, fake.startCalls)
	})

	t.Run("aborts when stop fails", func(t *testing.T) {
		fake := &fakeService{
			queryResponses: []svc.Status{{State: svc.StopPending}},
			controlErr:     nil,
		}
		m := &windowsManager{
			open:         func(string) (serviceController, error) { return fake, nil },
			stopTimeout:  10 * time.Millisecond,
			pollInterval: 2 * time.Millisecond,
		}
		streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

		code := m.restart(streams)

		assert.Equal(t, cli.ExitFailure, code)
		assert.Equal(t, 0, fake.startCalls, "start must not run when stop fails")
	})
}

func TestNewWindowsDispatchUnknownVerbReturnsUsageError(t *testing.T) {
	family := NewWindows(30 * time.Second)
	streams := cli.IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	code := family.Dispatch("bogus", nil, streams)
	assert.Equal(t, cli.ExitUsage, code)
}
