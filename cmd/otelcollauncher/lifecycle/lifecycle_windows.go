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
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const windowsServiceName = "splunk-otel-collector"

type serviceController interface {
	Start(args ...string) error
	Control(c svc.Cmd) (svc.Status, error)
	Query() (svc.Status, error)
	Close() error
}

type serviceOpener func(name string) (serviceController, error)

type scmService struct {
	mgr *mgr.Mgr
	svc *mgr.Service
}

func (s *scmService) Start(args ...string) error            { return s.svc.Start(args...) }
func (s *scmService) Control(c svc.Cmd) (svc.Status, error) { return s.svc.Control(c) }
func (s *scmService) Query() (svc.Status, error)            { return s.svc.Query() }

func (s *scmService) Close() error {
	err := s.svc.Close()
	if dErr := s.mgr.Disconnect(); dErr != nil {
		err = errors.Join(err, dErr)
	}
	return err
}

func openSCMService(name string) (serviceController, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the service control manager: %w", err)
	}
	s, err := m.OpenService(name)
	if err != nil {
		_ = m.Disconnect()
		return nil, fmt.Errorf("failed to open service %q: %w", name, err)
	}
	return &scmService{mgr: m, svc: s}, nil
}

type queryOnlyService struct {
	scHandle  windows.Handle
	svcHandle windows.Handle
}

func (s *queryOnlyService) Start(_ ...string) error {
	return errors.New("query-only service handle does not support Start")
}

func (s *queryOnlyService) Control(_ svc.Cmd) (svc.Status, error) {
	return svc.Status{}, errors.New("query-only service handle does not support Control")
}

func (s *queryOnlyService) Query() (svc.Status, error) {
	var t windows.SERVICE_STATUS_PROCESS
	var needed uint32
	err := windows.QueryServiceStatusEx(s.svcHandle, windows.SC_STATUS_PROCESS_INFO, (*byte)(unsafe.Pointer(&t)), uint32(unsafe.Sizeof(t)), &needed)
	if err != nil {
		return svc.Status{}, err
	}
	return svc.Status{
		State:                   svc.State(t.CurrentState),
		Accepts:                 svc.Accepted(t.ControlsAccepted),
		ProcessId:               t.ProcessId,
		Win32ExitCode:           t.Win32ExitCode,
		ServiceSpecificExitCode: t.ServiceSpecificExitCode,
	}, nil
}

func (s *queryOnlyService) Close() error {
	err := windows.CloseServiceHandle(s.svcHandle)
	if cErr := windows.CloseServiceHandle(s.scHandle); cErr != nil {
		err = errors.Join(err, cErr)
	}
	return err
}

func openSCMServiceForQuery(name string) (serviceController, error) {
	scHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the service control manager: %w", err)
	}
	namePointer, err := windows.UTF16PtrFromString(name)
	if err != nil {
		_ = windows.CloseServiceHandle(scHandle)
		return nil, fmt.Errorf("failed to encode service name %q: %w", name, err)
	}
	svcHandle, err := windows.OpenService(scHandle, namePointer, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		_ = windows.CloseServiceHandle(scHandle)
		return nil, fmt.Errorf("failed to open service %q: %w", name, err)
	}
	return &queryOnlyService{scHandle: scHandle, svcHandle: svcHandle}, nil
}

type windowsManager struct {
	open         serviceOpener
	openQuery    serviceOpener
	service      string
	startTimeout time.Duration
	stopTimeout  time.Duration
	pollInterval time.Duration
}

func NewWindows(stopTimeout time.Duration) *Manager {
	m := &windowsManager{
		open:         openSCMService,
		openQuery:    openSCMServiceForQuery,
		service:      windowsServiceName,
		startTimeout: 10 * time.Second,
		stopTimeout:  stopTimeout,
		pollInterval: 500 * time.Millisecond,
	}
	return &Manager{
		dispatch: func(v verb, _ []string, streams cli.IO) int {
			switch v {
			case verbStart:
				return m.start(streams)
			case verbStop:
				return m.stop(streams)
			case verbStatus:
				return m.status(streams)
			case verbRestart:
				return m.restart(streams)
			default:
				fmt.Fprintf(streams.Err, "unknown lifecycle command %q\n", string(v))
				return cli.ExitUsage
			}
		},
	}
}

