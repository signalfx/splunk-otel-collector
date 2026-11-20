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

package supportbundle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "splunk-otel-collector"

func defaultConfigDir() string {
	base := os.Getenv("PROGRAMDATA")
	if base == "" {
		base = `C:\ProgramData`
	}
	return filepath.Join(base, "Splunk", "OpenTelemetry Collector")
}

func defaultTempRoot() string {
	base := os.Getenv("PROGRAMFILES")
	if base == "" {
		base = `C:\Program Files`
	}
	return filepath.Join(base, "Splunk", "OpenTelemetry Collector")
}

func collectServiceStatus(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting status...")
	dest := filepath.Join(tempDir, logsDirName, "splunk-otel-collector.txt")

	content, err := queryServiceStatus(serviceName)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			logger.Warnf("Service %s does not exist: %v", serviceName, err)
			content = fmt.Sprintf("Service %s does not exist.\n", serviceName)
		} else {
			logger.Warnf("failed to query service %s: %v", serviceName, err)
			content = fmt.Sprintf("Failed to query service %s: %v\n", serviceName, err)
		}
	}
	writeOrWarn(logger, dest, content)
}

func collectServiceLogs(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting logs...")
	dest := filepath.Join(tempDir, logsDirName, "splunk-otel-collector.log")

	content, err := readApplicationEventLog(serviceName)
	if err != nil {
		logger.Warnf("Failed to read Application event log for %s: %v", serviceName, err)
		failure := fmt.Sprintf("Failed to read Application event log for %s: %v\n", serviceName, err)
		if content == "" {
			content = failure
		} else {
			content += "\n" + failure
		}
	} else if content == "" {
		logger.Warnf("No matching events found for %s in scanned records.", serviceName)
		content = fmt.Sprintf("No matching events found for %s in scanned records.\n", serviceName)
	}
	writeOrWarn(logger, dest, content)
}

func collectExtras(_ context.Context, logger *logger, tempDir string) {
	collectMSIInfo(logger, tempDir)
	collectServiceEnvironment(logger, tempDir)
}

// queryServiceStatus reports the Service Control Manager status of name,
// equivalent to `Get-Service <name>`, using golang.org/x/sys/windows/svc/mgr
// directly instead of shelling out to powershell.exe.
func queryServiceStatus(name string) (string, error) {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return "", fmt.Errorf("invalid service name %q: %w", name, err)
	}

	m, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return "", fmt.Errorf("connect to service manager: %w", err)
	}
	defer func() { _ = windows.CloseServiceHandle(m) }()

	h, err := windows.OpenService(m, namePtr, windows.SERVICE_QUERY_STATUS|windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return "", fmt.Errorf("open service %q: %w", name, err)
	}
	s := mgr.Service{Name: name, Handle: h}
	defer func() { _ = s.Close() }()

	status, err := s.Query()
	if err != nil {
		return "", fmt.Errorf("query service %q: %w", name, err)
	}

	displayName := name
	if cfg, err := s.Config(); err == nil {
		displayName = cfg.DisplayName
	}

	return formatServiceStatus(serviceStateString(status.State), name, displayName), nil
}

func formatServiceStatus(state, name, displayName string) string {
	var output strings.Builder
	writer := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Status\tName\tDisplayName")
	fmt.Fprintln(writer, "------\t----\t-----------")
	fmt.Fprintf(writer, "%s\t%s\t%s\n", state, name, displayName)
	_ = writer.Flush()
	return output.String()
}

func serviceStateString(s svc.State) string {
	switch s {
	case svc.Stopped:
		return "Stopped"
	case svc.StartPending:
		return "StartPending"
	case svc.StopPending:
		return "StopPending"
	case svc.Running:
		return "Running"
	case svc.ContinuePending:
		return "ContinuePending"
	case svc.PausePending:
		return "PausePending"
	case svc.Paused:
		return "Paused"
	default:
		return "Unknown"
	}
}

var (
	wevtapi       = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtQuery  = wevtapi.NewProc("EvtQuery")
	procEvtNext   = wevtapi.NewProc("EvtNext")
	procEvtRender = wevtapi.NewProc("EvtRender")
	procEvtClose  = wevtapi.NewProc("EvtClose")
)

const (
	evtQueryChannelPath       = 0x0001
	evtQueryReverseDirection  = 0x0200
	evtRenderEventXML         = 1
	eventLogBatchSize         = 16
	maxEventLogRecordsScanned = 5000
	maxEventLogXMLSize        = 4 << 20
)

