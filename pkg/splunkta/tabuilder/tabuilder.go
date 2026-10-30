// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package tabuilder dispatches an inputs.conf stanza to the logs receiver that
// handles its input kind.
package tabuilder

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/stanza"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/batchreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/monitorreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/scriptreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/tcpreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/udpreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/wineventlogreceiver"
)

// CreateReceiver builds a single logs receiver for an input stanza.
// Returns nil receiver for unsupported input kinds; the caller is responsible for logging.
func CreateReceiver(ctx context.Context, baseDir string, next consumer.Logs, input conf.Input, transforms []conf.Transform, props []conf.Prop, telemetrySettings component.TelemetrySettings) (receiver.Logs, error) {
	parsed, err := stanza.ParseName(input.Configuration.Stanza.Name)
	if err != nil {
		return nil, err
	}
	switch parsed.Kind {
	case "script":
		f := scriptreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), &scriptreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "batch":
		f := batchreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), batchreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "monitor":
		f := monitorreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), monitorreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	// inputs.conf.spec spells this [WinEventLog://<name>].
	case "WinEventLog":
		f := wineventlogreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), wineventlogreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "tcp":
		f := tcpreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), tcpreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "udp":
		f := udpreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), udpreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	default:
		return nil, nil
	}
}

func settings(f receiver.Factory, path string, telemetrySettings component.TelemetrySettings) receiver.Settings {
	return receiver.Settings{
		ID:                component.MustNewIDWithName(f.Type().String(), path),
		TelemetrySettings: telemetrySettings,
	}
}
