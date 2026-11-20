// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/auth"
)

// unset marks an index attribute that is absent, as opposed to set to "".
const unset = "<unset>"

func tokenContext(ctx context.Context, token auth.HecTokenConfig) context.Context {
	return context.WithValue(ctx, auth.ContextKey, token)
}

func authorizedContext(ctx context.Context) context.Context {
	return tokenContext(ctx, auth.HecTokenConfig{AllowedIndexes: []string{"allowed"}})
}

var (
	defaultToken          = auth.HecTokenConfig{DefaultIndex: "dflt", AllowedIndexes: []string{"allowed", "dflt"}}
	disallowedDefault     = auth.HecTokenConfig{DefaultIndex: "dflt", AllowedIndexes: []string{"allowed"}}
	anyIndexToken         = auth.HecTokenConfig{}
	anyIndexDefaultToken  = auth.HecTokenConfig{DefaultIndex: "dflt"}
	notAllowedIndexErrMsg = `Permanent error: index "dflt" is not allowed`
)

func TestIndexAuthorizedLogs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, resource, record string
		authenticated, denied  bool
	}{
		{"missing authorization", "", "", false, true},
		{"no index", "", "", true, false},
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

func TestIndexAuthorizedLogsTokenIndexes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name             string
		resource, record string
		wantIndex        string
		wantErr          string
		token            auth.HecTokenConfig
	}{
		{name: "default index", token: defaultToken, resource: unset, record: unset, wantIndex: "dflt"},
		{name: "empty indexes use default", token: defaultToken, resource: "", record: "", wantIndex: "dflt"},
		{name: "empty record overrides resource", token: defaultToken, resource: "allowed", record: "", wantIndex: "dflt"},
		{name: "resource index wins over default", token: defaultToken, resource: "allowed", record: unset, wantIndex: "allowed"},
		{name: "record index wins over default", token: defaultToken, resource: unset, record: "allowed", wantIndex: "allowed"},
		{name: "default index must be allowed", token: disallowedDefault, resource: unset, record: unset, wantErr: notAllowedIndexErrMsg},
		{name: "no default index", token: auth.HecTokenConfig{AllowedIndexes: []string{"allowed"}}, resource: unset, record: unset, wantIndex: ""},
		{name: "empty allowlist allows any resource index", token: anyIndexToken, resource: "any", record: unset, wantIndex: "any"},
		{name: "empty allowlist allows any record index", token: anyIndexToken, resource: unset, record: "any", wantIndex: "any"},
		{name: "empty allowlist with default index", token: anyIndexDefaultToken, resource: unset, record: unset, wantIndex: "dflt"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ld := plog.NewLogs()
			rl := ld.ResourceLogs().AppendEmpty()
			if tt.resource != unset {
				rl.Resource().Attributes().PutStr(indexAttribute, tt.resource)
			}
			record := rl.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
			record.Body().SetStr("body")
			if tt.record != unset {
				record.Attributes().PutStr(indexAttribute, tt.record)
			}
			sink := &consumertest.LogsSink{}
			err := indexAuthorizedLogs{sink}.ConsumeLogs(tokenContext(t.Context(), tt.token), ld)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.True(t, consumererror.IsPermanent(err))
				assert.Empty(t, sink.AllLogs())
				return
			}
			require.NoError(t, err)
			require.Len(t, sink.AllLogs(), 1)
			got := sink.AllLogs()[0].ResourceLogs().At(0)
			event := translator.LogToSplunkEvent(got.Resource(), got.ScopeLogs().At(0).LogRecords().At(0),
				translator.DefaultHecToOtelAttrs(), translator.DefaultOtelToHecFields(), "", "", "")
			assert.Equal(t, tt.wantIndex, event.Index)
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
		{"no index", "", "", true, false},
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

func TestIndexAuthorizedTracesTokenIndexes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		resource  string
		wantIndex string
		wantErr   string
		token     auth.HecTokenConfig
	}{
		{name: "default index", token: defaultToken, resource: unset, wantIndex: "dflt"},
		{name: "empty index uses default", token: defaultToken, resource: "", wantIndex: "dflt"},
		{name: "resource index wins over default", token: defaultToken, resource: "allowed", wantIndex: "allowed"},
		{name: "default index must be allowed", token: disallowedDefault, resource: unset, wantErr: notAllowedIndexErrMsg},
		{name: "empty allowlist allows any index", token: anyIndexToken, resource: "any", wantIndex: "any"},
		{name: "empty allowlist with default index", token: anyIndexDefaultToken, resource: unset, wantIndex: "dflt"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			td := ptrace.NewTraces()
			rs := td.ResourceSpans().AppendEmpty()
			if tt.resource != unset {
				rs.Resource().Attributes().PutStr(indexAttribute, tt.resource)
			}
			rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
			sink := &consumertest.TracesSink{}
			err := indexAuthorizedTraces{sink}.ConsumeTraces(tokenContext(t.Context(), tt.token), td)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, sink.AllTraces())
				return
			}
			require.NoError(t, err)
			require.Len(t, sink.AllTraces(), 1)
			got := sink.AllTraces()[0].ResourceSpans().At(0)
			event := translator.SpanToSplunkEvent(got.Resource(), got.ScopeSpans().At(0).Spans().At(0), translator.DefaultHecToOtelAttrs(), "", "", "")
			assert.Equal(t, tt.wantIndex, event.Index)
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