// readApplicationEventLog reads recent events from the Windows Application
// channel and returns those whose provider matches source.
func readApplicationEventLog(source string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	channel, err := windows.UTF16PtrFromString("Application")
	if err != nil {
		return "", err
	}
	query, err := windows.UTF16PtrFromString(fmt.Sprintf(
		"*[System[Provider[@Name='%s' or @EventSourceName='%s']]]",
		source, source,
	))
	if err != nil {
		return "", err
	}
	result, _, callErr := procEvtQuery.Call(
		0,
		uintptr(unsafe.Pointer(channel)),
		uintptr(unsafe.Pointer(query)),
		evtQueryChannelPath|evtQueryReverseDirection,
	)
	runtime.KeepAlive(channel)
	runtime.KeepAlive(query)
	if result == 0 {
		return "", fmt.Errorf("EvtQuery Application: %w", callErr)
	}
	defer func() { _, _, _ = procEvtClose.Call(result) }()

	var sb strings.Builder
	scanned := 0
	renderer := eventXMLRenderer{}
	for scanned < maxEventLogRecordsScanned {
		batchSize := min(eventLogBatchSize, maxEventLogRecordsScanned-scanned)
		events := make([]windows.Handle, batchSize)
		var returned uint32
		ok, _, nextErr := procEvtNext.Call(
			result,
			uintptr(batchSize),
			uintptr(unsafe.Pointer(&events[0])),
			0,
			0,
			uintptr(unsafe.Pointer(&returned)),
		)
		if ok == 0 {
			closeEventHandles(events[:returned])
			if errors.Is(nextErr, windows.ERROR_NO_MORE_ITEMS) {
				break
			}
			return sb.String(), fmt.Errorf("EvtNext: %w", nextErr)
		}
		if returned == 0 {
			break
		}

		entries, err := readEventLogBatch(events[:returned], &renderer)
		if err != nil {
			return sb.String(), err
		}
		scanned += int(returned)
		writeEventEntries(&sb, entries)
	}

	if scanned == 0 {
		return "", nil
	}
	return sb.String(), nil
}

func readEventLogBatch(events []windows.Handle, renderer *eventXMLRenderer) ([]eventLogEntry, error) {
	defer closeEventHandles(events)

	entries := make([]eventLogEntry, 0, len(events))
	for _, event := range events {
		xml, err := renderer.render(event)
		if err != nil {
			return entries, err
		}
		entry, err := parseEventLogEvent(xml)
		if err != nil {
			return entries, fmt.Errorf("parse rendered event: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func closeEventHandles(events []windows.Handle) {
	for _, event := range events {
		_, _, _ = procEvtClose.Call(uintptr(event))
	}
}

type eventXMLRenderer struct {
	buffer []uint16
}

func (r *eventXMLRenderer) render(event windows.Handle) ([]byte, error) {
	var bufferUsed, propertyCount uint32
	ok, _, callErr := procEvtRender.Call(
		0,
		uintptr(event),
		evtRenderEventXML,
		0,
		0,
		uintptr(unsafe.Pointer(&bufferUsed)),
		uintptr(unsafe.Pointer(&propertyCount)),
	)
	if ok != 0 {
		return nil, fmt.Errorf("EvtRender succeeded without an output buffer")
	}
	if !errors.Is(callErr, windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("EvtRender size: %w", callErr)
	}
	if bufferUsed == 0 || bufferUsed%2 != 0 || bufferUsed > maxEventLogXMLSize {
		return nil, fmt.Errorf("EvtRender returned invalid XML buffer size %d", bufferUsed)
	}

	if cap(r.buffer) < int(bufferUsed/2) {
		r.buffer = make([]uint16, bufferUsed/2)
	} else {
		r.buffer = r.buffer[:bufferUsed/2]
	}
	ok, _, callErr = procEvtRender.Call(
		0,
		uintptr(event),
		evtRenderEventXML,
		uintptr(bufferUsed),
		uintptr(unsafe.Pointer(&r.buffer[0])),
		uintptr(unsafe.Pointer(&bufferUsed)),
		uintptr(unsafe.Pointer(&propertyCount)),
	)
	if ok == 0 {
		return nil, fmt.Errorf("EvtRender: %w", callErr)
	}
	return []byte(windows.UTF16ToString(r.buffer[:bufferUsed/2])), nil
}

func writeEventEntries(sb *strings.Builder, entries []eventLogEntry) {
	for _, e := range entries {
		fmt.Fprintf(sb, "%s  %-13s  EventID=%d  Source=%s\n    %s\n",
			e.TimeGenerated.Format(time.RFC3339), eventLevelString(e.Level), e.EventID, e.Source, strings.Join(e.Strings, " | "))
	}
}

func eventLevelString(level uint16) string {
	switch level {
	case 0:
		return "LOG_ALWAYS"
	case 1:
		return "CRITICAL"
	case 2:
		return "ERROR"
	case 3:
		return "WARNING"
	case 4:
		return "INFO"
	case 5:
		return "VERBOSE"
	default:
		return "UNKNOWN"
	}
}
