// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"context"
	"errors"
	"os"
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
  001122.fffe.334455-0 seq 1 RESPONSE MANAGEMENT CURRENT_DATA_SET
    stepsRemoved 1
    meanPathDelay 125.5
  001122.fffe.334455-1 seq 2 RESPONSE MANAGEMENT PORT_DATA_SET
    portIdentity 001122.fffe.334455-1
    portState SLAVE
  001122.fffe.334455-2 seq 2 RESPONSE MANAGEMENT PORT_DATA_SET
    portIdentity 001122.fffe.334455-2
    portState MASTER
  001122.fffe.334455-1 seq 3 RESPONSE MANAGEMENT CLOCK_DESCRIPTION
    clockType 0x4000
`

func TestParsePMCStatus(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		clockState string
		present    bool
		wantErr    bool
	}{
		{name: "valid boundary clock", output: pmcResponse, present: true, clockState: "SLAVE"},
		{name: "local grandmaster", output: pmcResponseWithoutGrandmaster(), clockState: "MASTER"},
		{name: "local grandmaster ignores invalid offset and delay", output: pmcResponseWithoutGrandmasterWithInvalidMeasurements(), clockState: "MASTER"},
		{name: "transparent clock", output: "001122.fffe.334455-0 seq 0 RESPONSE MANAGEMENT CLOCK_DESCRIPTION\nclockType 0x2000", clockState: "UNKNOWN"},
		{name: "no response", output: "sending: GET TIME_STATUS_NP", wantErr: true},
		{name: "bad offset", output: strings.Replace(pmcResponse, "master_offset -42", "master_offset invalid", 1), wantErr: true},
		{name: "bad grandmaster flag", output: strings.Replace(pmcResponse, "gmPresent true", "gmPresent invalid", 1), wantErr: true},
		{name: "missing grandmaster", output: strings.Replace(pmcResponse, "gmPresent true", "", 1), wantErr: true},
		{name: "missing identity", output: strings.Replace(pmcResponse, "gmIdentity 001122.fffe.334455", "", 1), wantErr: true},
		{name: "bad path delay", output: strings.Replace(pmcResponse, "meanPathDelay 125.5", "meanPathDelay NaN", 1), wantErr: true},
		{name: "bad port state", output: strings.Replace(pmcResponse, "portState SLAVE", "portState invalid", 1), wantErr: true},
		{name: "bad clock type", output: strings.Replace(pmcResponse, "clockType 0x4000", "clockType invalid", 1), wantErr: true},
		{name: "oversized response", output: pmcResponse + strings.Repeat("x", 70<<10), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := parsePMCStatus([]byte(tt.output))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.present, status.gmPresent)
			require.Equal(t, tt.clockState, clockState(status.ports))
			if tt.name == "valid boundary clock" {
				require.Equal(t, int64(-42), status.offset)
				require.InDelta(t, 125.5, status.pathDelay, 1e-6)
				require.Equal(t, "001122.fffe.334455", status.gmIdentity)
				require.Equal(t, "BC", status.clockType)
				require.Len(t, status.ports, 2)
			}
			if tt.name == "transparent clock" {
				require.Equal(t, "P2P_TC", status.clockType)
				require.False(t, status.hasTimeStatus)
			}
		})
	}
}

func TestScrape(t *testing.T) {
	for _, present := range []bool{true, false} {
		t.Run(map[bool]string{true: "remote grandmaster", false: "local grandmaster"}[present], func(t *testing.T) {
			config := createDefaultConfig().(*Config)
			config.PMC.ClientSocketDirectory = t.TempDir()
			var gotName string
			var gotArgs []string
			var clientSocketPath string
			s := newScraper(config, receiver.Settings{})
			expectedClockState := "SLAVE"
			expectedPortStates := map[string]string{"001122.fffe.334455-1": "SLAVE", "001122.fffe.334455-2": "MASTER"}
			s.runPMC = func(_ context.Context, name string, args ...string) ([]byte, error) {
				gotName, gotArgs = name, args
				clientSocketPath = args[4]
				require.DirExists(t, filepath.Dir(clientSocketPath))
				require.Equal(t, config.PMC.ClientSocketDirectory, filepath.Dir(filepath.Dir(clientSocketPath)))
				response := pmcResponse
				if !present {
					response = pmcResponseWithoutGrandmasterWithInvalidMeasurements()
					expectedClockState = "MASTER"
					expectedPortStates = map[string]string{"001122.fffe.334455-1": "LISTENING", "001122.fffe.334455-2": "MASTER"}
				}
				return []byte(response), nil
			}
			metrics, err := s.scrape(context.Background())
			require.NoError(t, err)
			require.Equal(t, "pmc", gotName)
			require.Equal(t, []string{
				"-u", "-b", "0", "-i", clientSocketPath, "-s", config.SocketPath, "-d", "0",
				"GET TIME_STATUS_NP", "GET CURRENT_DATA_SET", "GET PORT_DATA_SET", "GET CLOCK_DESCRIPTION",
			}, gotArgs)
			_, err = os.Stat(filepath.Dir(clientSocketPath))
			require.ErrorIs(t, err, os.ErrNotExist)
			require.Equal(t, 1, metrics.ResourceMetrics().Len())
			rm := metrics.ResourceMetrics().At(0)
			value, ok := rm.Resource().Attributes().Get("ptp.socket_path")
			require.True(t, ok)
			require.Equal(t, config.SocketPath, value.Str())
			clockType, ok := rm.Resource().Attributes().Get("ptp.clock.type")
			require.True(t, ok)
			require.Equal(t, "BC", clockType.Str())
			require.Equal(t, 1, rm.ScopeMetrics().Len())
			require.Equal(t, metadata.ScopeName, rm.ScopeMetrics().At(0).Scope().Name())
			got := map[string]pmetric.Metric{}
			for _, metric := range allMetrics(rm.ScopeMetrics().At(0).Metrics()) {
				got[metric.Name()] = metric
			}
			want := []string{"ptp.clock.state", "ptp.grandmaster.present", "ptp.port.state"}
			for _, name := range want {
				require.Contains(t, got, name)
			}
			require.Equal(t, int64(1), got["ptp.clock.state"].Gauge().DataPoints().At(0).IntValue())
			clockStateValue, ok := got["ptp.clock.state"].Gauge().DataPoints().At(0).Attributes().Get("ptp.clock.state")
			require.True(t, ok)
			require.Equal(t, expectedClockState, clockStateValue.Str())
			ports := got["ptp.port.state"].Gauge().DataPoints()
			require.Equal(t, 2, ports.Len())
			portStates := map[string]string{}
			for i := 0; i < ports.Len(); i++ {
				identity, found := ports.At(i).Attributes().Get("ptp.port.identity")
				require.True(t, found)
				state, found := ports.At(i).Attributes().Get("ptp.port.state")
				require.True(t, found)
				portStates[identity.Str()] = state.Str()
			}
			require.Equal(t, expectedPortStates, portStates)
			if present {
				want = append(want, "ptp.grandmaster.info", "ptp.offset", "ptp.path.delay")
				for _, name := range want {
					require.Contains(t, got, name)
				}
				require.Equal(t, int64(1), got["ptp.grandmaster.present"].Gauge().DataPoints().At(0).IntValue())
				require.InDelta(t, float64(-42)/nanosecondsPerSecond, got["ptp.offset"].Gauge().DataPoints().At(0).DoubleValue(), 1e-18)
				require.InDelta(t, 125.5/nanosecondsPerSecond, got["ptp.path.delay"].Gauge().DataPoints().At(0).DoubleValue(), 1e-18)
				gmIdentity, found := got["ptp.grandmaster.info"].Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
				require.True(t, found)
				require.Equal(t, "001122.fffe.334455", gmIdentity.Str())
			} else {
				require.Equal(t, int64(0), got["ptp.grandmaster.present"].Gauge().DataPoints().At(0).IntValue())
			}
			require.Len(t, got, len(want))
		})
	}
}

func pmcResponseWithoutGrandmaster() string {
	response := strings.Replace(pmcResponse, "gmPresent true", "gmPresent false", 1)
	return strings.Replace(response, "portState SLAVE", "portState LISTENING", 1)
}

func pmcResponseWithoutGrandmasterWithInvalidMeasurements() string {
	response := pmcResponseWithoutGrandmaster()
	response = strings.Replace(response, "master_offset -42", "master_offset invalid", 1)
	return strings.Replace(response, "meanPathDelay 125.5", "meanPathDelay invalid", 1)
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
	require.ErrorContains(t, err, "missing supported management datasets")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeMissingPMC(t *testing.T) {
	config := createDefaultConfig().(*Config)
	config.PMC.Path = filepath.Join(t.TempDir(), "missing-pmc")
	s := newScraper(config, receiver.Settings{})
	metrics, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "query ptp4l with pmc")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeInvalidPMCClientSocketDirectory(t *testing.T) {
	config := createDefaultConfig().(*Config)
	config.PMC.ClientSocketDirectory = filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(config.PMC.ClientSocketDirectory, []byte("file"), 0o600))
	s := newScraper(config, receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		t.Fatal("pmc should not run when its client socket directory cannot be created")
		return nil, nil
	}
	metrics, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "create temporary pmc client socket directory")
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeDisabledMetrics(t *testing.T) {
	config := createDefaultConfig().(*Config)
	config.Metrics.PtpClockState.Enabled = false
	config.Metrics.PtpGrandmasterInfo.Enabled = false
	config.Metrics.PtpGrandmasterPresent.Enabled = false
	config.Metrics.PtpOffset.Enabled = false
	config.Metrics.PtpPathDelay.Enabled = false
	config.Metrics.PtpPortState.Enabled = false
	s := newScraper(config, receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte(pmcResponse), nil
	}
	metrics, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestScrapeGrandmasterChange(t *testing.T) {
	s := newScraper(createDefaultConfig().(*Config), receiver.Settings{})
	identity := "001122.fffe.334455"
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte(strings.Replace(pmcResponse, "gmIdentity 001122.fffe.334455", "gmIdentity "+identity, 1)), nil
	}
	first, err := s.scrape(context.Background())
	require.NoError(t, err)
	identity = "aabbcc.fffe.ddeeff"
	second, err := s.scrape(context.Background())
	require.NoError(t, err)
	firstGM := findMetric(t, first, "ptp.grandmaster.info")
	secondGM := findMetric(t, second, "ptp.grandmaster.info")
	firstID, ok := firstGM.Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
	require.True(t, ok)
	secondID, ok := secondGM.Gauge().DataPoints().At(0).Attributes().Get("ptp.grandmaster.identity")
	require.True(t, ok)
	require.Equal(t, "001122.fffe.334455", firstID.Str())
	require.Equal(t, "aabbcc.fffe.ddeeff", secondID.Str())
}

func TestScrapeTransparentClock(t *testing.T) {
	s := newScraper(createDefaultConfig().(*Config), receiver.Settings{})
	s.runPMC = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("001122.fffe.334455-0 seq 0 RESPONSE MANAGEMENT CLOCK_DESCRIPTION\nclockType 0x1000\n"), nil
	}
	metrics, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, metrics.ResourceMetrics().Len())
	rm := metrics.ResourceMetrics().At(0)
	clockType, ok := rm.Resource().Attributes().Get("ptp.clock.type")
	require.True(t, ok)
	require.Equal(t, "E2E_TC", clockType.Str())
	metric := findMetric(t, metrics, "ptp.clock.state")
	state, ok := metric.Gauge().DataPoints().At(0).Attributes().Get("ptp.clock.state")
	require.True(t, ok)
	require.Equal(t, "UNKNOWN", state.Str())
}

func findMetric(t *testing.T, metrics pmetric.Metrics, name string) pmetric.Metric {
	t.Helper()
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		for j := 0; j < metrics.ResourceMetrics().At(i).ScopeMetrics().Len(); j++ {
			for _, metric := range allMetrics(metrics.ResourceMetrics().At(i).ScopeMetrics().At(j).Metrics()) {
				if metric.Name() == name {
					return metric
				}
			}
		}
	}
	require.FailNow(t, "metric not found", name)
	return pmetric.Metric{}
}

func allMetrics(metrics pmetric.MetricSlice) []pmetric.Metric {
	result := make([]pmetric.Metric, 0, metrics.Len())
	for i := 0; i < metrics.Len(); i++ {
		result = append(result, metrics.At(i))
	}
	return result
}
