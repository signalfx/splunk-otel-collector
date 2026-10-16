// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/signalfx/splunk-otel-collector/internal/auth"
)

// These consumers validate the whole authenticated request against the HEC
// token's allowed indexes before forwarding any data to the exporter.
type indexAuthorizedLogs struct{ consumer.Logs }

func (c indexAuthorizedLogs) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	validate, err := indexValidator(ctx)
	if err != nil {
		return err
	}
	for _, rl := range ld.ResourceLogs().All() {
		for _, sl := range rl.ScopeLogs().All() {
			for _, record := range sl.LogRecords().All() {
				attrs := record.Attributes()
				if index, ok := attrs.Get(indexAttribute); !ok || index.AsString() == "" {
					attrs = rl.Resource().Attributes()
				}
				if err := validate(attrs); err != nil {
					return err
				}
			}
		}
	}
	return c.Logs.ConsumeLogs(ctx, ld)
}

type indexAuthorizedTraces struct{ consumer.Traces }

func (c indexAuthorizedTraces) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	validate, err := indexValidator(ctx)
	if err != nil {
		return err
	}
	for _, rs := range td.ResourceSpans().All() {
		// HEC trace translation selects its destination index from the resource.
		if err := validate(rs.Resource().Attributes()); err != nil {
			return err
		}
		for _, ss := range rs.ScopeSpans().All() {
			for _, span := range ss.Spans().All() {
				if err := validate(span.Attributes()); err != nil {
					return err
				}
			}
		}
	}
	return c.Traces.ConsumeTraces(ctx, td)
}

type indexAuthorizedMetrics struct{ consumer.Metrics }

func (c indexAuthorizedMetrics) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	validate, err := indexValidator(ctx)
	if err != nil {
		return err
	}
	for _, rm := range md.ResourceMetrics().All() {
		if err := validate(rm.Resource().Attributes()); err != nil {
			return err
		}
		for _, sm := range rm.ScopeMetrics().All() {
			for _, metric := range sm.Metrics().All() {
				var err error
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					err = validateDataPointIndexes(metric.Gauge().DataPoints(), validate)
				case pmetric.MetricTypeSum:
					err = validateDataPointIndexes(metric.Sum().DataPoints(), validate)
				case pmetric.MetricTypeHistogram:
					err = validateDataPointIndexes(metric.Histogram().DataPoints(), validate)
				case pmetric.MetricTypeExponentialHistogram:
					err = validateDataPointIndexes(metric.ExponentialHistogram().DataPoints(), validate)
				case pmetric.MetricTypeSummary:
					err = validateDataPointIndexes(metric.Summary().DataPoints(), validate)
				case pmetric.MetricTypeEmpty:
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return c.Metrics.ConsumeMetrics(ctx, md)
}

const indexAttribute = "com.splunk.index"

func indexValidator(ctx context.Context) (func(pcommon.Map) error, error) {
	token, ok := ctx.Value(auth.ContextKey).(auth.HecTokenConfig)
	if !ok {
		return nil, consumererror.NewPermanent(errors.New("missing HEC token authorization"))
	}
	allowed := make(map[string]struct{}, len(token.AllowedIndexes))
	for _, index := range token.AllowedIndexes {
		allowed[index] = struct{}{}
	}
	return func(attrs pcommon.Map) error {
		if index, ok := attrs.Get(indexAttribute); ok && index.AsString() != "" {
			if _, ok := allowed[index.AsString()]; !ok {
				return consumererror.NewPermanent(fmt.Errorf("index %q is not allowed", index.AsString()))
			}
		}
		return nil
	}, nil
}

func validateDataPointIndexes[T interface{ Attributes() pcommon.Map }, S interface {
	Len() int
	At(int) T
}](points S, validate func(pcommon.Map) error) error {
	for i := 0; i < points.Len(); i++ {
		if err := validate(points.At(i).Attributes()); err != nil {
			return err
		}
	}
	return nil
}
