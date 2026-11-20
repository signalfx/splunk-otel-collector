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
	pmcArgBoundaryHops     = "-b"
	pmcArgClientSocket     = "-i"
	pmcArgDomainNumber     = "-d"
	pmcArgServerSocket     = "-s"
	pmcArgUnixDomainSocket = "-u"
	pmcBoundaryHops        = "0"

	pmcGetClockDesc      = "GET CLOCK_DESCRIPTION"
	pmcGetCurrentDataSet = "GET CURRENT_DATA_SET"
	pmcGetPortDataSet    = "GET PORT_DATA_SET"
	pmcGetTimeStatus     = "GET TIME_STATUS_NP"

	pmcResponseMarker = " RESPONSE MANAGEMENT "

	pmcClockDescription      = "CLOCK_DESCRIPTION"
	pmcClockTypeField        = "clockType"
	pmcCurrentDataSet        = "CURRENT_DATA_SET"
	pmcGrandmasterIdentity   = "gmIdentity"
	pmcGrandmasterPresentKey = "gmPresent"
	pmcMasterOffsetField     = "master_offset"
	pmcMeanPathDelayField    = "meanPathDelay"
	pmcPortDataSet           = "PORT_DATA_SET"
	pmcPortIdentityField     = "portIdentity"
	pmcPortStateField        = "portState"
	pmcTimeStatusDataset     = "TIME_STATUS_NP"
	pmcResponsePreviewBytes  = 256
)

const (
	ptpPortStateDisabled     = "DISABLED"
	ptpPortStateFaulty       = "FAULTY"
	ptpPortStateGrandMaster  = "GRAND_MASTER"
	ptpPortStateInitializing = "INITIALIZING"
	ptpPortStateListening    = "LISTENING"
	ptpPortStateMaster       = "MASTER"
	ptpPortStatePassive      = "PASSIVE"
	ptpPortStatePreMaster    = "PRE_MASTER"
	ptpPortStateSlave        = "SLAVE"
	ptpPortStateUncalibrated = "UNCALIBRATED"
)

const (
	ptpClockStateUnsynchronized    = "UNSYNCHRONIZED"
	ptpClockTypeNameBoundary       = "BC"
	ptpClockTypeNameE2ETransparent = "E2E_TC"
	ptpClockTypeNameOrdinary       = "OC"
	ptpClockTypeNameP2PTransparent = "P2P_TC"
	ptpUnknown                     = "UNKNOWN"
)

const (
	clockIdentityFirstGroupHexLength  = 6
	clockIdentityGroupCount           = 3
	clockIdentitySecondGroupHexLength = 4
	clockIdentityThirdGroupHexLength  = 6
	nanosecondsPerSecond              = 1e9

	// These are the PTP ClockType values returned by LinuxPTP's
	// CLOCK_DESCRIPTION management response.
	ptpClockTypeBoundary       uint64 = 0x4000
	ptpClockTypeE2ETransparent uint64 = 0x1000
	ptpClockTypeOrdinary       uint64 = 0x8000
	ptpClockTypeP2PTransparent uint64 = 0x2000

	// PTP portNumber is an unsigned 16-bit value.
	ptpPortNumberBitSize = 16
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
	status, err := s.collectStatus(ctx)
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
		s.mb.RecordPtpClockStateDataPoint(now, 1, metadata.MapAttributePtpClockState[clockState(status.ports)])
	}
	for _, port := range status.ports {
		s.mb.RecordPtpPortStateDataPoint(now, 1, port.identity, metadata.MapAttributePtpPortState[port.state])
	}
	rb := s.mb.NewResourceBuilder()
	rb.SetPtpSocketPath(s.config.SocketPath)
	if status.clockType != "" {
		setPtpClockTypeAttribute(rb, status.clockType)
	}
	return s.mb.Emit(metadata.WithResource(rb.Emit())), nil
}