func (m *windowsManager) start(streams cli.IO) int {
	s, err := m.open(m.service)
	if err != nil {
		if serviceNotInstalled(err) {
			fmt.Fprintln(streams.Err, "splunk-otel-collector service is not installed")
			return cli.ExitFailure
		}
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
			fmt.Fprintln(streams.Out, "splunk-otel-collector is already running")
			return cli.ExitOK
		}
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}

	if waitErr := waitForState(s.Query, svc.Running, m.startTimeout, m.pollInterval); waitErr != nil {
		st, err := s.Query()
		switch {
		case err != nil:
			fmt.Fprintf(streams.Err, "splunk-otel-collector did not reach the running state: %v\n", waitErr)
		case st.State == svc.Running:
			fmt.Fprintln(streams.Out, "started splunk-otel-collector")
			return cli.ExitOK
		case st.State == svc.StartPending:
			fmt.Fprintf(streams.Err, "splunk-otel-collector did not reach the running state within %s; it may still be starting: %v\n", m.startTimeout, waitErr)
		default:
			fmt.Fprintf(streams.Err, "splunk-otel-collector did not reach the running state: %v\n", waitErr)
		}
		return cli.ExitFailure
	}
	fmt.Fprintln(streams.Out, "started splunk-otel-collector")
	return cli.ExitOK
}

func (m *windowsManager) stop(streams cli.IO) int {
	s, err := m.open(m.service)
	if err != nil {
		if serviceNotInstalled(err) {
			fmt.Fprintf(streams.Err, "splunk-otel-collector service is not installed\n")
			return cli.ExitFailure
		}
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	defer s.Close()

	if _, err := s.Control(svc.Stop); err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			fmt.Fprintln(streams.Out, "splunk-otel-collector is already stopped")
			return cli.ExitOK
		}
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}

	if err := waitForState(s.Query, svc.Stopped, m.stopTimeout, m.pollInterval); err != nil {
		fmt.Fprintf(streams.Err, "splunk-otel-collector did not stop within %s: %v\n", m.stopTimeout, err)
		return cli.ExitFailure
	}
	fmt.Fprintln(streams.Out, "stopped splunk-otel-collector")
	return cli.ExitOK
}

func (m *windowsManager) status(streams cli.IO) int {
	s, err := m.openQuery(m.service)
	if err != nil {
		if serviceNotInstalled(err) {
			fmt.Fprintln(streams.Out, "splunk-otel-collector service is not installed")
			return cli.ExitNotRunning
		}
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	defer s.Close()

	st, err := s.Query()
	if err != nil {
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}

	fmt.Fprintf(streams.Out, "splunk-otel-collector is %s\n", stateName(st.State))
	if st.State == svc.Running {
		return cli.ExitOK
	}
	return cli.ExitNotRunning
}

func (m *windowsManager) restart(streams cli.IO) int {
	if code := m.stop(streams); code != cli.ExitOK {
		return code
	}
	return m.start(streams)
}

func serviceNotInstalled(err error) bool {
	return errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST)
}

func waitForState(query func() (svc.Status, error), want svc.State, timeout, poll time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		st, err := query()
		if err != nil {
			return err
		}
		if st.State == want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for state %s", stateName(want))
		}
		time.Sleep(poll)
	}
}

func stateName(s svc.State) string {
	switch s {
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "starting"
	case svc.StopPending:
		return "stopping"
	case svc.Running:
		return "running"
	case svc.ContinuePending:
		return "resuming"
	case svc.PausePending:
		return "pausing"
	case svc.Paused:
		return "paused"
	default:
		return fmt.Sprintf("unknown (%d)", int(s))
	}
}
