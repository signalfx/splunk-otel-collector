// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/exporter/splunkoutputsexporter"
)

func TestWithSubExporterOverridesBuiltInHTTPOut(t *testing.T) {
	baseDir := writeTA(t, "[httpout]\nuri = https://example.com/services/collector/event\nhttpEventCollectorToken = token\n")
	fake := &fakeSubExporterFactory{scheme: "httpout"}

	factory := splunkoutputsexporter.NewFactory(splunkoutputsexporter.WithSubExporter(fake))
	exp, err := factory.CreateLogs(context.Background(), newExporterSettings(), splunkoutputsexporter.Config{BaseDir: baseDir})
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Len(t, fake.requests, 1)
	require.Equal(t, baseDir, fake.requests[0].BaseDir)
	require.Empty(t, fake.requests[0].Path)
	require.Equal(t, "httpout", fake.requests[0].Output.Configuration.Stanza.Name)
	require.NotNil(t, fake.requests[0].Output.Configuration.Stanza.Params.Get("httpEventCollectorToken"))
	require.Equal(t, "token", fake.requests[0].Output.Configuration.Stanza.Params.Get("httpEventCollectorToken").Value)
	require.NotNil(t, fake.requests[0].Output.Configuration.Stanza.Params.Get("uri"))
	require.Equal(t, "https://example.com/services/collector/event", fake.requests[0].Output.Configuration.Stanza.Params.Get("uri").Value)
}

func TestWithSubExporterReadsRegisteredOutputSchemes(t *testing.T) {
	baseDir := writeTA(t, "[httpout]\nuri = https://example.com/services/collector/event\nhttpEventCollectorToken = token\n\n[tcpout:primary]\nserver = splunk:9997\n")
	httpout := &fakeSubExporterFactory{scheme: "httpout"}
	tcpout := &fakeSubExporterFactory{scheme: "tcpout"}

	factory := splunkoutputsexporter.NewFactory(
		splunkoutputsexporter.WithSubExporter(httpout),
		splunkoutputsexporter.WithSubExporter(tcpout),
	)
	exp, err := factory.CreateLogs(context.Background(), newExporterSettings(), splunkoutputsexporter.Config{BaseDir: baseDir})
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Len(t, httpout.requests, 1)
	require.Len(t, tcpout.requests, 1)
	require.Equal(t, "primary", tcpout.requests[0].Path)
	require.Equal(t, "tcpout:primary", tcpout.requests[0].Output.Configuration.Stanza.Name)
	require.NotNil(t, tcpout.requests[0].Output.Configuration.Stanza.Params.Get("server"))
	require.Equal(t, "splunk:9997", tcpout.requests[0].Output.Configuration.Stanza.Params.Get("server").Value)
}

func TestWithSubExporterRegistersCustomScheme(t *testing.T) {
	baseDir := writeTA(t, "[s2s://primary]\nserver = splunk:9997\n")
	fake := &fakeSubExporterFactory{scheme: "s2s"}

	factory := splunkoutputsexporter.NewFactory(splunkoutputsexporter.WithSubExporter(fake))
	exp, err := factory.CreateLogs(context.Background(), newExporterSettings(), splunkoutputsexporter.Config{BaseDir: baseDir})
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Len(t, fake.requests, 1)
	require.Equal(t, "primary", fake.requests[0].Path)
	require.Equal(t, "s2s://primary", fake.requests[0].Output.Configuration.Stanza.Name)
	require.NotNil(t, fake.requests[0].Output.Configuration.Stanza.Params.Get("server"))
	require.Equal(t, "splunk:9997", fake.requests[0].Output.Configuration.Stanza.Params.Get("server").Value)
}

func TestWithSubExporterDoesNotMatchCaseVariantScheme(t *testing.T) {
	baseDir := writeTA(t, "[S2S://primary]\nserver = splunk:9997\n")
	fake := &fakeSubExporterFactory{scheme: "s2s"}

	factory := splunkoutputsexporter.NewFactory(splunkoutputsexporter.WithSubExporter(fake))
	exp, err := factory.CreateLogs(context.Background(), newExporterSettings(), splunkoutputsexporter.Config{BaseDir: baseDir})
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Empty(t, fake.requests)
}

func TestWithSubExporterSkipsUnsupportedScheme(t *testing.T) {
	baseDir := writeTA(t, "[tcpout:primary]\nserver = splunk:9997\n")

	factory := splunkoutputsexporter.NewFactory()
	exp, err := factory.CreateLogs(context.Background(), newExporterSettings(), splunkoutputsexporter.Config{BaseDir: baseDir})
	require.NoError(t, err)
	require.NotNil(t, exp)
}

type fakeSubExporterFactory struct {
	scheme   string
	requests []splunkoutputsexporter.ExporterRequest
}

func (f *fakeSubExporterFactory) Scheme() string {
	return f.scheme
}

func (f *fakeSubExporterFactory) CreateLogs(_ context.Context, _ exporter.Settings, request splunkoutputsexporter.ExporterRequest) (exporter.Logs, error) {
	f.requests = append(f.requests, request)
	return fakeLogsExporter{}, nil
}

type fakeLogsExporter struct{}

func (fakeLogsExporter) Start(context.Context, component.Host) error {
	return nil
}

func (fakeLogsExporter) Shutdown(context.Context) error {
	return nil
}

func (fakeLogsExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (fakeLogsExporter) ConsumeLogs(context.Context, plog.Logs) error {
	return nil
}

func newExporterSettings() exporter.Settings {
	return exporter.Settings{
		ID:                component.MustNewID("splunk_outputs"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}
}

func writeTA(t *testing.T, outputsConf string) string {
	t.Helper()
	baseDir := t.TempDir()
	defaultDir := filepath.Join(baseDir, "etc", "system", "default")
	require.NoError(t, os.MkdirAll(defaultDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(defaultDir, "outputs.conf"), []byte(outputsConf), 0o600))
	return baseDir
}
