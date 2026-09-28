// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

const pmcResponse = `sending: GET TIME_STATUS_NP
  001122.fffe.334455-0 seq 0 RESPONSE MANAGEMENT TIME_STATUS_NP
    master_offset -42
    ingress_time 123456789
    gmPresent true
    gmIdentity 001122.fffe.334455
`

func TestParseTimeStatus(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		offset  int64
		present bool
		wantErr bool
	}{
		{name: "valid", output: pmcResponse, offset: -42, present: true},
		{name: "no grandmaster", output: strings.Replace(pmcResponse, "gmPresent true", "gmPresent false", 1), offset: -42},
		{name: "no response", output: "sending: GET TIME_STATUS_NP", wantErr: true},
		{name: "bad offset", output: strings.Replace(pmcResponse, "master_offset -42", "master_offset invalid", 1), wantErr: true},
		{name: "bad grandmaster flag", output: strings.Replace(pmcResponse, "gmPresent true", "gmPresent invalid", 1), wantErr: true},
		{name: "missing grandmaster", output: strings.Replace(pmcResponse, "gmPresent true", "", 1), wantErr: true},
		{name: "oversized response", output: pmcResponse + strings.Repeat("x", 70<<10), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, present, err := parseTimeStatus([]byte(tt.output))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.offset, offset)
			require.Equal(t, tt.present, present)
		})
	}
}

func TestScrape(t *testing.T) {
	for _, present := range []bool{true, false} {
		t.Run(map[bool]string{true: "grandmaster", false: "no grandmaster"}[present], func(t *testing.T) {
			config := createDefaultConfig().(*Config)
			var gotName string
			var gotArgs []string
			s := newScraper(config, receiver.Settings{})
			s.runPMC = func(_ context.Context, name string, args ...string) ([]byte, error) {
				gotName, gotArgs = name, args
				response := pmcResponse
				if !present {
					response = strings.Replace(response, "gmPresent true", "gmPresent false", 1)
				}
				return []byte(response), nil
			}
			metrics, err := s.scrape(context.Background())
			require.NoError(t, err)
			require.Equal(t, "pmc", gotName)
			require.Equal(t, []string{"-u", "-b", "0", "-s", config.SocketPath, "-d", "0", "GET TIME_STATUS_NP"}, gotArgs)
			require.Equal(t, 1, metrics.ResourceMetrics().Len())
			rm := metrics.ResourceMetrics().At(0)
			value, ok := rm.Resource().Attributes().Get("ptp.socket_path")
			require.True(t, ok)
			require.Equal(t, config.SocketPath, value.Str())
			require.Equal(t, 1, rm.ScopeMetrics().Len())
			require.Equal(t, metadata.ScopeName, rm.ScopeMetrics().At(0).Scope().Name())
			got := map[string]int64{}
			for _, metric := range allMetrics(rm.ScopeMetrics().At(0).Metrics()) {
				got[metric.Name()] = metric.Gauge().DataPoints().At(0).IntValue()
			}
			want := map[string]int64{"ptp.grandmaster.present": 0}
			if present {
				want["ptp.grandmaster.present"] = 1
				want["ptp.offset"] = -42
			}
			require.Equal(t, want, got)
		})
	}
}

func TestScrapeQueryError(t *testing.T) {
	s := newScraper(createDefaultConfig().(*Config), receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("connection failed"), errors.New("exit status 1")
	}
	metrics, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "connection failed")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeInvalidResponse(t *testing.T) {
	s := newScraper(createDefaultConfig().(*Config), receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("unexpected response"), nil
	}
	metrics, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "missing TIME_STATUS_NP")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeMissingPMC(t *testing.T) {
	config := createDefaultConfig().(*Config)
	config.PMCPath = filepath.Join(t.TempDir(), "missing-pmc")
	s := newScraper(config, receiver.Settings{})
	metrics, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "query ptp4l with pmc")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeDisabledMetrics(t *testing.T) {
	config := createDefaultConfig().(*Config)
	config.Metrics.PtpGrandmasterPresent.Enabled = false
	config.Metrics.PtpOffset.Enabled = false
	s := newScraper(config, receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte(pmcResponse), nil
	}
	metrics, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func allMetrics(metrics pmetric.MetricSlice) []pmetric.Metric {
	result := make([]pmetric.Metric, 0, metrics.Len())
	for i := 0; i < metrics.Len(); i++ {
		result = append(result, metrics.At(i))
	}
	return result
}
