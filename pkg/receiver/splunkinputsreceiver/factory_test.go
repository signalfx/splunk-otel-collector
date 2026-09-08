// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/receiver/splunkinputsreceiver"
)

func TestWithSubReceiverRegistersCustomScheme(t *testing.T) {
	splunkHome := writeTA(t, "[custom:///thing]\nsourcetype = custom\n")
	fake := &fakeSubReceiverFactory{scheme: "custom"}

	factory := splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake))
	rcvr := createAndStart(t, factory, splunkHome)
	require.NotNil(t, rcvr)
	require.True(t, fake.called)
	require.Equal(t, filepath.Join(splunkHome, "etc", "apps", "Splunk_TA_test"), fake.request.BaseDir)
	require.Equal(t, "/thing", fake.request.Path)
	require.Equal(t, "custom:///thing", fake.request.Input.Configuration.Stanza.Name)
}

func TestWithSubReceiverOverridesBuiltInAndHandlesEmptySchemeAsScript(t *testing.T) {
	splunkHome := writeTA(t, "[modinput]\ninterval = -1\n")
	fake := &fakeSubReceiverFactory{scheme: "script"}

	factory := splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake))
	rcvr := createAndStart(t, factory, splunkHome)
	require.NotNil(t, rcvr)
	require.True(t, fake.called)
	require.Equal(t, filepath.Join(splunkHome, "etc", "apps", "Splunk_TA_test"), fake.request.BaseDir)
	require.Equal(t, "modinput", fake.request.Input.Configuration.Stanza.Name)
}

func TestWithSubReceiverDoesNotMatchCaseVariantScheme(t *testing.T) {
	splunkHome := writeTA(t, "[Custom://thing]\n")
	fake := &fakeSubReceiverFactory{scheme: "custom"}

	factory := splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake))
	rcvr, err := factory.CreateLogs(context.Background(), newReceiverSettings(), splunkinputsreceiver.Config{BaseDir: splunkHome}, nopConsumer{})
	require.NoError(t, err)
	require.NotNil(t, rcvr)
	err = rcvr.Start(context.Background(), nil)
	require.ErrorContains(t, err, `unsupported scheme "Custom"`)
	require.False(t, fake.called)
}

func TestWithSubReceiverRejectsUnsupportedScheme(t *testing.T) {
	splunkHome := writeTA(t, "[unsupported://thing]\n")

	factory := splunkinputsreceiver.NewFactory()
	rcvr, err := factory.CreateLogs(context.Background(), newReceiverSettings(), splunkinputsreceiver.Config{BaseDir: splunkHome}, nopConsumer{})
	require.NoError(t, err)
	require.NotNil(t, rcvr)
	err = rcvr.Start(context.Background(), nil)
	require.ErrorContains(t, err, `unsupported scheme "unsupported"`)
}

func TestWithSubReceiverSkipsDisabledCustomStanza(t *testing.T) {
	splunkHome := writeTA(t, "[custom:///thing]\ndisabled = 1\n")
	fake := &fakeSubReceiverFactory{scheme: "custom"}

	factory := splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake))
	rcvr := createAndStart(t, factory, splunkHome)
	require.NotNil(t, rcvr)
	require.False(t, fake.called)
}

func TestSystemAndTAStanzasBothFire(t *testing.T) {
	splunkHome := t.TempDir()

	// TA stanza
	taDefaultDir := filepath.Join(splunkHome, "etc", "apps", "Splunk_TA_test", "default")
	require.NoError(t, os.MkdirAll(taDefaultDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(taDefaultDir, "inputs.conf"),
		[]byte("[custom:///ta-thing]\nsourcetype = ta\n"), 0o600))

	// system-only stanza (not in any TA)
	systemLocalDir := filepath.Join(splunkHome, "etc", "system", "local")
	require.NoError(t, os.MkdirAll(systemLocalDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(systemLocalDir, "inputs.conf"),
		[]byte("[custom:///system-thing]\nsourcetype = system\n"), 0o600))

	fake := &fakeSubReceiverFactory{scheme: "custom"}
	rcvr := createAndStart(t, splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake)), splunkHome)
	require.NotNil(t, rcvr)
	require.Equal(t, 2, fake.callCount)
}

