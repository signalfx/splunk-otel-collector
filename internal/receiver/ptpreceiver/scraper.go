// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

type ptpScraper struct {
	config *Config
	mb     *metadata.MetricsBuilder
	runPMC func(context.Context, string, ...string) ([]byte, error)
}

type ptpPort struct {
	identity string
	state    string
}

type ptpStatus struct {
	gmIdentity    string
	clockType     string
	ports         []ptpPort
	offset        int64
	pathDelay     float64
	gmPresent     bool
	hasTimeStatus bool
	hasPathDelay  bool
}

func newScraper(config *Config, settings receiver.Settings) *ptpScraper {
	return &ptpScraper{config: config, mb: metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings)}
}

func (s *ptpScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	run := s.runPMC
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		}
	}
	output, err := run(ctx, s.config.PMCPath, "-u", "-b", "0", "-s", s.config.SocketPath,
		"-d", strconv.Itoa(s.config.DomainNumber), "GET TIME_STATUS_NP", "GET CURRENT_DATA_SET",
		"GET PORT_DATA_SET", "GET CLOCK_DESCRIPTION")
	if err != nil {
		return metrics, fmt.Errorf("query ptp4l with pmc: %w: %s", err, strings.TrimSpace(string(output)))
	}
	status, err := parsePMCStatus(output)
	if err != nil {
		return metrics, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	if status.hasTimeStatus {
		var grandmasterPresent int64
		if status.gmPresent {
			grandmasterPresent = 1
		}
		s.mb.RecordPtpGrandmasterPresentDataPoint(now, grandmasterPresent)
		if status.gmPresent {
			s.mb.RecordPtpOffsetDataPoint(now, status.offset)
			s.mb.RecordPtpGrandmasterInfoDataPoint(now, 1, status.gmIdentity)
		}
	}
	if status.hasPathDelay && status.gmPresent {
		s.mb.RecordPtpPathDelayDataPoint(now, status.pathDelay)
	}
	if len(status.ports) > 0 || status.clockType != "" {
		s.mb.RecordPtpClockStateDataPoint(now, 1, clockState(status.ports))
	}
	for _, port := range status.ports {
		s.mb.RecordPtpPortStateDataPoint(now, 1, port.identity, port.state)
	}
	rb := s.mb.NewResourceBuilder()
	rb.SetPtpSocketPath(s.config.SocketPath)
	if status.clockType != "" {
		rb.SetPtpClockType(status.clockType)
	}
	return s.mb.Emit(metadata.WithResource(rb.Emit())), nil
}

