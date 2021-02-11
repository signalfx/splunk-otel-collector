// Copyright 2021, OpenTelemetry Authors
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

package smartagentreceiver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/zap"
)

var expectedCPUMetrics = map[string]pdata.MetricDataType{
	"cpu.idle":                 pdata.MetricDataTypeDoubleSum,
	"cpu.interrupt":            pdata.MetricDataTypeDoubleSum,
	"cpu.nice":                 pdata.MetricDataTypeDoubleSum,
	"cpu.num_processors":       pdata.MetricDataTypeIntGauge,
	"cpu.softirq":              pdata.MetricDataTypeDoubleSum,
	"cpu.steal":                pdata.MetricDataTypeDoubleSum,
	"cpu.system":               pdata.MetricDataTypeDoubleSum,
	"cpu.user":                 pdata.MetricDataTypeDoubleSum,
	"cpu.utilization":          pdata.MetricDataTypeDoubleGauge,
	"cpu.utilization_per_core": pdata.MetricDataTypeDoubleGauge,
	"cpu.wait":                 pdata.MetricDataTypeDoubleSum,
}

func newConfig(nameVal, monitorType string, intervalSeconds int) Config {
	return Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: fmt.Sprintf("%s/%s", typeStr, nameVal),
		},
		monitorConfig: &cpu.Config{
			MonitorConfig: config.MonitorConfig{
				Type:            monitorType,
				IntervalSeconds: intervalSeconds,
				ExtraDimensions: map[string]string{
					"required_dimension": "required_value",
				},
			},
			ReportPerCPU: true,
		},
	}
}

func TestSmartAgentReceiver(t *testing.T) {
	cfg := newConfig("valid", "cpu", 10)
	consumer := new(consumertest.MetricsSink)
	receiver := NewReceiver(zap.NewNop(), cfg, consumer)

	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	assert.EqualValues(t, "smartagentvalid", cfg.monitorConfig.MonitorConfigCore().MonitorID)
	monitor, isMonitor := receiver.monitor.(*cpu.Monitor)
	require.True(t, isMonitor)

	monitorOutput := monitor.Output
	_, isOutput := monitorOutput.(*Output)
	assert.True(t, isOutput)

	assert.Eventuallyf(t, func() bool {
		// confirm single occurrence of total metrics as sanity in lieu of
		// out of scope cpu monitor verification.
		seenTotalMetric := map[string]bool{}

		allMetrics := consumer.AllMetrics()
		for _, m := range allMetrics {
			resourceMetrics := m.ResourceMetrics()
			for i := 0; i < resourceMetrics.Len(); i++ {
				resourceMetric := resourceMetrics.At(i)
				instrumentationLibraryMetrics := resourceMetric.InstrumentationLibraryMetrics()
				for j := 0; j < instrumentationLibraryMetrics.Len(); j++ {
					instrumentationLibraryMetric := instrumentationLibraryMetrics.At(j)
					metrics := instrumentationLibraryMetric.Metrics()
					for k := 0; k < metrics.Len(); k++ {
						metric := metrics.At(k)
						name := metric.Name()
						dataType := metric.DataType()
						expectedDataType := expectedCPUMetrics[name]
						require.NotEqual(t, pdata.MetricDataTypeNone, expectedDataType, "received unexpected none type for %s", name)
						assert.Equal(t, expectedDataType, dataType)
						var labels pdata.StringMap
						switch dataType {
						case pdata.MetricDataTypeIntGauge:
							ig := metric.IntGauge()
							for l := 0; l < ig.DataPoints().Len(); l++ {
								igdp := ig.DataPoints().At(l)
								labels = igdp.LabelsMap()
								var val interface{} = igdp.Value()
								_, ok := val.(int64)
								assert.True(t, ok, "invalid value of MetricDataTypeIntGauge metric %s", name)
							}
						case pdata.MetricDataTypeDoubleGauge:
							dg := metric.DoubleGauge()
							for l := 0; l < dg.DataPoints().Len(); l++ {
								dgdp := dg.DataPoints().At(l)
								labels = dgdp.LabelsMap()
								var val interface{} = dgdp.Value()
								_, ok := val.(float64)
								assert.True(t, ok, "invalid value of MetricDataTypeDoubleGauge metric %s", name)
							}
						case pdata.MetricDataTypeDoubleSum:
							ds := metric.DoubleSum()
							for l := 0; l < ds.DataPoints().Len(); l++ {
								dsdp := ds.DataPoints().At(l)
								labels = dsdp.LabelsMap()
								var val interface{} = dsdp.Value()
								_, ok := val.(float64)
								assert.True(t, ok, "invalid value of MetricDataTypeDoubleSum metric %s", name)
							}
						default:
							t.Errorf("unexpected type %#v for metric %s", metric.DataType(), name)
						}

						labelVal, ok := labels.Get("required_dimension")
						require.True(t, ok)
						assert.Equal(t, "required_value", labelVal)

						// mark metric as having been seen
						cpuNum, _ := labels.Get("cpu")
						seenName := fmt.Sprintf("%s%s", name, cpuNum)
						assert.False(t, seenTotalMetric[seenName], "unexpectedly repeated metric: %v", seenName)
						seenTotalMetric[seenName] = true
					}
				}
			}
		}
		return len(allMetrics) > 0
	}, 5*time.Second, 1*time.Millisecond, "failed to receive expected cpu metrics")

	metrics := consumer.AllMetrics()
	assert.Greater(t, len(metrics), 0)
	err = receiver.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestStartReceiverWithInvalidMonitorConfig(t *testing.T) {
	cfg := newConfig("invalid", "cpu", -123)
	receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	assert.EqualError(t, err,
		"config validation failed for \"smartagent/invalid\": intervalSeconds must be greater than 0s (-123 provided)",
	)
}

func TestStartReceiverWithUnknownMonitorType(t *testing.T) {
	cfg := newConfig("invalid", "notamonitortype", 1)
	receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	assert.EqualError(t, err,
		"failed creating monitor \"notamonitortype\": unable to find MonitorFactory for \"notamonitortype\"",
	)
}

func TestMultipleStartAndShutdownInvocations(t *testing.T) {
	cfg := newConfig("valid", "cpu", 1)
	receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Equal(t, err, componenterror.ErrAlreadyStarted)

	err = receiver.Shutdown(context.Background())
	require.NoError(t, err)

	err = receiver.Shutdown(context.Background())
	require.Error(t, err)
	assert.Equal(t, err, componenterror.ErrAlreadyStopped)
}

func TestOutOfOrderShutdownInvocations(t *testing.T) {
	cfg := newConfig("valid", "cpu", 1)
	receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())

	err := receiver.Shutdown(context.Background())
	require.Error(t, err)
	assert.EqualError(t, err,
		"smartagentreceiver's Shutdown() called before Start() or with invalid monitor state",
	)
}

