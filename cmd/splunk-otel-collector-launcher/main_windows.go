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
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/signalfx/splunk-otel-collector/internal/opampsupervisor/launcher"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
)

var (
	kernel32API = windows.NewLazySystemDLL("kernel32.dll")

	allocConsoleProc = kernel32API.NewProc("AllocConsole")
	freeConsoleProc  = kernel32API.NewProc("FreeConsole")
)

// run starts as a Windows service when invoked by the Service Control Manager
// and falls back to interactive mode for command-line runs.
func run(args, env []string, paths launcher.Paths) error {
	// Allocate a console so it can later deliver CTRL_BREAK_EVENT to a child process group.
	if err := allocConsole(); err != nil && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("alloc console: %w", err)
	}
	defer func() {
		_ = freeConsole()
	}()

	if err := svc.Run("", &serviceHandler{args: args, env: env, paths: paths}); err != nil {
		if errors.Is(err, windows.ERROR_FAILED_SERVICE_CONTROLLER_CONNECT) {
			return runInteractive(args, env, paths)
		}
		return fmt.Errorf("failed to start launcher service handler: %w", err)
	}
	return nil
}

// runInteractive executes the selected child process directly when the launcher
// is not running under the Windows Service Control Manager.
func runInteractive(args, env []string, paths launcher.Paths) error {
	cmdSpec, err := launcher.PrepareCommand(args, env, paths)
	if err != nil {
		return err
	}
	child, err := startChild(cmdSpec, true, nil)
	if err != nil {
		return err
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	select {
	case result := <-child.done:
		return errors.Join(result.outputErr, result.waitErr)
	case <-interrupt:
		result := child.shutdown(gracefulShutdownTimeout)
		return errors.Join(result.outputErr, result.waitErr)
	}
}

// serviceHandler keeps the installed Windows service as splunk-otel-collector
// while the launcher manages the child process: otelcol in direct mode
// or opampsupervisor in supervisor mode.
type serviceHandler struct {
	args  []string
	env   []string
	paths launcher.Paths
}

// Execute implements the Windows service lifecycle for a child-process launcher.
func (h *serviceHandler) Execute(serviceArgs []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const accepts = svc.AcceptStop | svc.AcceptShutdown

	if len(serviceArgs) == 0 {
		return false, uint32(windows.ERROR_INVALID_SERVICENAME)
	}

	elog, err := eventlog.Open(serviceArgs[0])
	if err != nil {
		return false, uint32(windows.ERROR_EVENTLOG_CANT_START)
	}
	defer func() {
		_ = elog.Close()
	}()

	status <- svc.Status{State: svc.StartPending}
	cmdSpec, err := launcher.PrepareCommand(h.args, h.env, h.paths)
	if err != nil {
		_ = elog.Error(3, fmt.Sprintf("failed to prepare launcher command: %v", err))
		return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
	}

	child, err := startChild(cmdSpec, false, elog)
	if err != nil {
		_ = elog.Error(3, fmt.Sprintf("failed to start child process: %v", err))
		return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
	}

	status <- svc.Status{State: svc.Running, Accepts: accepts}

	// The service follows the launcher-owned child process lifecycle while also
	// responding to Service Control Manager requests.
	for {
		select {
		case result := <-child.done:
			if result.outputErr != nil {
				_ = elog.Error(3, fmt.Sprintf("errors occurred while forwarding child output: %v", result.outputErr))
			}
			if result.waitErr != nil {
				_ = elog.Error(3, fmt.Sprintf("child process exited with an error: %v", result.waitErr))
				return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
			}
			return false, 0
		case request, ok := <-requests:
			if !ok {
				return false, 0
			}
			switch request.Cmd {
			case svc.Interrogate:
				status <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending, WaitHint: uint32(gracefulShutdownTimeout / time.Millisecond)}
				result := child.shutdown(gracefulShutdownTimeout)
				if result.outputErr != nil {
					_ = elog.Error(3, fmt.Sprintf("errors occurred while forwarding child output: %v", result.outputErr))
				}
				if result.waitErr != nil {
					_ = elog.Error(3, fmt.Sprintf("errors occurred while shutting down the service: %v", result.waitErr))
					return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
				}
				status <- svc.Status{State: svc.Stopped}
				return false, 0
			default:
				_ = elog.Error(3, fmt.Sprintf("unexpected service control request #%d", request.Cmd))
				return false, uint32(windows.ERROR_INVALID_SERVICE_CONTROL)
			}
		}
	}
}

// windowsEventLog is the subset of the Windows Event Log API the launcher uses
// to bridge service-mode child stdout and stderr.
type windowsEventLog interface {
	Info(eid uint32, msg string) error
	Error(eid uint32, msg string) error
}

// eventLogSink maps child stdout/stderr lines to Windows Event Log records.
// Write failures are non-fatal so logging issues cannot block pipe draining.
type eventLogSink struct {
	elog windowsEventLog
}

func (s eventLogSink) Info(msg string) {
	_ = s.elog.Info(1, msg)
}

