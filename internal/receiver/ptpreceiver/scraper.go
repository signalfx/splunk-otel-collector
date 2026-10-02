// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

const (
	pmcArgUnixDomainSocket = "-u"
	pmcArgBoundaryHops     = "-b"
	pmcArgClientSocket     = "-i"
	pmcArgServerSocket     = "-s"
	pmcArgDomainNumber     = "-d"
	pmcBoundaryHops        = "0"

	pmcGetTimeStatus     = "GET TIME_STATUS_NP"
	pmcGetCurrentDataSet = "GET CURRENT_DATA_SET"
	pmcGetPortDataSet    = "GET PORT_DATA_SET"
	pmcGetClockDesc      = "GET CLOCK_DESCRIPTION"

	pmcResponseMarker = " RESPONSE MANAGEMENT "

	pmcTimeStatusDataset     = "TIME_STATUS_NP"
	pmcCurrentDataSet        = "CURRENT_DATA_SET"
	pmcPortDataSet           = "PORT_DATA_SET"
	pmcClockDescription      = "CLOCK_DESCRIPTION"
	pmcMasterOffsetField     = "master_offset"
	pmcGrandmasterPresentKey = "gmPresent"
	pmcGrandmasterIdentity   = "gmIdentity"
	pmcMeanPathDelayField    = "meanPathDelay"
	pmcClockTypeField        = "clockType"
	pmcPortIdentityField     = "portIdentity"
	pmcPortStateField        = "portState"
)

const (
	ptpPortStateInitializing = "INITIALIZING"
	ptpPortStateFaulty       = "FAULTY"
	ptpPortStateDisabled     = "DISABLED"
	ptpPortStateListening    = "LISTENING"
	ptpPortStatePreMaster    = "PRE_MASTER"
	ptpPortStateMaster       = "MASTER"
	ptpPortStatePassive      = "PASSIVE"
	ptpPortStateUncalibrated = "UNCALIBRATED"
	ptpPortStateSlave        = "SLAVE"
)

const (
	ptpClockTypeNameOrdinary       = "OC"
	ptpClockTypeNameBoundary       = "BC"
	ptpClockTypeNameP2PTransparent = "P2P_TC"
	ptpClockTypeNameE2ETransparent = "E2E_TC"
	ptpUnknown                     = "UNKNOWN"
	ptpClockStateUnsynchronized    = "UNSYNCHRONIZED"
)