func TestWithSubReceiverRequestIncludesPropsAndTransforms(t *testing.T) {
	splunkHome := writeTA(t, "[custom:///thing]\nsourcetype = custom\n")
	taDefaultDir := filepath.Join(splunkHome, "etc", "apps", "Splunk_TA_test", "default")
	require.NoError(t, os.WriteFile(filepath.Join(taDefaultDir, "props.conf"), []byte("[custom]\nTRANSFORMS-routing = route\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(taDefaultDir, "transforms.conf"), []byte("[route]\nREGEX = ^(.*)$\nFORMAT = $1\n"), 0o600))
	fake := &fakeSubReceiverFactory{scheme: "custom"}

	factory := splunkinputsreceiver.NewFactory(splunkinputsreceiver.WithSubReceiver(fake))
	rcvr := createAndStart(t, factory, splunkHome)
	require.NotNil(t, rcvr)
	require.True(t, fake.called)
	require.Len(t, fake.request.Props, 1)
	require.Equal(t, "custom", fake.request.Props[0].Name)
	require.Len(t, fake.request.Props[0].Transforms, 1)
	require.Equal(t, "route", fake.request.Props[0].Transforms[0].Stanza[0])
	require.Len(t, fake.request.Transforms, 1)
	require.Equal(t, "route", fake.request.Transforms[0].Name)
	require.Equal(t, "^(.*)$", fake.request.Transforms[0].Regex)
}

// createAndStart creates a receiver via the factory and calls Start, failing the test on any error.
func createAndStart(t *testing.T, factory receiver.Factory, splunkHome string) receiver.Logs {
	t.Helper()
	rcvr, err := factory.CreateLogs(context.Background(), newReceiverSettings(), splunkinputsreceiver.Config{BaseDir: splunkHome}, nopConsumer{})
	require.NoError(t, err)
	require.NoError(t, rcvr.Start(context.Background(), nil))
	t.Cleanup(func() { _ = rcvr.Shutdown(context.Background()) })
	return rcvr
}

type fakeSubReceiverFactory struct {
	scheme    string
	request   splunkinputsreceiver.ReceiverRequest
	callCount int
	called    bool
}

func (f *fakeSubReceiverFactory) Scheme() string {
	return f.scheme
}

func (f *fakeSubReceiverFactory) CreateLogs(_ context.Context, _ receiver.Settings, request splunkinputsreceiver.ReceiverRequest, _ consumer.Logs) (receiver.Logs, error) {
	f.called = true
	f.callCount++
	f.request = request
	return fakeReceiver{}, nil
}

type fakeReceiver struct{}

func (fakeReceiver) Start(context.Context, component.Host) error {
	return nil
}

func (fakeReceiver) Shutdown(context.Context) error {
	return nil
}

type nopConsumer struct{}

func (nopConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (nopConsumer) ConsumeLogs(context.Context, plog.Logs) error {
	return nil
}

func newReceiverSettings() receiver.Settings {
	return receiver.Settings{
		ID:                component.MustNewID("splunk_inputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
}

// writeTA creates a minimal Splunk home layout with one TA and returns the
// splunk home path. The TA is placed at <splunkHome>/etc/apps/ta/.
func writeTA(t *testing.T, inputsConf string) string {
	t.Helper()
	splunkHome := t.TempDir()
	taDefaultDir := filepath.Join(splunkHome, "etc", "apps", "Splunk_TA_test", "default")
	require.NoError(t, os.MkdirAll(taDefaultDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(taDefaultDir, "inputs.conf"), []byte(inputsConf), 0o600))
	return splunkHome
}
