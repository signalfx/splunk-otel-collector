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

package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/internal/opampsupervisor/launcher"
)

const (
	launcherTestModeEnv      = "SPLUNK_OTEL_LAUNCHER_TEST_MODE"
	launcherTestReadyFileEnv = "SPLUNK_OTEL_LAUNCHER_TEST_READY_FILE"
	ignoreInterruptsTestMode = "ignore-interrupts"
)

func TestMain(m *testing.M) {
	if os.Getenv(launcherTestModeEnv) == ignoreInterruptsTestMode {
		// Re-run this test binary as the child process so the test can verify the
		// force shutdown. This will swallow shutdown signals so the process can only
		// be terminated forcibly. Registers a handler rather than calling signal.Ignore
		// since on Windows an unhandled console control event terminates the process
		// with STATUS_CONTROL_C_EXIT instead of being ignored.
		interrupts := make(chan os.Signal, 8)
		signal.Notify(interrupts, os.Interrupt)
		if err := os.WriteFile(os.Getenv(launcherTestReadyFileEnv), nil, 0o600); err != nil {
			os.Exit(2)
		}
		for {
			<-interrupts
		}
	}
	os.Exit(m.Run())
}

func TestWaitForChildWaitsForOutputForwarding(t *testing.T) {
	outputDone := make(chan error)
	waitStarted := make(chan struct{})
	waitReleased := make(chan struct{})

	done := waitForChild(outputDone, func() error {
		close(waitStarted)
		<-waitReleased
		return nil
	})

	select {
	case <-waitStarted:
		t.Fatal("wait called before output forwarding completed")
	default:
	}

	outputDone <- nil

	<-waitStarted
	close(waitReleased)
	result := <-done

	require.NoError(t, result.outputErr)
	require.NoError(t, result.waitErr)
}

func TestWaitForChildReturnsOutputAndWaitErrors(t *testing.T) {
	outputErr := errors.New("read failed")
	waitErr := errors.New("wait failed")
	outputDone := make(chan error, 1)
	outputDone <- outputErr

	result := <-waitForChild(outputDone, func() error {
		return waitErr
	})

	require.ErrorIs(t, result.outputErr, outputErr)
	require.ErrorIs(t, result.waitErr, waitErr)
}

func TestWaitForChildWithoutOutputForwardingWaitsImmediately(t *testing.T) {
	waitErr := errors.New("wait failed")

	result := <-waitForChild(nil, func() error {
		return waitErr
	})

	require.NoError(t, result.outputErr)
	assert.ErrorIs(t, result.waitErr, waitErr)
}

func TestIsControlCExitCode(t *testing.T) {
	assert.True(t, isControlCExitCode(int(uint32(windows.STATUS_CONTROL_C_EXIT))))
	assert.False(t, isControlCExitCode(1))
}

func TestIsControlCExitWithEmptyExitError(t *testing.T) {
	assert.False(t, isControlCExit(&exec.ExitError{}))
}

func TestWaitForKilledChild_PreservesUnexpectedWaitError(t *testing.T) {
	waitErr := errors.New("wait failed")
	done := make(chan childResult, 1)
	done <- childResult{waitErr: waitErr}

	result := waitForKilledChild(nil, done)

	require.ErrorIs(t, result.waitErr, waitErr)
	require.EqualError(t, result.shutdownWarning, "forcible termination was requested but waiting for the child failed")
}

func TestWaitForKilledChild_ReturnsImmediatelyWhenKillFails(t *testing.T) {
	killErr := errors.New("kill failed")
	resultCh := make(chan childResult, 1)
	go func() {
		resultCh <- waitForKilledChild(killErr, make(chan childResult))
	}()

	select {
	case result := <-resultCh:
		require.ErrorIs(t, result.waitErr, killErr)
		require.NoError(t, result.shutdownWarning)
	case <-time.After(time.Second):
		t.Fatal("killAndWait blocked after the force-kill failed")
	}
}