// parsePMCStatus reads each response separately because pmc may return several
// PORT_DATA_SET and CLOCK_DESCRIPTION responses for a multi-port clock.
func parsePMCStatus(output []byte) (ptpStatus, error) {
	var status ptpStatus
	datasets := make(map[string][]map[string]string)
	var current map[string]string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for i := 0; i+2 < len(fields); i++ {
			if fields[i] == "RESPONSE" && fields[i+1] == "MANAGEMENT" {
				name := fields[i+2]
				current = make(map[string]string)
				datasets[name] = append(datasets[name], current)
				break
			}
		}
		if current != nil && len(fields) == 2 {
			current[fields[0]] = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return status, fmt.Errorf("read pmc response: %w", err)
	}
	if len(datasets) == 0 {
		return status, errors.New("pmc response missing supported management datasets")
	}

	if responses := datasets["TIME_STATUS_NP"]; len(responses) > 0 {
		fields := responses[0]
		var err error
		status.offset, err = parseIntField(fields, "master_offset", "TIME_STATUS_NP")
		if err != nil {
			return status, err
		}
		present, ok := fields["gmPresent"]
		if !ok {
			return status, errors.New("pmc response missing TIME_STATUS_NP gmPresent")
		}
		status.gmPresent, err = strconv.ParseBool(present)
		if err != nil {
			return status, fmt.Errorf("invalid gmPresent %q: %w", present, err)
		}
		status.hasTimeStatus = true
		if status.gmPresent {
			status.gmIdentity = fields["gmIdentity"]
			if !validClockIdentity(status.gmIdentity) {
				return status, fmt.Errorf("invalid TIME_STATUS_NP gmIdentity %q", status.gmIdentity)
			}
		}
	}

	if responses := datasets["CURRENT_DATA_SET"]; len(responses) > 0 {
		value, ok := responses[0]["meanPathDelay"]
		if !ok {
			return status, errors.New("pmc response missing CURRENT_DATA_SET meanPathDelay")
		}
		var err error
		status.pathDelay, err = strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(status.pathDelay) || math.IsInf(status.pathDelay, 0) {
			return status, fmt.Errorf("invalid meanPathDelay %q", value)
		}
		status.hasPathDelay = true
	}

	if responses := datasets["CLOCK_DESCRIPTION"]; len(responses) > 0 {
		value, ok := responses[0]["clockType"]
		if !ok {
			return status, errors.New("pmc response missing CLOCK_DESCRIPTION clockType")
		}
		clockType, err := strconv.ParseUint(value, 0, 16)
		if err != nil {
			return status, fmt.Errorf("invalid clockType %q: %w", value, err)
		}
		status.clockType = clockTypeName(clockType)
	}

	seenPorts := make(map[string]bool)
	for _, fields := range datasets["PORT_DATA_SET"] {
		identity := fields["portIdentity"]
		state := fields["portState"]
		if !validPortIdentity(identity) || !validPortState(state) {
			return status, fmt.Errorf("invalid PORT_DATA_SET portIdentity %q or portState %q", identity, state)
		}
		if !seenPorts[identity] {
			status.ports = append(status.ports, ptpPort{identity: identity, state: state})
			seenPorts[identity] = true
		}
	}
	if !status.hasTimeStatus && len(status.ports) == 0 && status.clockType == "" {
		return status, errors.New("pmc response missing TIME_STATUS_NP, PORT_DATA_SET, and CLOCK_DESCRIPTION")
	}
	return status, nil
}

func parseIntField(fields map[string]string, key, dataset string) (int64, error) {
	value, ok := fields[key]
	if !ok {
		return 0, fmt.Errorf("pmc response missing %s %s", dataset, key)
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	return parsed, nil
}

func validClockIdentity(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 3 || len(parts[0]) != 6 || len(parts[1]) != 4 || len(parts[2]) != 6 {
		return false
	}
	for _, part := range parts {
		if _, err := strconv.ParseUint(part, 16, len(part)*4); err != nil {
			return false
		}
	}
	return true
}

func validPortIdentity(value string) bool {
	clockID, portNumber, ok := strings.Cut(value, "-")
	if !ok || !validClockIdentity(clockID) {
		return false
	}
	_, err := strconv.ParseUint(portNumber, 10, 16)
	return err == nil
}

func validPortState(state string) bool {
	switch state {
	case "INITIALIZING", "FAULTY", "DISABLED", "LISTENING", "PRE_MASTER", "MASTER", "PASSIVE", "UNCALIBRATED", "SLAVE":
		return true
	default:
		return false
	}
}

func clockTypeName(clockType uint64) string {
	switch clockType {
	case 0x8000:
		return "ordinary"
	case 0x4000:
		return "boundary"
	case 0x2000:
		return "p2p_transparent"
	case 0x1000:
		return "e2e_transparent"
	default:
		return "unknown"
	}
}

func clockState(ports []ptpPort) string {
	if len(ports) == 0 {
		return "UNKNOWN"
	}
	for _, state := range []string{"SLAVE", "MASTER", "UNCALIBRATED", "FAULTY"} {
		for _, port := range ports {
			if port.state == state {
				return state
			}
		}
	}
	return "UNSYNCHRONIZED"
}
