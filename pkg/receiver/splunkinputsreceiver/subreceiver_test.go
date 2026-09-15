// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// failingReceiver is a mock receiver whose Start always returns an error.
type failingReceiver struct{}

func (failingReceiver) Start(context.Context, component.Host) error {
	return errors.New("start failed")
}

func (failingReceiver) Shutdown(context.Context) error { return nil }

// failingSubReceiverFactory always returns a failingReceiver.
type failingSubReceiverFactory struct{}

func (f *failingSubReceiverFactory) Scheme() string { return "monitor" }

func (f *failingSubReceiverFactory) CreateLogs(_ context.Context, _ receiver.Settings, _ ReceiverRequest, _ consumer.Logs) (receiver.Logs, error) {
	return failingReceiver{}, nil
}

func TestStartReceiversContinuesOnStartFailure(t *testing.T) {
	splunkHome := t.TempDir()
	taDir := makeTA(t, splunkHome, "splunk_ta_syslog")

	options := newFactoryOptions(WithSubReceiver(&failingSubReceiverFactory{}))
	settings := receiver.Settings{
		ID:                component.MustNewID("splunk_inputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}

	started, err := options.startReceivers(context.Background(), nil, splunkHome, taDir, testNopConsumer{}, settings)

	// Start failed but error is not propagated — receiver is skipped
	require.NoError(t, err)
	assert.Empty(t, started)
}