func TestWaitForKilledChild_PreservesAlreadyTerminatedWaitError(t *testing.T) {
	tests := map[string]error{
		"process done":          os.ErrProcessDone,
		"process handle closed": syscall.EINVAL,
		"process terminated": &os.SyscallError{
			Syscall: "TerminateProcess",
			Err:     windows.ERROR_ACCESS_DENIED,
		},
	}

	for name, killErr := range tests {
		t.Run(name, func(t *testing.T) {
			waitErr := errors.New("wait failed")
			done := make(chan childResult, 1)
			done <- childResult{waitErr: waitErr}

			result := waitForKilledChild(killErr, done)

			require.ErrorIs(t, result.waitErr, waitErr)
			require.EqualError(t, result.shutdownWarning, "child exited before forcible termination was needed")
		})
	}
}

func TestWaitForKilledChild_PreservesDuplicateHandleAccessDenied(t *testing.T) {
	killErr := &os.SyscallError{
		Syscall: "DuplicateHandle",
		Err:     windows.ERROR_ACCESS_DENIED,
	}

	result := waitForKilledChild(killErr, make(chan childResult))

	require.ErrorIs(t, result.waitErr, windows.ERROR_ACCESS_DENIED)
	require.NoError(t, result.shutdownWarning)
}

func TestWaitForKilledChild_SuppressesExpectedForcedExit(t *testing.T) {
	done := make(chan childResult, 1)
	done <- childResult{waitErr: &exec.ExitError{}}

	result := waitForKilledChild(nil, done)

	require.NoError(t, result.waitErr)
	require.NoError(t, result.shutdownWarning)
}

func TestShutdown_PreservesAlreadyCompletedWaitError(t *testing.T) {
	waitErr := errors.New("wait failed")
	done := make(chan childResult, 1)
	done <- childResult{waitErr: waitErr}

	child := &childProcess{
		cmd:  &exec.Cmd{Process: &os.Process{Pid: 1234}},
		done: done,
	}
	result := child.shutdown(time.Hour)

	require.ErrorIs(t, result.waitErr, waitErr)
	require.NoError(t, result.shutdownWarning)
}

func TestShutdown_KillsImmediatelyWhenShutdownSignalFails(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "ready")
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		launcherTestModeEnv+"="+ignoreInterruptsTestMode,
		launcherTestReadyFileEnv+"="+readyFile,
	)
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
	})

	child := &childProcess{
		cmd:  cmd,
		done: waitForChild(nil, cmd.Wait),
	}
	require.Eventually(t, func() bool {
		_, err := os.Stat(readyFile)
		return err == nil
	}, 10*time.Second, 10*time.Millisecond, "timed out waiting for child process to be ready")

	result := waitForShutdown(t, child, gracefulShutdownTimeout)

	require.Error(t, result.shutdownWarning)
	require.Contains(t, result.shutdownWarning.Error(), "failed to send graceful shutdown signal")
	require.NoError(t, result.outputErr)
	require.NoError(t, result.waitErr)

	require.NotNil(t, cmd.ProcessState)
	require.True(t, cmd.ProcessState.Exited())
}

func TestShutdown_KillsChildThatIgnoresShutdownSignal(t *testing.T) {
	readyFile := filepath.Join(t.TempDir(), "ready")
	child, err := startChild(launcher.Command{
		Path: os.Args[0],
		Env: append(os.Environ(),
			launcherTestModeEnv+"="+ignoreInterruptsTestMode,
			launcherTestReadyFileEnv+"="+readyFile,
		),
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = child.cmd.Process.Kill()
	})
	require.Eventually(t, func() bool {
		_, err := os.Stat(readyFile)
		return err == nil
	}, 10*time.Second, 10*time.Millisecond, "timed out waiting for child process to be ready")

	result := waitForShutdown(t, child, time.Second) // setting shorter grace period to not slow down tests too much

	require.Error(t, result.shutdownWarning)
	require.Contains(t, result.shutdownWarning.Error(), "did not exit within 1s and was forcibly terminated")
	require.NoError(t, result.outputErr)
	require.NoError(t, result.waitErr)

	require.NotNil(t, child.cmd.ProcessState)
	require.True(t, child.cmd.ProcessState.Exited())
}

func waitForShutdown(t *testing.T, child *childProcess, gracePeriod time.Duration) childResult {
	t.Helper()

	resultCh := make(chan childResult, 1)
	go func() {
		resultCh <- child.shutdown(gracePeriod)
	}()

	select {
	case result := <-resultCh:
		return result
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not terminate the child process")
		return childResult{}
	}
}
