// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/signalfx/splunk-otel-collector/internal/auth"
)

func authorizedContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, auth.ContextKey, auth.HecTokenConfig{AllowedIndexes: []string{"allowed"}})
}

func TestIndexAuthorizedLogs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, resource, record string
		authenticated, denied  bool
	}{
		{"missing authorization", "", "", false, true},
		{"default index", "", "", true, false},
		{"resource index", "allowed", "", true, false},
		{"denied resource", "denied", "", true, true},
		{"denied record", "allowed", "denied", true, true},
		{"record override", "denied", "allowed", true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ld := plog.NewLogs()
			rl := ld.ResourceLogs().AppendEmpty()
			rl.Resource().Attributes().PutStr(indexAttribute, tt.resource)
			records := rl.ScopeLogs().AppendEmpty().LogRecords()
			records.AppendEmpty().Attributes().PutStr(indexAttribute, "allowed")
			records.AppendEmpty().Attributes().PutStr(indexAttribute, tt.record)
			ld.MarkReadOnly()
			sink := &consumertest.LogsSink{}
			c := indexAuthorizedLogs{sink}
			ctx := t.Context()
			if tt.authenticated {
				ctx = authorizedContext(ctx)
			}
			err := c.ConsumeLogs(ctx, ld)
			if tt.denied {
				require.Error(t, err)
				assert.True(t, consumererror.IsPermanent(err))
				assert.Empty(t, sink.AllLogs())
			} else {
				require.NoError(t, err)
				assert.Equal(t, 2, sink.LogRecordCount())
			}
		})
	}
}

func TestIndexAuthorizedTraces(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, resource, span  string
		authenticated, denied bool
	}{
		{"missing authorization", "", "", false, true},
		{"default index", "", "", true, false},
		{"resource index", "allowed", "", true, false},
		{"denied resource", "denied", "", true, true},
		{"denied span", "allowed", "denied", true, true},
		{"span cannot bypass resource", "denied", "allowed", true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			td := ptrace.NewTraces()
			rs := td.ResourceSpans().AppendEmpty()
			rs.Resource().Attributes().PutStr(indexAttribute, tt.resource)
			rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty().Attributes().PutStr(indexAttribute, tt.span)
			td.MarkReadOnly()
			sink := &consumertest.TracesSink{}
			c := indexAuthorizedTraces{sink}
			ctx := t.Context()
			if tt.authenticated {
				ctx = authorizedContext(ctx)
			}
			err := c.ConsumeTraces(ctx, td)
			if tt.denied {
				require.Error(t, err)
				assert.True(t, consumererror.IsPermanent(err))
				assert.Empty(t, sink.AllTraces())
			} else {
				require.NoError(t, err)
				assert.Equal(t, 1, sink.SpanCount())
			}
		})
	}
}

func TestIndexAuthorizedMetrics(t *testing.T) {
	t.Parallel()
	for _, kind := range []pmetric.MetricType{pmetric.MetricTypeGauge, pmetric.MetricTypeSum, pmetric.MetricTypeHistogram, pmetric.MetricTypeExponentialHistogram, pmetric.MetricTypeSummary} {
		for _, index := range []string{"allowed", "denied"} {
			t.Run(kind.String()+"/"+index, func(t *testing.T) {
				t.Parallel()
				md := pmetric.NewMetrics()
				m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
				switch kind {
				case pmetric.MetricTypeGauge:
					m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr(indexAttribute, index)
				case pmetric.MetricTypeSum:
					m.SetEmptySum().DataPoints().AppendEmpty().Attributes().PutStr(indexAttribute, index)
				case pmetric.MetricTypeHistogram:
					m.SetEmptyHistogram().DataPoints().AppendEmpty().Attributes().PutStr(indexAttribute, index)
				case pmetric.MetricTypeExponentialHistogram:
					m.SetEmptyExponentialHistogram().DataPoints().AppendEmpty().Attributes().PutStr(indexAttribute, index)
				case pmetric.MetricTypeSummary:
					m.SetEmptySummary().DataPoints().AppendEmpty().Attributes().PutStr(indexAttribute, index)
				}
				md.MarkReadOnly()
				sink := &consumertest.MetricsSink{}
				c := indexAuthorizedMetrics{sink}
				err := c.ConsumeMetrics(authorizedContext(t.Context()), md)
				if index == "denied" {
					require.EqualError(t, err, `Permanent error: index "denied" is not allowed`)
					assert.True(t, consumererror.IsPermanent(err))
					assert.Empty(t, sink.AllMetrics())
				} else {
					require.NoError(t, err)
					assert.Equal(t, 1, sink.DataPointCount())
				}
			})
		}
	}
	md := pmetric.NewMetrics()
	md.ResourceMetrics().AppendEmpty().Resource().Attributes().PutStr(indexAttribute, "denied")
	c := indexAuthorizedMetrics{&consumertest.MetricsSink{}}
	require.EqualError(t, c.ConsumeMetrics(authorizedContext(t.Context()), md), `Permanent error: index "denied" is not allowed`)
	require.EqualError(t, c.ConsumeMetrics(t.Context(), md), "Permanent error: missing HEC token authorization")
}
