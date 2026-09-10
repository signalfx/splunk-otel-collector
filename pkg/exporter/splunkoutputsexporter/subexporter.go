// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.uber.org/zap"

	"github.com/splunk/tarunner/pkg/splunkta/conf"
	"github.com/splunk/tarunner/pkg/splunkta/stanza"
	"github.com/splunk/tarunner/pkg/splunkta/tabuilder"
)

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

	outputs, err := tabuilder.ReadOutputGroups(splunkHome)
	if err != nil {
		return nil, fmt.Errorf("splunk_outputs: %w (base_dir: %s)", err, splunkHome)
	}

	exporters, err := o.createExporters(ctx, splunkHome, outputs, settings)
	if err != nil {
		return nil, fmt.Errorf("splunk_outputs: %w", err)
	}
	return packExporters(exporters), nil
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