// collectStatus queries the ptp4l management socket and parses its replies.
// The private client socket directory is removed after pmc exits.
func (s *ptpScraper) collectStatus(ctx context.Context) (ptpStatus, error) {
	clientSocketDir, err := os.MkdirTemp(s.config.PMC.ClientSocketDirectory, "ptp-pmc-")
	if err != nil {
		return ptpStatus{}, fmt.Errorf("create temporary pmc client socket directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(clientSocketDir)
	}()
	clientSocketPath := filepath.Join(clientSocketDir, "pmc.sock")
	output, err := s.queryPMC(ctx, clientSocketPath)
	if err != nil {
		return ptpStatus{}, fmt.Errorf("query ptp4l with pmc: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return parsePMCStatus(output)
}

// queryPMC runs pmc with the configured server socket and a writable client socket.
func (s *ptpScraper) queryPMC(ctx context.Context, clientSocketPath string) ([]byte, error) {
	run := s.runPMC
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		}
	}
	return run(ctx, s.config.PMC.Path,
		pmcArgUnixDomainSocket, pmcArgBoundaryHops, pmcBoundaryHops,
		pmcArgClientSocket, clientSocketPath,
		pmcArgServerSocket, s.config.SocketPath,
		pmcArgDomainNumber, strconv.Itoa(s.config.DomainNumber),
		pmcGetTimeStatus, pmcGetCurrentDataSet, pmcGetPortDataSet, pmcGetClockDesc)
}

func setPtpClockTypeAttribute(rb *metadata.ResourceBuilder, clockType string) {
	switch clockType {
	case ptpClockTypeNameOrdinary:
		rb.SetPtpClockTypeOC()
	case ptpClockTypeNameBoundary:
		rb.SetPtpClockTypeBC()
	case ptpClockTypeNameP2PTransparent:
		rb.SetPtpClockTypeP2PTC()
	case ptpClockTypeNameE2ETransparent:
		rb.SetPtpClockTypeE2ETC()
	case ptpUnknown:
		rb.SetPtpClockTypeUNKNOWN()
	}
}

func parsePMCStatus(output []byte) (ptpStatus, error) {
	var status ptpStatus
	datasets, err := parsePMCDatasets(output)
	if err != nil {
		return status, err
	}
	if len(datasets) == 0 {
		preview := strings.TrimSpace(string(output))
		if len(preview) > pmcResponsePreviewBytes {
			preview = preview[:pmcResponsePreviewBytes] + "..."
		}
		return status, fmt.Errorf("pmc response missing supported management datasets (output: %q)", preview)
	}

	var parseErrors []error
	status, err = parsePMCDatasetResponses(datasets[pmcTimeStatusDataset], pmcTimeStatusDataset, status, parseTimeStatus)
	if err != nil {
		parseErrors = append(parseErrors, err)
	}
	status, err = parsePMCDatasetResponses(datasets[pmcCurrentDataSet], pmcCurrentDataSet, status, parsePathDelay)
	if err != nil {
		parseErrors = append(parseErrors, err)
	}
	status, err = parsePMCDatasetResponses(datasets[pmcClockDescription], pmcClockDescription, status, parseClockType)
	if err != nil {
		parseErrors = append(parseErrors, err)
	}
	if err := parsePortDataSets(datasets[pmcPortDataSet], &status); err != nil {
		parseErrors = append(parseErrors, err)
	}
	if err := errors.Join(parseErrors...); err != nil {
		return status, err
	}
	if !status.hasTimeStatus && len(status.ports) == 0 && status.clockType == "" {
		return status, fmt.Errorf("pmc response missing %s, %s, and %s", pmcTimeStatusDataset, pmcPortDataSet, pmcClockDescription)
	}
	return status, nil
}

func parsePMCDatasetResponses(datasets []pmcDataset, name string, status ptpStatus, parse func(pmcDataset, *ptpStatus) error) (ptpStatus, error) {
	if len(datasets) == 0 {
		return status, nil
	}

	var firstValidStatus *ptpStatus
	for _, dataset := range datasets {
		candidate := status
		if err := parse(dataset, &candidate); err == nil && firstValidStatus == nil {
			firstValidStatus = &candidate
		}
	}
	if firstValidStatus == nil {
		return status, fmt.Errorf("pmc response has no valid %s dataset", name)
	}
	return *firstValidStatus, nil
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
		// pmc prints each response header followed by indented `field value`
		// lines. Associate each two-token field line with the most recent header;
		// for example, `    portState SLAVE` is stored as `portState: SLAVE`.
		// Lines with a different shape and preamble before the first header are ignored.
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
		return fmt.Errorf("invalid %s %s %q", pmcTimeStatusDataset, pmcGrandmasterIdentity, status.gmIdentity)
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
		return fmt.Errorf("invalid %s %q", pmcMeanPathDelayField, value)
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
		return fmt.Errorf("invalid %s %q: %w", pmcClockTypeField, value, err)
	}
	status.clockType = clockTypeName(clockType)
	return nil
}

func parsePortDataSets(datasets []pmcDataset, status *ptpStatus) error {
	seenPorts := make(map[string]bool)
	validResponses := 0
	for _, dataset := range datasets {
		identity, identityErr := getRequiredField(dataset, pmcPortIdentityField)
		state, stateErr := getRequiredField(dataset, pmcPortStateField)
		if identityErr != nil || stateErr != nil {
			continue
		}
		if !validPortIdentity(identity) || !validPortState(state) {
			continue
		}
		validResponses++
		if !seenPorts[identity] {
			status.ports = append(status.ports, ptpPort{identity: identity, state: state})
			seenPorts[identity] = true
		}
	}
	if len(datasets) > 0 && validResponses == 0 {
		return fmt.Errorf("pmc response has no valid %s dataset", pmcPortDataSet)
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
		ptpPortStateSlave, ptpPortStateGrandMaster:
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
	// A boundary clock can have both slave and master ports. Prefer SLAVE to
	// describe its upstream synchronization; GRAND_MASTER has the clock-level
	// meaning MASTER. The remaining priority is UNCALIBRATED, then FAULTY.
	if hasPortState(ports, ptpPortStateSlave) {
		return ptpPortStateSlave
	}
	if hasPortState(ports, ptpPortStateMaster, ptpPortStateGrandMaster) {
		return ptpPortStateMaster
	}
	if hasPortState(ports, ptpPortStateUncalibrated) {
		return ptpPortStateUncalibrated
	}
	if hasPortState(ports, ptpPortStateFaulty) {
		return ptpPortStateFaulty
	}
	return ptpClockStateUnsynchronized
}

func hasPortState(ports []ptpPort, states ...string) bool {
	for _, port := range ports {
		for _, state := range states {
			if port.state == state {
				return true
			}
		}
	}
	return false
}