func TestInvalidMonitorStateAtShutdown(t *testing.T) {
	cfg := newConfig("valid", "cpu", 1)
	receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
	receiver.monitor = new(interface{})

	err := receiver.Shutdown(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid monitor state at Shutdown(): (*interface {})")
}

func TestConfirmStartingReceiverWithInvalidMonitorInstancesDoesntPanic(t *testing.T) {
	tests := []struct {
		name           string
		monitorFactory func() interface{}
		expectedError  string
	}{
		{"anonymous struct", func() interface{} { return struct{}{} }, ""},
		{"anonymous struct pointer", func() interface{} { return &struct{}{} }, ""},
		{"nil interface pointer", func() interface{} { return new(interface{}) }, ": invalid struct instance: (*interface {})"},
		{"nil", func() interface{} { return nil }, ": invalid struct instance: <nil>"},
		{"boolean", func() interface{} { return false }, ": invalid struct instance: false"},
		{"string", func() interface{} { return "asdf" }, ": invalid struct instance: \"asdf\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			monitors.MonitorFactories["notarealmonitor"] = test.monitorFactory

			cfg := newConfig("invalid", "notarealmonitor", 123)
			receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
			err := receiver.Start(context.Background(), componenttest.NewNopHost())
			require.Error(tt, err)
			assert.Contains(tt, err.Error(),
				fmt.Sprintf("failed creating monitor \"notarealmonitor\": unable to set Output field of monitor%s", test.expectedError),
			)
		})
	}
}

func TestSmartAgentReceiverNonCollectd(t *testing.T) {
	cfg := newConfig("valid", "cpu", 10)
	consumer := new(consumertest.MetricsSink)
	receiver := NewReceiver(zap.NewNop(), cfg, consumer)

	receiver.Start(context.Background(), componenttest.NewNopHost())

	// Simple test to ensure that collectd config is not loaded for non-collectd monitors.
	require.Empty(t, receiver.config.collectdConfig)

	receiver.Shutdown(context.Background())
}
