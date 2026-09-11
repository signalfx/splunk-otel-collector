// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/splunk/tarunner/pkg/splunkta/conf"
	"github.com/splunk/tarunner/pkg/splunkta/stanza"
	"github.com/splunk/tarunner/pkg/splunkta/tabuilder"
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
		return aggregateExporter{exporters: exporters}
	}
}

type (
	Output        = conf.Output
	Configuration = conf.Configuration
	Stanza        = conf.Stanza
	Params        = conf.Params
	Param         = conf.Param
)

// ExporterRequest is passed to a sub-exporter factory for one outputs.conf
// stanza. Path is the parsed target from the stanza name.
type ExporterRequest struct {
	BaseDir string
	Path    string
	Output  Output
}

// SubExporterFactory creates a logs exporter for one outputs.conf stanza kind.
//
// Scheme returns the output stanza kind to match. Kinds are matched
// case-sensitively, matching Splunk UF behavior.
type SubExporterFactory interface {
	Scheme() string
	CreateLogs(context.Context, exporter.Settings, ExporterRequest) (exporter.Logs, error)
}

// Option configures the splunk_outputs factory.
type Option func(*factoryOptions)

// WithSubExporter registers a sub-exporter factory by Scheme. If another
// factory is already registered for the same scheme, it is replaced.
func WithSubExporter(f SubExporterFactory) Option {
	return func(o *factoryOptions) {
		if f == nil {
			return
		}
		o.subExporters[f.Scheme()] = f
	}
}

type factoryOptions struct {
	subExporters map[string]SubExporterFactory
}

func newFactoryOptions(opts ...Option) factoryOptions {
	options := factoryOptions{
		subExporters: map[string]SubExporterFactory{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func (o factoryOptions) createLogsFunc(ctx context.Context, settings exporter.Settings, config component.Config) (exporter.Logs, error) {
	cfg := config.(Config)

	splunkHome, err := tabuilder.ResolveSplunkHome(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("splunk_outputs: %w", err)
	}

	return newSplunkOutputsExporter(ctx, splunkHome, o, settings), nil
}

func (o factoryOptions) createExporters(ctx context.Context, baseDir string, outputs []Output, settings exporter.Settings) ([]exporter.Logs, error) {
	var exporters []exporter.Logs
	for _, output := range outputs {
		e, err := o.createExporter(ctx, baseDir, output, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter %q: %w", output.Configuration.Stanza.Name, err)
		}
		if e == nil {
			settings.Logger.Info("splunk_outputs: skipping unsupported output stanza", zap.String("stanza", output.Configuration.Stanza.Name))
			continue
		}
		exporters = append(exporters, e)
	}
	return exporters, nil
}

func (o factoryOptions) createExporter(ctx context.Context, baseDir string, output Output, settings exporter.Settings) (exporter.Logs, error) {
	parsed, err := stanza.ParseOutputName(output.Configuration.Stanza.Name)
	if err != nil {
		return nil, err
	}
	if f, ok := o.subExporters[parsed.Kind]; ok {
		return f.CreateLogs(ctx, settings, ExporterRequest{
			BaseDir: baseDir,
			Path:    parsed.Target,
			Output:  output,
		})
	}
	return tabuilder.CreateOutputExporter(&output, settings.Logger, settings.TelemetrySettings)
}
