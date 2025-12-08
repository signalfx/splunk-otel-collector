package http

import (
	"sync"
	"testing"

	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/event"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/trace"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func TestMonitorCanReachMicrosoft(t *testing.T) {
	output := &fakeOutput{}
	output.sendDatapointsCall.Add(1)

	monitor := &Monitor{
		Output:      output,
		monitorName: "http-monitor",
	}

	// Configure the monitor to check microsoft.com
	err := monitor.Configure(&Config{
		MonitorConfig: config.MonitorConfig{
			IntervalSeconds: 1,
		},
		HTTPConfig: httpclient.HTTPConfig{
			UseHTTPS: true,
		},
		Host:   "www.microsoft.com",
		Port:   443,
		Path:   "/",
		Method: "GET",
	})
	require.NoError(t, err)
	defer monitor.Shutdown()

	output.sendDatapointsCall.Wait()
	require.NotEmpty(t, output.datapoints, "Expected to receive datapoints")

	// Check for specific metrics
	var foundStatusCode, foundResponseTime, foundCertExpiry, foundCertValid bool
	for _, dp := range output.datapoints {
		switch dp.Metric {
		case "http.status_code":
			foundStatusCode = true
			assert.Equal(t, int64(200), dp.Value.(datapoint.IntValue).Int())
		case "http.response_time":
			foundResponseTime = true
			assert.Greater(t, dp.Value.(datapoint.FloatValue).Float(), float64(0))
		case "http.cert_expiry":
			foundCertExpiry = true
			assert.Greater(t, dp.Value.(datapoint.FloatValue).Float(), float64(0))
		case "http.cert_valid":
			foundCertValid = true
			assert.Greater(t, dp.Value.(datapoint.IntValue).Int(), int64(0))
		}
	}

	require.True(t, foundStatusCode, "Expected to find http.status_code metric")
	require.True(t, foundResponseTime, "Expected to find http.response_time metric")
	require.True(t, foundCertExpiry, "Expected to find http.cert_expiry metric")
	require.True(t, foundCertValid, "Expected to find http.cert_valid metric")
}

type fakeOutput struct {
	sendDatapointsCall sync.WaitGroup
	datapoints         []*datapoint.Datapoint
}

var _ types.FilteringOutput = (*fakeOutput)(nil)

// AddDatapointExclusionFilter implements types.FilteringOutput.
func (fo *fakeOutput) AddDatapointExclusionFilter(_ dpfilters.DatapointFilter) {
	panic("unimplemented")
}

// AddExtraDimension implements types.FilteringOutput.
func (fo *fakeOutput) AddExtraDimension(_, _ string) {
	panic("unimplemented")
}

// Copy implements types.FilteringOutput.
func (fo *fakeOutput) Copy() types.Output {
	panic("unimplemented")
}

// EnabledMetrics implements types.FilteringOutput.
func (fo *fakeOutput) EnabledMetrics() []string {
	panic("unimplemented")
}

// HasAnyExtraMetrics implements types.FilteringOutput.
func (fo *fakeOutput) HasAnyExtraMetrics() bool {
	panic("unimplemented")
}

// HasEnabledMetricInGroup implements types.FilteringOutput.
func (fo *fakeOutput) HasEnabledMetricInGroup(_ string) bool {
	panic("unimplemented")
}

// SendDimensionUpdate implements types.FilteringOutput.
func (fo *fakeOutput) SendDimensionUpdate(_ *types.Dimension) {
	panic("unimplemented")
}

// SendEvent implements types.FilteringOutput.
func (fo *fakeOutput) SendEvent(_ *event.Event) {
	panic("unimplemented")
}

// SendMetrics implements types.FilteringOutput.
func (fo *fakeOutput) SendMetrics(_ ...pmetric.Metric) {
	panic("unimplemented")
}

// SendSpans implements types.FilteringOutput.
func (fo *fakeOutput) SendSpans(_ ...*trace.Span) {
	panic("unimplemented")
}

func (fo *fakeOutput) SendDatapoints(dps ...*datapoint.Datapoint) {
	fo.datapoints = append(fo.datapoints, dps...)
	fo.sendDatapointsCall.Done()
}
