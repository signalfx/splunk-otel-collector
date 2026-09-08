// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
)

var nopInstance = &nopExporter{}

type nopExporter struct {
	component.StartFunc
	component.ShutdownFunc
}

func (nopExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (nopExporter) ConsumeLogs(context.Context, plog.Logs) error {
	return nil
}

type aggregateExporter struct {
	exporters []exporter.Logs
}

func (a aggregateExporter) Start(ctx context.Context, host component.Host) error {
	var errs []error
	for _, e := range a.exporters {
		errs = append(errs, e.Start(ctx, host))
	}
	return errors.Join(errs...)
}

func (a aggregateExporter) Shutdown(ctx context.Context) error {
	var errs []error
	for _, e := range a.exporters {
		errs = append(errs, e.Shutdown(ctx))
	}
	return errors.Join(errs...)
}

func (a aggregateExporter) Capabilities() consumer.Capabilities {
	var capabilities consumer.Capabilities
	for _, e := range a.exporters {
		if e.Capabilities().MutatesData {
			capabilities.MutatesData = true
			break
		}
	}
	return capabilities
}

func (a aggregateExporter) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	var errs []error
	for _, e := range a.exporters {
		errs = append(errs, e.ConsumeLogs(ctx, logs))
	}
	return errors.Join(errs...)
}

func packExporters(exporters []exporter.Logs) exporter.Logs {
	switch len(exporters) {
	case 0:
		return nopInstance
	case 1:
		return exporters[0]
	default:
		return aggregateExporter{
			exporters: exporters,
		}
	}
}
