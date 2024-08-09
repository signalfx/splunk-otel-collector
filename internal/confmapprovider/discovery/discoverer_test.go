// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	otelcolextension "go.opentelemetry.io/collector/extension"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestInnerDiscoveryExecution(t *testing.T) {
	tests := []struct {
		name        string
		observers   map[component.ID]otelcolextension.Extension
		receivers   map[component.ID]otelcolreceiver.Logs
		expectedMsg string
	}{
		{
			name: "happy_path",
			observers: map[component.ID]otelcolextension.Extension{
				component.MustNewID("observer00"): &mockExtension{},
				component.MustNewID("observer01"): &mockExtension{},
				component.MustNewID("observer02"): &mockExtension{},
			},
			receivers: map[component.ID]otelcolreceiver.Logs{
				component.MustNewID("receiver00"): &mockReceiverLogs{},
				component.MustNewID("receiver01"): &mockReceiverLogs{},
			},
		},
		{
			name: "fail_start_extension",
			observers: map[component.ID]otelcolextension.Extension{
				component.MustNewID("observer00"): &mockExtension{},
				component.MustNewID("observer01"): &mockExtension{mockComponent{startErr: fmt.Errorf("extension_start_error")}},
				component.MustNewID("observer02"): &mockExtension{},
			},
			receivers: map[component.ID]otelcolreceiver.Logs{
				component.MustNewID("receiver00"): &mockReceiverLogs{},
			},
			expectedMsg: "extension_start_error",
		},
		{
			name: "fail_start_receiver",
			observers: map[component.ID]otelcolextension.Extension{
				component.MustNewID("observer00"): &mockExtension{},
				component.MustNewID("observer01"): &mockExtension{},
			},
			receivers: map[component.ID]otelcolreceiver.Logs{
				component.MustNewID("receiver00"): &mockReceiverLogs{},
				component.MustNewID("receiver01"): &mockReceiverLogs{mockComponent{startErr: fmt.Errorf("receiver_start_error")}},
				component.MustNewID("receiver02"): &mockReceiverLogs{},
			},
			expectedMsg: "receiver_start_error",
		},
		{
			name: "fail_shutdown_no_error_msg",
			observers: map[component.ID]otelcolextension.Extension{
				component.MustNewID("observer00"): &mockExtension{},
				component.MustNewID("observer01"): &mockExtension{},
			},
			receivers: map[component.ID]otelcolreceiver.Logs{
				component.MustNewID("receiver00"): &mockReceiverLogs{},
				component.MustNewID("receiver01"): &mockReceiverLogs{mockComponent{shutdownErr: fmt.Errorf("receiver_shutdown_error")}},
				component.MustNewID("receiver02"): &mockReceiverLogs{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := newDiscoverer(zap.NewNop())
			require.NoError(t, err)
			require.NotNil(t, d)

			d.duration = 1 * time.Second
			err = d.performDiscovery(tt.receivers, tt.observers)

			for _, observer := range tt.observers {
				mockExtension := observer.(*mockExtension)
				if mockExtension.started {
					assert.True(t, mockExtension.shutdown)
				} else {
					assert.False(t, mockExtension.shutdown)
				}
			}

			for _, receiver := range tt.receivers {
				mockReceiver := receiver.(*mockReceiverLogs)
				if mockReceiver.started {
					assert.True(t, mockReceiver.shutdown)
				} else {
					assert.False(t, mockReceiver.shutdown)
				}
			}

			if tt.expectedMsg != "" {
				assert.ErrorContains(t, err, tt.expectedMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockExtension struct {
	mockComponent
}

var _ otelcolextension.Extension = (*mockExtension)(nil)

type mockReceiverLogs struct {
	mockComponent
}

var _ otelcolreceiver.Logs = (*mockReceiverLogs)(nil)

type mockComponent struct {
	startErr    error
	shutdownErr error
	started     bool
	shutdown    bool
}

var _ component.Component = (*mockComponent)(nil)

func (m *mockComponent) Start(context.Context, component.Host) error {
	if m.startErr != nil {
		return m.startErr
	}
	m.started = true
	return nil
}

func (m *mockComponent) Shutdown(context.Context) error {
	m.shutdown = true
	if m.shutdownErr != nil {
		return m.shutdownErr
	}
	return nil
}

func TestDiscovererDurationFromEnv(t *testing.T) {
	t.Cleanup(func() func() {
		initial, ok := os.LookupEnv("SPLUNK_DISCOVERY_DURATION")
		os.Unsetenv("SPLUNK_DISCOVERY_DURATION")
		return func() {
			if ok {
				os.Setenv("SPLUNK_DISCOVERY_DURATION", initial)
			}
		}
	}())
	d, err := newDiscoverer(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Second, d.duration)

	os.Setenv("SPLUNK_DISCOVERY_DURATION", "10h")
	d, err = newDiscoverer(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Hour, d.duration)

	os.Setenv("SPLUNK_DISCOVERY_DURATION", "invalid")

	zc, observedLogs := observer.New(zap.DebugLevel)
	d, err = newDiscoverer(zap.New(zc))
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Second, d.duration)

	require.Eventually(t, func() bool {
		for _, m := range observedLogs.All() {
			if strings.Contains(m.Message, "Invalid SPLUNK_DISCOVERY_DURATION. Using default of 10s") {
				return m.ContextMap()["duration"] == "invalid"
			}
		}
		return false
	}, 2*time.Second, time.Millisecond)
}

func TestDetermineCurrentStatus(t *testing.T) {
	for _, test := range []struct {
		current, observed, expected discovery.StatusType
	}{
		{"failed", "failed", "failed"},
		{"failed", "partial", "partial"},
		{"failed", "successful", "successful"},
		{"partial", "failed", "partial"},
		{"partial", "partial", "partial"},
		{"partial", "successful", "successful"},
		{"successful", "failed", "successful"},
		{"successful", "partial", "successful"},
		{"successful", "successful", "successful"},
	} {
		t.Run(fmt.Sprintf("%s:%s->%s", test.current, test.observed, test.expected), func(t *testing.T) {
			require.Equal(t, test.expected, determineCurrentStatus(test.current, test.observed))
		})
	}
}
