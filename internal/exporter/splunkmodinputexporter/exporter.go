// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/goccy/go-json"
	translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// All signals and component instances share stdout. Serialize complete JSON lines
// so a large event cannot interleave with another exporter's event.
var stdoutMu sync.Mutex

type modinputExporter struct {
	writer io.Writer
	logger *zap.Logger
	config Config
}

func newModinputExporter(cfg *Config, set component.TelemetrySettings, writer io.Writer) *modinputExporter {
	return &modinputExporter{writer: writer, logger: set.Logger, config: *cfg}
}

func (se *modinputExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()
	toHecAttrs := translator.DefaultOtelToHecFields()
	for _, rl := range ld.ResourceLogs().All() {
		for _, sl := range rl.ScopeLogs().All() {
			for _, record := range sl.LogRecords().All() {
				if err := ctx.Err(); err != nil {
					return err
				}
				event := translator.LogToSplunkEvent(rl.Resource(), record, toOtelAttrs, toHecAttrs, se.config.Source, se.config.Sourcetype, se.config.Index)
				if event != nil {
					if err := se.writeEvent(event); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (se *modinputExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()
	for _, rs := range td.ResourceSpans().All() {
		for _, ss := range rs.ScopeSpans().All() {
			for _, span := range ss.Spans().All() {
				if err := ctx.Err(); err != nil {
					return err
				}
				event := translator.SpanToSplunkEvent(rs.Resource(), span, toOtelAttrs, se.config.Source, se.config.Sourcetype, se.config.Index)
				if err := se.writeEvent(event); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (se *modinputExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()
	for _, rm := range md.ResourceMetrics().All() {
		for _, sm := range rm.ScopeMetrics().All() {
			for _, metric := range sm.Metrics().All() {
				if err := ctx.Err(); err != nil {
					return err
				}
				for _, event := range translator.MetricToSplunkEvent(rm.Resource(), metric, se.logger, toOtelAttrs, se.config.Source, se.config.Sourcetype, se.config.Index) {
					if err := ctx.Err(); err != nil {
						return err
					}
					if err := se.writeEvent(event); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (se *modinputExporter) writeEvent(event *translator.Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return consumererror.NewPermanent(fmt.Errorf("marshal HEC event: %w", err))
	}
	b = append(b, '\n')
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	n, err := se.writer.Write(b)
	if err != nil {
		return fmt.Errorf("write HEC event to stdout: %w", err)
	}
	if n != len(b) {
		return fmt.Errorf("write HEC event to stdout: %w", io.ErrShortWrite)
	}
	return nil
}
