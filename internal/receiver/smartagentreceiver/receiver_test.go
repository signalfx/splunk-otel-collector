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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

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
			},
		},
	}
}

func TestSmartAgentReceiver(t *testing.T) {
	cfg := newConfig("valid", "cpu", 1)
	observed, logs := observer.New(zapcore.DebugLevel)
	receiver := NewReceiver(zap.New(observed), cfg, consumertest.NewMetricsNop())
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	assert.EqualValues(t, "smartagentvalid", cfg.monitorConfig.MonitorConfigCore().MonitorID)

	monitorOutput := receiver.monitor.(*cpu.Monitor).Output
	_, ok := monitorOutput.(*Output)
	assert.True(t, ok)

	assert.Eventuallyf(t, func() bool {
		filtered := logs.FilterMessageSnippet("SendDatapoints has been called.")
		return len(filtered.All()) == 1
	}, 5*time.Second, 1*time.Millisecond, "failed to receive any metrics from monitor")

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
		{"anonymous struct", func() interface{} { return struct{}{} }, "struct {}{}"},
		{"anonymous struct pointer", func() interface{} { return &struct{}{} }, "&struct {}{}"},
		{"nil interface pointer", func() interface{} { return new(interface{}) }, "(*interface {})"},
		{"nil", func() interface{} { return nil }, "<nil>"},
		{"boolean", func() interface{} { return false }, "false"},
		{"string", func() interface{} { return "asdf" }, "\"asdf\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			monitors.MonitorFactories["notarealmonitor"] = test.monitorFactory

			cfg := newConfig("invalid", "notarealmonitor", 123)
			receiver := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
			err := receiver.Start(context.Background(), componenttest.NewNopHost())
			require.Error(tt, err)
			assert.Contains(tt, err.Error(),
				fmt.Sprintf("failed creating monitor \"notarealmonitor\": invalid monitor instance: %s", test.expectedError),
			)
		})
	}
}
