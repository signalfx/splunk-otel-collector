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

	"github.com/signalfx/splunk-otel-collector/internal/supervisor/launcher"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
	outputDrainTimeout      = 2 * time.Second
	startupFailureWindow    = 500 * time.Millisecond
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
	case err := <-child.processDone:
		_ = child.drainOutput(outputDrainTimeout)
		return err
	case <-interrupt:
		err := child.shutdown(gracefulShutdownTimeout)
		_ = child.drainOutput(outputDrainTimeout)
		return err
	}
}

// serviceHandler keeps the installed Windows service as splunk-otel-collector
// while the launcher manages exactly one child process: otelcol in direct mode
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
	if err, exited := child.exitedWithin(startupFailureWindow); exited {
		if outputErr := child.drainOutput(outputDrainTimeout); outputErr != nil {
			_ = elog.Error(3, fmt.Sprintf("errors occurred while draining child output: %v", outputErr))
		}
		if err != nil {
			_ = elog.Error(3, fmt.Sprintf("child process exited during startup with an error: %v", err))
		} else {
			_ = elog.Error(3, "child process exited during startup")
		}
		return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
	}

	status <- svc.Status{State: svc.Running, Accepts: accepts}

	for {
		select {
		case err := <-child.processDone:
			if outputErr := child.drainOutput(outputDrainTimeout); outputErr != nil {
				_ = elog.Error(3, fmt.Sprintf("errors occurred while draining child output: %v", outputErr))
			}
			if err != nil {
				_ = elog.Error(3, fmt.Sprintf("child process exited with an error: %v", err))
				return false, uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
			}
			return false, 0
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				status <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending, WaitHint: uint32(gracefulShutdownTimeout / time.Millisecond)}
				shutdownErr := child.shutdown(gracefulShutdownTimeout)
				if outputErr := child.drainOutput(outputDrainTimeout); outputErr != nil {
					_ = elog.Error(3, fmt.Sprintf("errors occurred while draining child output: %v", outputErr))
				}
				if shutdownErr != nil {
					_ = elog.Error(3, fmt.Sprintf("errors occurred while shutting down the service: %v", shutdownErr))
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

// childProcess wraps the immediate child selected by the launcher. The child is
// either otelcol.exe in direct mode or opampsupervisor.exe in supervisor mode.
type childProcess struct {
	cmd           *exec.Cmd
	processDone   <-chan error
	outputDone    <-chan error
	outputClosers []io.Closer
}

// windowsEventLog is the subset of the Windows Event Log API the launcher uses
// to bridge service-mode child stdout and stderr.
type windowsEventLog interface {
	Info(eid uint32, msg string) error
	Error(eid uint32, msg string) error
}

// eventLogSink maps child stdout/stderr lines to Windows Event Log severities.
type eventLogSink struct {
	elog windowsEventLog
}

// Info writes a child stdout line as an informational Event Log record.
func (s eventLogSink) Info(msg string) {
	// Event Log write failures are intentionally non-fatal. The launcher must
	// keep draining the child stdout pipe so logging trouble does not block the
	// collector or supervisor process.
	_ = s.elog.Info(1, msg)
}

// Error writes a child stderr line as an error Event Log record.
func (s eventLogSink) Error(msg string) {
	// Event Log write failures are intentionally non-fatal. The launcher must
	// keep draining the child stderr pipe so logging trouble does not block the
	// collector or supervisor process.
	_ = s.elog.Error(3, msg)
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

	var stdoutReader *os.File
	var stderrReader *os.File
	var stdoutWriter *os.File
	var stderrWriter *os.File
	var outputDone <-chan error
	var outputClosers []io.Closer
	if elog == nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		var err error
		stdoutReader, stdoutWriter, err = os.Pipe()
		if err != nil {
			return nil, err
		}
		stderrReader, stderrWriter, err = os.Pipe()
		if err != nil {
			_ = stdoutReader.Close()
			_ = stdoutWriter.Close()
			return nil, err
		}
		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	}

	if err := cmd.Start(); err != nil {
		closeFiles(stdoutReader, stdoutWriter, stderrReader, stderrWriter)
		return nil, err
	}
	if elog != nil {
		_ = stdoutWriter.Close()
		_ = stderrWriter.Close()
		outputDone = forwardChildOutput(stdoutReader, stderrReader, eventLogSink{elog: elog})
		outputClosers = []io.Closer{stdoutReader, stderrReader}
	}

	processDone := make(chan error, 1)
	go func() {
		processDone <- cmd.Wait()
	}()

	return &childProcess{
		cmd:           cmd,
		processDone:   processDone,
		outputDone:    outputDone,
		outputClosers: outputClosers,
	}, nil
}

// closeFiles closes each non-nil file and ignores close errors during cleanup.
func closeFiles(files ...*os.File) {
	for _, file := range files {
		if file != nil {
			_ = file.Close()
		}
	}
}

// exitedWithin reports whether the child process exited inside the startup
// window. The service uses this to catch immediate startup failures before it
// reports Running to the Service Control Manager.
func (p *childProcess) exitedWithin(timeout time.Duration) (error, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-p.processDone:
		return err, true
	case <-timer.C:
		return nil, false
	}
}

// drainOutput waits briefly for service-mode stdout/stderr forwarding to finish
// after the child exits. If another process inherited the pipe and keeps it
// open, the launcher closes its readers and continues service shutdown.
func (p *childProcess) drainOutput(timeout time.Duration) error {
	if p == nil || p.outputDone == nil {
		return nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-p.outputDone:
		return err
	case <-timer.C:
		for _, closer := range p.outputClosers {
			_ = closer.Close()
		}
		select {
		case err := <-p.outputDone:
			return errors.Join(fmt.Errorf("timed out draining child output after %s", timeout), err)
		case <-time.After(100 * time.Millisecond):
			return fmt.Errorf("timed out draining child output after %s", timeout)
		}
	}
}

// shutdown asks the child process group to exit gracefully, then kills only the
// immediate child if it does not exit in time.
func (p *childProcess) shutdown(timeout time.Duration) error {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	select {
	case err := <-p.processDone:
		return err
	default:
	}

	if err := sendShutdownSignal(p.cmd.Process); err != nil {
		select {
		case waitErr := <-p.processDone:
			return errors.Join(fmt.Errorf("send graceful shutdown signal: %w", err), waitErr)
		default:
		}
		return errors.Join(fmt.Errorf("send graceful shutdown signal: %w", err), p.killAndWait())
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-p.processDone:
		return err
	case <-timer.C:
		return errors.Join(fmt.Errorf("child process did not exit within %s", timeout), p.killAndWait())
	}
}

// killAndWait is the hard-stop fallback. It deliberately kills only the
// launcher-owned immediate child; in supervisor mode, opampsupervisor remains
// responsible for shutting down its collector child during the grace period.
func (p *childProcess) killAndWait() error {
	err := p.cmd.Process.Kill()
	waitErr := <-p.processDone
	return errors.Join(err, waitErr)
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
