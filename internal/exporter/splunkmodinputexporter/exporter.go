// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type modinputExporter struct {
	enc    *json.Encoder
	logger *zap.Logger
	config Config
}

// newModinputExporter takes an *os.File instead of an io.Writer on purpose:
// os.File.Write holds the file's write lock for the whole call, so each event
// (written with a single Write) cannot interleave with events written
// concurrently by other signals or exporter instances sharing stdout. An
// arbitrary io.Writer gives no such guarantee.
func newModinputExporter(cfg *Config, set component.TelemetrySettings, out *os.File) *modinputExporter {
	return &modinputExporter{enc: json.NewEncoder(fileWriter{out}), logger: set.Logger, config: *cfg}
}

// fileWriter tags errors from the underlying file so writeEvent can tell them
// apart from marshal errors, since Encoder.Encode returns both unwrapped.
type fileWriter struct{ f *os.File }

type writeError struct{ err error }

func (e *writeError) Error() string { return e.err.Error() }
func (e *writeError) Unwrap() error { return e.err }

func (w fileWriter) Write(b []byte) (int, error) {
	n, err := w.f.Write(b)
	if err != nil {
		err = &writeError{err}
	}
	return n, err
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

// writeEvent encodes the event and its trailing newline into a single Write
// call, avoiding the extra copy that appending the newline to a marshaled
// slice would require.
func (se *modinputExporter) writeEvent(event *translator.Event) error {
	if err := se.enc.Encode(event); err != nil {
		var we *writeError
		if errors.As(err, &we) {
			return fmt.Errorf("write HEC event to stdout: %w", we.err)
		}
		return consumererror.NewPermanent(fmt.Errorf("marshal HEC event: %w", err))
	}
	return nil
}