func (s eventLogSink) Error(msg string) {
	_ = s.elog.Error(3, msg)
}

// childResult keeps process wait errors separate from child output forwarding
// errors so logging failures do not look like child process failures.
type childResult struct {
	waitErr   error
	outputErr error
}

// childProcess wraps the immediate child selected by the launcher. The child is
// either otelcol.exe in direct mode or opampsupervisor.exe in supervisor mode.
type childProcess struct {
	cmd *exec.Cmd
	// outputClosers unblock StdoutPipe/StderrPipe forwarding during hard-stop cleanup.
	outputClosers []io.ReadCloser
	done          <-chan childResult
}

// startChild launches the selected binary in its own process group. Interactive
// runs inherit console stdio; service runs bridge child stdout/stderr to Event
// Log through the provided event log handle.
func startChild(cmdSpec launcher.Command, interactive bool, elog windowsEventLog) (*childProcess, error) {
	cmd := exec.Command(cmdSpec.Path, cmdSpec.Args...)
	cmd.Env = cmdSpec.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}
	if interactive {
		cmd.Stdin = os.Stdin
	}

	var outputDone <-chan error
	var outputClosers []io.ReadCloser
	if elog == nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		stdoutReader, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		stderrReader, err := cmd.StderrPipe()
		if err != nil {
			_ = stdoutReader.Close()
			return nil, err
		}
		outputClosers = []io.ReadCloser{stdoutReader, stderrReader}
	}

	if err := cmd.Start(); err != nil {
		for _, closer := range outputClosers {
			_ = closer.Close()
		}
		return nil, err
	}
	if elog != nil {
		outputDone = forwardChildOutput(outputClosers[0], outputClosers[1], eventLogSink{elog: elog})
	}

	return &childProcess{
		cmd:           cmd,
		outputClosers: outputClosers,
		done:          waitForChild(outputDone, cmd.Wait),
	}, nil
}

// waitForChild waits for stdout/stderr forwarding before cmd.Wait, as required
// by StdoutPipe and StderrPipe.
func waitForChild(outputDone <-chan error, waitForProcess func() error) <-chan childResult {
	done := make(chan childResult, 1)
	go func() {
		var outputErr error
		if outputDone != nil {
			outputErr = <-outputDone
		}

		waitErr := waitForProcess()
		done <- childResult{outputErr: outputErr, waitErr: waitErr}
	}()
	return done
}

// shutdown asks the child process group to exit gracefully, then kills only the
// immediate child if it does not exit in time.
func (p *childProcess) shutdown(timeout time.Duration) childResult {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return childResult{}
	}

	// Check for an already-exited child before signaling.
	select {
	case result := <-p.done:
		return result
	default:
	}

	if err := sendShutdownSignal(p.cmd.Process); err != nil {
		select {
		case result := <-p.done:
			result.waitErr = errors.Join(fmt.Errorf("failed to send graceful shutdown signal: %w", err), result.waitErr)
			return result
		default:
		}
		result := p.killAndWait()
		result.waitErr = errors.Join(fmt.Errorf("failed to send graceful shutdown signal: %w", err), result.waitErr)
		return result
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case result := <-p.done:
		if isControlCExit(result.waitErr) {
			result.waitErr = nil
		}
		return result
	case <-timer.C:
		result := p.killAndWait()
		result.waitErr = errors.Join(fmt.Errorf("child process did not exit within %s", timeout), result.waitErr)
		return result
	}
}

// Windows reports a process closed by CTRL_BREAK_EVENT as STATUS_CONTROL_C_EXIT;
// only that exit code is normal for launcher-requested graceful shutdown.
func isControlCExit(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return isControlCExitCode(exitErr.ExitCode())
}

func isControlCExitCode(exitCode int) bool {
	return uint32(exitCode) == uint32(windows.STATUS_CONTROL_C_EXIT)
}

// killAndWait is the hard-stop fallback. It deliberately kills only the
// launcher-owned immediate child; in supervisor mode, opampsupervisor remains
// responsible for shutting down its collector child during the grace period.
func (p *childProcess) killAndWait() childResult {
	for _, closer := range p.outputClosers {
		_ = closer.Close()
	}
	err := p.cmd.Process.Kill()
	result := <-p.done
	result.waitErr = errors.Join(err, result.waitErr)
	return result
}

// sendShutdownSignal uses the Windows console API because os.Interrupt is not
// supported for signaling another process on Windows.
func sendShutdownSignal(process *os.Process) error {
	if err := windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(process.Pid)); err != nil {
		return fmt.Errorf("send CTRL_BREAK_EVENT to process group %d: %w", process.Pid, err)
	}
	return nil
}

// allocConsole creates the console a Windows service needs before it can send
// console control events to a child process group.
func allocConsole() error {
	ret, _, err := allocConsoleProc.Call()
	if ret == 0 {
		return err
	}
	return nil
}

// freeConsole releases the service console allocated by allocConsole.
func freeConsole() error {
	ret, _, err := freeConsoleProc.Call()
	if ret == 0 {
		return err
	}
	return nil
}