func TestIndexAuthorizedMetricsTokenIndexes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		resource  string
		dataPoint string
		wantIndex string
		wantErr   string
		token     auth.HecTokenConfig
	}{
		{name: "default index", token: defaultToken, resource: unset, dataPoint: unset, wantIndex: "dflt"},
		{name: "empty index uses default", token: defaultToken, resource: "", dataPoint: unset, wantIndex: "dflt"},
		{name: "resource index wins over default", token: defaultToken, resource: "allowed", dataPoint: unset, wantIndex: "allowed"},
		{name: "default index must be allowed", token: disallowedDefault, resource: unset, dataPoint: unset, wantErr: notAllowedIndexErrMsg},
		{name: "empty allowlist allows any index", token: anyIndexToken, resource: "any", dataPoint: "other", wantIndex: "any"},
		{name: "empty allowlist with default index", token: anyIndexDefaultToken, resource: unset, dataPoint: unset, wantIndex: "dflt"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			md := pmetric.NewMetrics()
			rm := md.ResourceMetrics().AppendEmpty()
			if tt.resource != unset {
				rm.Resource().Attributes().PutStr(indexAttribute, tt.resource)
			}
			m := rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
			m.SetName("m")
			dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
			if tt.dataPoint != unset {
				dp.Attributes().PutStr(indexAttribute, tt.dataPoint)
			}
			sink := &consumertest.MetricsSink{}
			err := indexAuthorizedMetrics{sink}.ConsumeMetrics(tokenContext(t.Context(), tt.token), md)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, sink.AllMetrics())
				return
			}
			require.NoError(t, err)
			require.Len(t, sink.AllMetrics(), 1)
			got := sink.AllMetrics()[0].ResourceMetrics().At(0)
			events := translator.MetricToSplunkEvent(got.Resource(), got.ScopeMetrics().At(0).Metrics().At(0), zap.NewNop(), translator.DefaultHecToOtelAttrs(), "", "", "")
			require.Len(t, events, 1)
			assert.Equal(t, tt.wantIndex, events[0].Index)
		})
	}
}

func TestIndexAuthorizedConsumersMutateData(t *testing.T) {
	t.Parallel()
	for _, c := range []interface{ Capabilities() consumer.Capabilities }{
		indexAuthorizedLogs{consumertest.NewNop()},
		indexAuthorizedTraces{consumertest.NewNop()},
		indexAuthorizedMetrics{consumertest.NewNop()},
	} {
		assert.True(t, c.Capabilities().MutatesData)
	}
}