const (
	clockIdentityGroupCount           = 3
	clockIdentityFirstGroupHexLength  = 6
	clockIdentitySecondGroupHexLength = 4
	clockIdentityThirdGroupHexLength  = 6

	// PTP portNumber is an unsigned 16-bit value.
	ptpPortNumberBitSize = 16

	// These are the PTP ClockType values returned by LinuxPTP's
	// CLOCK_DESCRIPTION management response.
	ptpClockTypeOrdinary       uint64 = 0x8000
	ptpClockTypeBoundary       uint64 = 0x4000
	ptpClockTypeP2PTransparent uint64 = 0x2000
	ptpClockTypeE2ETransparent uint64 = 0x1000

	nanosecondsPerSecond = 1e9
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

type pmcDataset struct {
	fields map[string]string
	name   string
}

func newScraper(config *Config, settings receiver.Settings) *ptpScraper {
	return &ptpScraper{config: config, mb: metadata.NewMetricsBuilder(config.MetricsBuilderConfig, settings)}
}

func (s *ptpScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	clientSocketDir, err := os.MkdirTemp(s.config.PMC.ClientSocketDirectory, "ptp-pmc-")
	if err != nil {
		return metrics, fmt.Errorf("create temporary pmc client socket directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(clientSocketDir)
	}()
	clientSocketPath := filepath.Join(clientSocketDir, "pmc.sock")
	run := s.runPMC
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		}
	}
	output, err := run(ctx, s.config.PMC.Path,
		pmcArgUnixDomainSocket, pmcArgBoundaryHops, pmcBoundaryHops,
		pmcArgClientSocket, clientSocketPath,
		pmcArgServerSocket, s.config.SocketPath,
		pmcArgDomainNumber, strconv.Itoa(s.config.DomainNumber),
		pmcGetTimeStatus, pmcGetCurrentDataSet, pmcGetPortDataSet, pmcGetClockDesc)
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
			// LinuxPTP reports offset in nanoseconds; this metric uses seconds.
			s.mb.RecordPtpOffsetDataPoint(now, float64(status.offset)/nanosecondsPerSecond)
			s.mb.RecordPtpGrandmasterInfoDataPoint(now, 1, status.gmIdentity)
		}
	}
	// Offset and path delay are only meaningful relative to a remote GM. Do not
	// emit either measurement when LinuxPTP reports the local clock as GM.
	if status.hasPathDelay && status.gmPresent {
		// pmc formats meanPathDelay in nanoseconds; this metric uses seconds.
		s.mb.RecordPtpPathDelayDataPoint(now, status.pathDelay/nanosecondsPerSecond)
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

func parsePMCStatus(output []byte) (ptpStatus, error) {
	var status ptpStatus
	datasets, err := parsePMCDatasets(output)
	if err != nil {
		return status, err
	}
	if len(datasets) == 0 {
		return status, errors.New("pmc response missing supported management datasets")
	}

	if dataset, ok := firstPMCDataset(datasets, pmcTimeStatusDataset); ok {
		if err := parseTimeStatus(dataset, &status); err != nil {
			return status, err
		}
	}
	if dataset, ok := firstPMCDataset(datasets, pmcCurrentDataSet); ok {
		if err := parsePathDelay(dataset, &status); err != nil {
			return status, err
		}
	}
	if dataset, ok := firstPMCDataset(datasets, pmcClockDescription); ok {
		if err := parseClockType(dataset, &status); err != nil {
			return status, err
		}
	}
	if err := parsePortDataSets(datasets[pmcPortDataSet], &status); err != nil {
		return status, err
	}
	if !status.hasTimeStatus && len(status.ports) == 0 && status.clockType == "" {
		return status, errors.New("pmc response missing TIME_STATUS_NP, PORT_DATA_SET, and CLOCK_DESCRIPTION")
	}
	return status, nil
}

func parsePMCDatasets(output []byte) (map[string][]pmcDataset, error) {
	datasets := make(map[string][]pmcDataset)
	var current *pmcDataset
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if name, ok := pmcResponseName(line); ok {
			dataset := pmcDataset{fields: make(map[string]string), name: name}
			datasets[name] = append(datasets[name], dataset)
			current = &dataset
			continue
		}
		fields := strings.Fields(line)
		if current != nil && len(fields) == 2 {
			current.fields[fields[0]] = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read pmc response: %w", err)
	}
	return datasets, nil
}

func pmcResponseName(line string) (string, bool) {
	// pmc response headers contain "RESPONSE MANAGEMENT <dataset>" after
	// the source port identity and sequence number.
	_, response, ok := strings.Cut(line, pmcResponseMarker)
	if !ok {
		return "", false
	}
	name, _, _ := strings.Cut(response, " ")
	return name, name != ""
}

func firstPMCDataset(datasets map[string][]pmcDataset, name string) (pmcDataset, bool) {
	responses := datasets[name]
	if len(responses) == 0 {
		return pmcDataset{}, false
	}
	return responses[0], true
}

func parseTimeStatus(dataset pmcDataset, status *ptpStatus) error {
	present, err := parseBoolField(dataset, pmcGrandmasterPresentKey)
	if err != nil {
		return err
	}
	status.gmPresent = present
	status.hasTimeStatus = true
	if !status.gmPresent {
		return nil
	}
	status.offset, err = parseIntField(dataset, pmcMasterOffsetField)
	if err != nil {
		return err
	}
	status.gmIdentity, err = getRequiredField(dataset, pmcGrandmasterIdentity)
	if err != nil {
		return err
	}
	if !validClockIdentity(status.gmIdentity) {
		return fmt.Errorf("invalid TIME_STATUS_NP gmIdentity %q", status.gmIdentity)
	}
	return nil
}

func parsePathDelay(dataset pmcDataset, status *ptpStatus) error {
	if status.hasTimeStatus && !status.gmPresent {
		return nil
	}
	value, err := getRequiredField(dataset, pmcMeanPathDelayField)
	if err != nil {
		return err
	}
	status.pathDelay, err = strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(status.pathDelay) || math.IsInf(status.pathDelay, 0) {
		return fmt.Errorf("invalid meanPathDelay %q", value)
	}
	status.hasPathDelay = true
	return nil
}

func parseClockType(dataset pmcDataset, status *ptpStatus) error {
	value, err := getRequiredField(dataset, pmcClockTypeField)
	if err != nil {
		return err
	}
	clockType, err := strconv.ParseUint(value, 0, 16)
	if err != nil {
		return fmt.Errorf("invalid clockType %q: %w", value, err)
	}
	status.clockType = clockTypeName(clockType)
	return nil
}

func parsePortDataSets(datasets []pmcDataset, status *ptpStatus) error {
	seenPorts := make(map[string]bool)
	for _, dataset := range datasets {
		identity, err := getRequiredField(dataset, pmcPortIdentityField)
		if err != nil {
			return err
		}
		state, err := getRequiredField(dataset, pmcPortStateField)
		if err != nil {
			return err
		}
		if !validPortIdentity(identity) {
			return fmt.Errorf("invalid PORT_DATA_SET portIdentity %q", identity)
		}
		if !validPortState(state) {
			return fmt.Errorf("invalid PORT_DATA_SET portState %q", state)
		}
		if !seenPorts[identity] {
			status.ports = append(status.ports, ptpPort{identity: identity, state: state})
			seenPorts[identity] = true
		}
	}
	return nil
}

func getRequiredField(dataset pmcDataset, name string) (string, error) {
	value, ok := dataset.fields[name]
	if !ok {
		return "", fmt.Errorf("pmc response missing %s %s", dataset.name, name)
	}
	return value, nil
}

func parseIntField(dataset pmcDataset, name string) (int64, error) {
	value, err := getRequiredField(dataset, name)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	return parsed, nil
}

func parseBoolField(dataset pmcDataset, name string) (bool, error) {
	value, err := getRequiredField(dataset, name)
	if err != nil {
		return false, err
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	return parsed, nil
}

// validClockIdentity checks LinuxPTP's format for an eight-byte ClockIdentity.
// LinuxPTP prints it as three groups of 3, 2, and 3 bytes, for example
// 001122.fffe.334455. Each byte is two hexadecimal characters, so the groups
// contain 3*2=6, 2*2=4, and 3*2=6 hex characters.
func validClockIdentity(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != clockIdentityGroupCount {
		return false
	}
	groupHexLengths := [...]int{
		clockIdentityFirstGroupHexLength,
		clockIdentitySecondGroupHexLength,
		clockIdentityThirdGroupHexLength,
	}
	for i, part := range parts {
		if len(part) != groupHexLengths[i] {
			return false
		}
		if _, err := hex.DecodeString(part); err != nil {
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
	_, err := strconv.ParseUint(portNumber, 10, ptpPortNumberBitSize)
	return err == nil
}

func validPortState(state string) bool {
	switch state {
	case ptpPortStateInitializing, ptpPortStateFaulty, ptpPortStateDisabled, ptpPortStateListening,
		ptpPortStatePreMaster, ptpPortStateMaster, ptpPortStatePassive, ptpPortStateUncalibrated,
		ptpPortStateSlave:
		return true
	default:
		return false
	}
}

func clockTypeName(clockType uint64) string {
	switch clockType {
	case ptpClockTypeOrdinary:
		return ptpClockTypeNameOrdinary
	case ptpClockTypeBoundary:
		return ptpClockTypeNameBoundary
	case ptpClockTypeP2PTransparent:
		return ptpClockTypeNameP2PTransparent
	case ptpClockTypeE2ETransparent:
		return ptpClockTypeNameE2ETransparent
	default:
		return ptpUnknown
	}
}

func clockState(ports []ptpPort) string {
	if len(ports) == 0 {
		return ptpUnknown
	}
	for _, state := range []string{ptpPortStateSlave, ptpPortStateMaster, ptpPortStateUncalibrated, ptpPortStateFaulty} {
		for _, port := range ports {
			if port.state == state {
				return state
			}
		}
	}
	return ptpClockStateUnsynchronized
}
