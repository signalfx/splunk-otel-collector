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

// These consumers apply the HEC token's index settings to the whole
// authenticated request before forwarding any data to the exporter: events
// without an index get the token's default index, and every index must be
// allowed by the token.
type indexAuthorizedLogs struct{ consumer.Logs }

func (indexAuthorizedLogs) Capabilities() consumer.Capabilities {
	// Setting the token's default index modifies the data; without this override the
	// embedded exporter's false value would be inherited and a fanout could share read-only data.
	return consumer.Capabilities{MutatesData: true}
}

func (c indexAuthorizedLogs) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	policy, err := tokenIndexPolicy(ctx)
	if err != nil {
		return err
	}
	for _, rl := range ld.ResourceLogs().All() {
		resourceIndex := indexOf(rl.Resource().Attributes())
		for _, sl := range rl.ScopeLogs().All() {
			for _, record := range sl.LogRecords().All() {
				attrs := record.Attributes()
				if index := indexOf(attrs); index != "" {
					if err := policy.check(index); err != nil {
						return err
					}
					continue
				}
				if err := policy.check(resourceIndex); err != nil {
					return err
				}
				// HEC log translation lets a record attribute override the resource
				// one even when empty, leaving the event without an index.
				if _, ok := attrs.Get(indexAttribute); ok || resourceIndex == "" {
					if err := policy.resolve(attrs, ""); err != nil {
						return err
					}
				}
			}
		}
	}
	return c.Logs.ConsumeLogs(ctx, ld)
}

type indexAuthorizedTraces struct{ consumer.Traces }

func (indexAuthorizedTraces) Capabilities() consumer.Capabilities {
	// Setting the token's default index modifies the data; without this override the
	// embedded exporter's false value would be inherited and a fanout could share read-only data.
	return consumer.Capabilities{MutatesData: true}
}

func (c indexAuthorizedTraces) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	policy, err := tokenIndexPolicy(ctx)
	if err != nil {
		return err
	}
	for _, rs := range td.ResourceSpans().All() {
		// HEC trace translation selects its destination index from the resource.
		attrs := rs.Resource().Attributes()
		if err := policy.resolve(attrs, indexOf(attrs)); err != nil {
			return err
		}
		for _, ss := range rs.ScopeSpans().All() {
			for _, span := range ss.Spans().All() {
				if err := policy.check(indexOf(span.Attributes())); err != nil {
					return err
				}
			}
		}
	}
	return c.Traces.ConsumeTraces(ctx, td)
}

type indexAuthorizedMetrics struct{ consumer.Metrics }

func (indexAuthorizedMetrics) Capabilities() consumer.Capabilities {
	// Setting the token's default index modifies the data; without this override the
	// embedded exporter's false value would be inherited and a fanout could share read-only data.
	return consumer.Capabilities{MutatesData: true}
}

func (c indexAuthorizedMetrics) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	policy, err := tokenIndexPolicy(ctx)
	if err != nil {
		return err
	}
	for _, rm := range md.ResourceMetrics().All() {
		// HEC metric translation selects its destination index from the resource.
		attrs := rm.Resource().Attributes()
		if err := policy.resolve(attrs, indexOf(attrs)); err != nil {
			return err
		}
		for _, sm := range rm.ScopeMetrics().All() {
			for _, metric := range sm.Metrics().All() {
				var err error
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					err = checkDataPointIndexes(metric.Gauge().DataPoints(), policy)
				case pmetric.MetricTypeSum:
					err = checkDataPointIndexes(metric.Sum().DataPoints(), policy)
				case pmetric.MetricTypeHistogram:
					err = checkDataPointIndexes(metric.Histogram().DataPoints(), policy)
				case pmetric.MetricTypeExponentialHistogram:
					err = checkDataPointIndexes(metric.ExponentialHistogram().DataPoints(), policy)
				case pmetric.MetricTypeSummary:
					err = checkDataPointIndexes(metric.Summary().DataPoints(), policy)
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

// indexOf returns the index the HEC translation reads from attrs, which
// treats a non-string value as empty.
func indexOf(attrs pcommon.Map) string {
	if v, ok := attrs.Get(indexAttribute); ok {
		return v.Str()
	}
	return ""
}

type indexPolicy struct {
	// allowed is nil when the token allows any index, matching HEC semantics
	// for an empty index list.
	allowed      map[string]struct{}
	defaultIndex string
}

func tokenIndexPolicy(ctx context.Context) (indexPolicy, error) {
	token, ok := ctx.Value(auth.ContextKey).(auth.HecTokenConfig)
	if !ok {
		return indexPolicy{}, consumererror.NewPermanent(errors.New("missing HEC token authorization"))
	}
	policy := indexPolicy{defaultIndex: token.DefaultIndex}
	if len(token.AllowedIndexes) > 0 {
		policy.allowed = make(map[string]struct{}, len(token.AllowedIndexes))
		for _, index := range token.AllowedIndexes {
			policy.allowed[index] = struct{}{}
		}
	}
	return policy, nil
}

// resolve sets the token's default index on attrs when index, the destination
// index of the event, is empty, and then checks the resulting index.
func (p indexPolicy) resolve(attrs pcommon.Map, index string) error {
	if index == "" && p.defaultIndex != "" {
		index = p.defaultIndex
		attrs.PutStr(indexAttribute, index)
	}
	return p.check(index)
}

func (p indexPolicy) check(index string) error {
	if index == "" || p.allowed == nil {
		return nil
	}
	if _, ok := p.allowed[index]; !ok {
		return consumererror.NewPermanent(fmt.Errorf("index %q is not allowed", index))
	}
	return nil
}

func checkDataPointIndexes[T interface{ Attributes() pcommon.Map }, S interface {
	Len() int
	At(int) T
}](points S, policy indexPolicy) error {
	for i := 0; i < points.Len(); i++ {
		if err := policy.check(indexOf(points.At(i).Attributes())); err != nil {
			return err
		}
	}
	return nil
}
