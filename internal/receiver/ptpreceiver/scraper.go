// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
		"-d", strconv.Itoa(s.config.DomainNumber), "GET TIME_STATUS_NP")
	if err != nil {
		return metrics, fmt.Errorf("query ptp4l with pmc: %w: %s", err, strings.TrimSpace(string(output)))
	}
	offset, present, err := parseTimeStatus(output)
	if err != nil {
		return metrics, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	var grandmasterPresent int64
	if present {
		grandmasterPresent = 1
	}
	s.mb.RecordPtpGrandmasterPresentDataPoint(now, grandmasterPresent)
	if present {
		s.mb.RecordPtpOffsetDataPoint(now, offset)
	}
	rb := s.mb.NewResourceBuilder()
	rb.SetPtpSocketPath(s.config.SocketPath)
	return s.mb.Emit(metadata.WithResource(rb.Emit())), nil
}

// parseTimeStatus only accepts fields from a successful TIME_STATUS_NP response.
func parseTimeStatus(output []byte) (int64, bool, error) {
	var offset int64
	var present bool
	var haveOffset, havePresent, inResponse bool
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "RESPONSE MANAGEMENT TIME_STATUS_NP") {
			inResponse = true
			continue
		}
		if !inResponse {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "master_offset":
			var err error
			offset, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return 0, false, fmt.Errorf("invalid master_offset %q: %w", fields[1], err)
			}
			haveOffset = true
		case "gmPresent":
			var err error
			present, err = strconv.ParseBool(fields[1])
			if err != nil {
				return 0, false, fmt.Errorf("invalid gmPresent %q: %w", fields[1], err)
			}
			havePresent = true
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, false, fmt.Errorf("read pmc response: %w", err)
	}
	if !haveOffset || !havePresent {
		return 0, false, errors.New("pmc response missing TIME_STATUS_NP master_offset or gmPresent")
	}
	return offset, present, nil
}
