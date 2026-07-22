// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rollingspanlatencyprocessor

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
)

func newTestProcessor(t *testing.T, cfg Config) (*rollingSpanLatencyProcessor, *consumertest.TracesSink) {
	t.Helper()
	sink := new(consumertest.TracesSink)
	telemetry := component.TelemetrySettings{
		Logger:        zap.NewNop(),
		MeterProvider: noop.NewMeterProvider(),
	}
	p, err := newProcessor(cfg, telemetry, sink)
	if err != nil {
		t.Fatalf("newProcessor: %v", err)
	}
	return p, sink
}

// makeTraces builds a single-span ptrace.Traces. resAttrs are written as
// resource attributes; spanName and durationNs describe the span. now is
// used as the span's EndTimestamp so the EWMA clock advances correctly across
// successive calls.
func makeTraces(resAttrs map[string]string, spanName string, durationNs int64, now time.Time) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	for k, v := range resAttrs {
		rs.Resource().Attributes().PutStr(k, v)
	}
	ss := rs.ScopeSpans().AppendEmpty()
	sp := ss.Spans().AppendEmpty()
	sp.SetName(spanName)
	endNs := now.UnixNano()
	sp.SetStartTimestamp(pcommon.Timestamp(endNs - durationNs))
	sp.SetEndTimestamp(pcommon.Timestamp(endNs))
	return td
}

// defaultResAttrs returns resource attrs that satisfy all three default key
// attributes so tests don't need to repeat the map literal.
func defaultResAttrs(namespace, service, env string) map[string]string {
	return map[string]string{
		"service.namespace":           namespace,
		"service.name":                service,
		"deployment.environment.name": env,
	}
}

// keyFor returns the stats-map key the processor would derive for the given
// resource attrs and span name under cfg.
func keyFor(cfg Config, resAttrs map[string]string, spanName string) string {
	vals := make([]string, len(cfg.ResourceKeyAttributes))
	for i, attrKey := range cfg.ResourceKeyAttributes {
		vals[i] = resAttrs[attrKey]
	}
	return buildKey(vals, spanName)
}

// warmProcessor feeds count spans of durationNs into p, advancing the span
// end timestamps by stepDur each time. Returns the final span timestamp.
func warmProcessor(p *rollingSpanLatencyProcessor, resAttrs map[string]string, spanName string, durationNs int64, count int, stepDur time.Duration) time.Time {
	now := time.Unix(1_000_000, 0) // well past epoch so timestamps are valid
	for i := 0; i < count; i++ {
		now = now.Add(stepDur)
		_ = p.ConsumeTraces(context.Background(), makeTraces(resAttrs, spanName, durationNs, now))
	}
	return now
}

// collectLabels gathers all latency.category attribute values seen in the sink.
func collectLabels(sink *consumertest.TracesSink, attrKey string) []string {
	var labels []string
	for _, td := range sink.AllTraces() {
		rss := td.ResourceSpans()
		for i := 0; i < rss.Len(); i++ {
			sss := rss.At(i).ScopeSpans()
			for j := 0; j < sss.Len(); j++ {
				spans := sss.At(j).Spans()
				for k := 0; k < spans.Len(); k++ {
					if v, ok := spans.At(k).Attributes().Get(attrKey); ok {
						labels = append(labels, v.Str())
					}
				}
			}
		}
	}
	return labels
}

var baseAttrs = defaultResAttrs("ns", "svc", "prod")

func TestProcessor_NoLabelBelowWarmup(t *testing.T) {
	cfg := defaultConfig()
	p, sink := newTestProcessor(t, cfg)

	warmProcessor(p, baseAttrs, "op", int64(100e6), 5, time.Second)

	if labels := collectLabels(sink, cfg.AttributeKey); len(labels) > 0 {
		t.Errorf("expected no labels before warmup, got %v", labels)
	}
}

func TestProcessor_NormalSpanNotLabeled(t *testing.T) {
	cfg := defaultConfig()
	p, sink := newTestProcessor(t, cfg)

	now := warmProcessor(p, baseAttrs, "op", int64(100e6), 50, time.Second)
	sink.Reset()

	now = now.Add(time.Second)
	// 102ms on a 100ms baseline: (102-100)/1ms_floor = 2σ, below slow_threshold=3.
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", 102e6, now))

	if labels := collectLabels(sink, cfg.AttributeKey); len(labels) > 0 {
		t.Errorf("normal span should not be labeled, got %v", labels)
	}
}

func TestProcessor_SlowSpanLabeled(t *testing.T) {
	cfg := defaultConfig()
	p, sink := newTestProcessor(t, cfg)

	// Tight distribution: 100ms ± 1ms alternating → small stddev.
	now := time.Unix(1_000_000, 0)
	for i := 0; i < 100; i++ {
		now = now.Add(time.Second)
		dur := int64(99e6)
		if i%2 == 0 {
			dur = int64(101e6)
		}
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", dur, now))
	}
	sink.Reset()

	now = now.Add(time.Second)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(200e6), now))

	labels := collectLabels(sink, cfg.AttributeKey)
	if len(labels) == 0 {
		t.Error("expected slow or very_slow label on span far above tight baseline")
	}
	for _, l := range labels {
		if l != attributeValueSlow && l != attributeValueVerySlow {
			t.Errorf("unexpected label value %q", l)
		}
	}
}

func TestProcessor_VerySlowSpanLabeled(t *testing.T) {
	cfg := defaultConfig()
	p, sink := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 100; i++ {
		now = now.Add(time.Second)
		dur := int64(99e6)
		if i%2 == 0 {
			dur = int64(101e6)
		}
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", dur, now))
	}
	sink.Reset()

	now = now.Add(time.Second)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(500e6), now))

	labels := collectLabels(sink, cfg.AttributeKey)
	found := false
	for _, l := range labels {
		if l == attributeValueVerySlow {
			found = true
		}
	}
	if !found {
		t.Errorf("expected very_slow label, got %v", labels)
	}
}

// TestProcessor_DemoScenario_NormalSpanNotMislabeled simulates the demo
// traffic pattern: warmup on jittered normal spans, then repeated rounds of
// slow + very-slow + normal traffic. Verifies that normal-latency spans are
// never labeled and that slow/very-slow spans are labeled correctly even after
// outliers feed back into the baseline.
func TestProcessor_DemoScenario_NormalSpanNotMislabeled(t *testing.T) {
	cfg := defaultConfig()
	cfg.HalfLife = 2 * time.Second
	p, sink := newTestProcessor(t, cfg)

	// Warmup: 35 iterations × 2 spans, jittered ±2ms. Exceeds warmup_count=30.
	now := time.Unix(1_000_000, 0)
	jitters := []int64{-2, -1, 0, 1, 2, -2, 0, 1, -1, 2}
	for i := 0; i < 35; i++ {
		now = now.Add(200 * time.Millisecond)
		jitter := jitters[i%len(jitters)]
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "process-order", (50+jitter)*int64(time.Millisecond), now))
		now = now.Add(200 * time.Millisecond)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "query-inventory", (50+jitter)*int64(time.Millisecond), now))
	}
	sink.Reset()

	slowLabeled := 0
	verySlowLabeled := 0

	// Run 5 rounds of: slow(67ms) + very-slow(80ms) + 5×normal pairs.
	for round := 0; round < 5; round++ {
		now = now.Add(200 * time.Millisecond)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "process-order", 67*int64(time.Millisecond), now))

		now = now.Add(200 * time.Millisecond)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "process-order", 80*int64(time.Millisecond), now))

		// Normal spans — must NOT be labeled.
		for i := 0; i < 5; i++ {
			now = now.Add(200 * time.Millisecond)
			jitter := jitters[i%len(jitters)]
			_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "process-order", (50+jitter)*int64(time.Millisecond), now))
			now = now.Add(200 * time.Millisecond)
			_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "query-inventory", (50+jitter)*int64(time.Millisecond), now))
		}
	}

	for _, td := range sink.AllTraces() {
		rss := td.ResourceSpans()
		for i := 0; i < rss.Len(); i++ {
			sss := rss.At(i).ScopeSpans()
			for j := 0; j < sss.Len(); j++ {
				spans := sss.At(j).Spans()
				for k := 0; k < spans.Len(); k++ {
					sp := spans.At(k)
					dur := int64(sp.EndTimestamp()-sp.StartTimestamp()) / int64(time.Millisecond)
					v, hasLabel := sp.Attributes().Get(cfg.AttributeKey)

					if sp.Name() == "query-inventory" && hasLabel {
						t.Errorf("query-inventory span should never be labeled, got %q", v.Str())
					}
					if sp.Name() == "process-order" {
						if dur <= 55 && hasLabel {
							t.Errorf("normal process-order span (%dms) should not be labeled, got %q", dur, v.Str())
						}
						if hasLabel {
							switch v.Str() {
							case attributeValueSlow:
								slowLabeled++
							case attributeValueVerySlow:
								verySlowLabeled++
							}
						}
					}
				}
			}
		}
	}

	if slowLabeled == 0 {
		t.Error("expected at least one slow label on 67ms spans")
	}
	if verySlowLabeled == 0 {
		t.Error("expected at least one very_slow label on 80ms spans")
	}
}

func TestProcessor_IndependentBaselinePerSpanName(t *testing.T) {
	cfg := defaultConfig()
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 50; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "fast-op", int64(10e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "slow-op", int64(500e6), now))
	}

	fastMean, _, _ := p.getOrCreateStats(keyFor(cfg, baseAttrs, "fast-op")).snapshot()
	slowMean, _, _ := p.getOrCreateStats(keyFor(cfg, baseAttrs, "slow-op")).snapshot()

	if fastMean >= slowMean {
		t.Errorf("fast-op mean (%.2fms) should be less than slow-op mean (%.2fms)", fastMean/1e6, slowMean/1e6)
	}
}

func TestProcessor_SameSpanNameDifferentServicesHaveIndependentBaselines(t *testing.T) {
	cfg := defaultConfig()
	p, _ := newTestProcessor(t, cfg)

	attrsA := defaultResAttrs("ns", "service-a", "prod")
	attrsB := defaultResAttrs("ns", "service-b", "prod")

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 60; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsA, "POST /items", int64(10e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsB, "POST /items", int64(500e6), now))
	}

	keyA := keyFor(cfg, attrsA, "POST /items")
	keyB := keyFor(cfg, attrsB, "POST /items")
	if keyA == keyB {
		t.Fatal("keys must differ for different service names")
	}

	meanA, _, _ := p.getOrCreateStats(keyA).snapshot()
	meanB, _, _ := p.getOrCreateStats(keyB).snapshot()

	if meanA >= meanB {
		t.Errorf("service-a mean (%.2fms) should be less than service-b mean (%.2fms)", meanA/1e6, meanB/1e6)
	}
}

func TestProcessor_SameSpanNameDifferentNamespacesHaveIndependentBaselines(t *testing.T) {
	cfg := defaultConfig()
	p, _ := newTestProcessor(t, cfg)

	attrsProd := defaultResAttrs("ns-prod", "svc", "prod")
	attrsStaging := defaultResAttrs("ns-staging", "svc", "prod")

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 60; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsProd, "query", int64(20e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsStaging, "query", int64(300e6), now))
	}

	keyProd := keyFor(cfg, attrsProd, "query")
	keyStaging := keyFor(cfg, attrsStaging, "query")
	if keyProd == keyStaging {
		t.Fatal("keys must differ for different namespaces")
	}

	meanProd, _, _ := p.getOrCreateStats(keyProd).snapshot()
	meanStaging, _, _ := p.getOrCreateStats(keyStaging).snapshot()

	if meanProd >= meanStaging {
		t.Errorf("prod mean (%.2fms) should be less than staging mean (%.2fms)", meanProd/1e6, meanStaging/1e6)
	}
}

func TestProcessor_SameSpanNameDifferentEnvironmentsHaveIndependentBaselines(t *testing.T) {
	cfg := defaultConfig()
	p, _ := newTestProcessor(t, cfg)

	attrsEast := defaultResAttrs("ns", "svc", "us-east-1")
	attrsWest := defaultResAttrs("ns", "svc", "us-west-2")

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 60; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsEast, "SELECT", int64(5e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsWest, "SELECT", int64(400e6), now))
	}

	keyEast := keyFor(cfg, attrsEast, "SELECT")
	keyWest := keyFor(cfg, attrsWest, "SELECT")
	if keyEast == keyWest {
		t.Fatal("keys must differ for different deployment environments")
	}

	meanEast, _, _ := p.getOrCreateStats(keyEast).snapshot()
	meanWest, _, _ := p.getOrCreateStats(keyWest).snapshot()

	if meanEast >= meanWest {
		t.Errorf("east mean (%.2fms) should be less than west mean (%.2fms)", meanEast/1e6, meanWest/1e6)
	}
}

func TestProcessor_CustomResourceKeyAttributes(t *testing.T) {
	cfg := defaultConfig()
	cfg.ResourceKeyAttributes = []string{"service.name"}
	p, _ := newTestProcessor(t, cfg)

	attrsA := map[string]string{"service.namespace": "ns-a", "service.name": "svc"}
	attrsB := map[string]string{"service.namespace": "ns-b", "service.name": "svc"}

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 60; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsA, "op", int64(100e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(attrsB, "op", int64(100e6), now))
	}

	keyA := keyFor(cfg, attrsA, "op")
	keyB := keyFor(cfg, attrsB, "op")

	if keyA != keyB {
		t.Errorf("with single-key config, different namespaces should share the same key: keyA=%q keyB=%q", keyA, keyB)
	}
}

func TestEvict_RemovesStaleEntries(t *testing.T) {
	cfg := defaultConfig()
	cfg.IdleTimeout = time.Hour
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 15; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-a", int64(100e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), now))
	}

	keyA := keyFor(cfg, baseAttrs, "op-a")
	keyB := keyFor(cfg, baseAttrs, "op-b")

	evictTime := now.Add(cfg.IdleTimeout + time.Second)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), evictTime))

	p.evict(evictTime)

	p.statsMu.RLock()
	_, aExists := p.statsMap[keyA]
	_, bExists := p.statsMap[keyB]
	p.statsMu.RUnlock()

	if aExists {
		t.Error("op-a should have been evicted after idle timeout")
	}
	if !bExists {
		t.Error("op-b should not have been evicted — it was recently observed")
	}
}

func TestEvict_DoesNotRemoveActiveEntries(t *testing.T) {
	cfg := defaultConfig()
	cfg.IdleTimeout = time.Hour
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 15; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(100e6), now))
	}

	key := keyFor(cfg, baseAttrs, "op")

	p.evict(now.Add(cfg.IdleTimeout - time.Second))

	p.statsMu.RLock()
	_, exists := p.statsMap[key]
	p.statsMu.RUnlock()

	if !exists {
		t.Error("op should not be evicted before idle timeout elapses")
	}
}

func TestEvict_RelearnsAfterEviction(t *testing.T) {
	cfg := defaultConfig()
	cfg.IdleTimeout = time.Hour
	p, sink := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 50; i++ {
		now = now.Add(time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(100e6), now))
	}

	p.evict(now.Add(cfg.IdleTimeout + time.Second))

	key := keyFor(cfg, baseAttrs, "op")
	p.statsMu.RLock()
	_, exists := p.statsMap[key]
	p.statsMu.RUnlock()
	if exists {
		t.Fatal("entry should have been evicted")
	}

	sink.Reset()
	for i := 0; i < cfg.WarmupCount-1; i++ {
		now = now.Add(cfg.IdleTimeout + time.Second)
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(100e6), now))
	}

	if labels := collectLabels(sink, cfg.AttributeKey); len(labels) > 0 {
		t.Errorf("should not label during re-warmup after eviction, got %v", labels)
	}
}

func TestMaxBaselines_CapEnforced(t *testing.T) {
	cfg := defaultConfig()
	cfg.MaxBaselines = 2
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)

	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-a", int64(100e6), now))
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), now))

	p.statsMu.RLock()
	sizeBefore := len(p.statsMap)
	p.statsMu.RUnlock()
	if sizeBefore != 2 {
		t.Fatalf("expected 2 entries before cap, got %d", sizeBefore)
	}

	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-c", int64(100e6), now))

	p.statsMu.RLock()
	sizeAfter := len(p.statsMap)
	p.statsMu.RUnlock()
	if sizeAfter != 2 {
		t.Errorf("expected map to remain at cap (2), got %d", sizeAfter)
	}

	if p.droppedTotal.Load() != 1 {
		t.Errorf("expected droppedTotal=1, got %d", p.droppedTotal.Load())
	}
}

func TestMaxBaselines_ExistingKeyAllowedAfterCap(t *testing.T) {
	cfg := defaultConfig()
	cfg.MaxBaselines = 1
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(100e6), now))

	now = now.Add(time.Second)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op", int64(110e6), now))

	if p.droppedTotal.Load() != 0 {
		t.Errorf("existing key should not count as dropped, got %d", p.droppedTotal.Load())
	}
}

func TestMaxBaselines_DroppedCounterResetAfterEvictionSweep(t *testing.T) {
	cfg := defaultConfig()
	cfg.MaxBaselines = 1
	cfg.IdleTimeout = time.Hour
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-a", int64(100e6), now))
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), now)) // dropped

	if p.droppedTotal.Load() != 1 {
		t.Fatalf("expected 1 drop before sweep, got %d", p.droppedTotal.Load())
	}

	p.evict(now.Add(cfg.IdleTimeout + time.Second))

	if p.droppedTotal.Load() != 0 {
		t.Errorf("droppedTotal should be reset to 0 after eviction sweep, got %d", p.droppedTotal.Load())
	}
}

func TestEvict_ChurnWarningWhenHighTurnover(t *testing.T) {
	cfg := defaultConfig()
	cfg.IdleTimeout = time.Hour
	cfg.ChurnWarningRatio = 0.01
	p, _ := newTestProcessor(t, cfg)

	now := time.Unix(1_000_000, 0)
	for i := 0; i < 20; i++ {
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-a", int64(100e6), now))
		_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), now))
		now = now.Add(time.Second)
	}

	evictTime := now.Add(cfg.IdleTimeout)
	_ = p.ConsumeTraces(context.Background(), makeTraces(baseAttrs, "op-b", int64(100e6), evictTime))

	p.evict(evictTime)

	p.statsMu.RLock()
	_, aExists := p.statsMap[keyFor(cfg, baseAttrs, "op-a")]
	_, bExists := p.statsMap[keyFor(cfg, baseAttrs, "op-b")]
	p.statsMu.RUnlock()

	if aExists {
		t.Error("op-a should have been evicted")
	}
	if !bExists {
		t.Error("op-b should remain")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name:    "valid default",
			cfg:     defaultConfig(),
			wantErr: nil,
		},
		{
			name:    "zero half life",
			cfg:     Config{HalfLife: 0, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errInvalidHalfLife,
		},
		{
			name:    "very slow not greater than slow",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 3, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errVerySlowMustExceedSlow,
		},
		{
			name:    "empty attribute key",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errEmptyAttributeKey,
		},
		{
			name:    "empty resource key attributes",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errEmptyResourceKeyAttributes,
		},
		{
			name:    "zero idle timeout",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: 0, EvictionInterval: time.Minute, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errInvalidIdleTimeout,
		},
		{
			name:    "zero eviction interval",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: 0, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errInvalidEvictionInterval,
		},
		{
			name:    "negative max baselines",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: -1, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errNegativeMaxBaselines,
		},
		{
			name:    "churn warning ratio out of range",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0, WarmupCount: 30, MinStddev: 1e6},
			wantErr: errInvalidChurnWarningRatio,
		},
		{
			name:    "zero warmup count",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 0, MinStddev: 1e6},
			wantErr: errInvalidWarmupCount,
		},
		{
			name:    "negative min stddev",
			cfg:     Config{HalfLife: time.Hour, SlowThreshold: 3, VerySlowThreshold: 4, AttributeKey: "k", ResourceKeyAttributes: []string{"a"}, IdleTimeout: time.Hour, EvictionInterval: time.Minute, MaxBaselines: 0, ChurnWarningRatio: 0.5, WarmupCount: 30, MinStddev: -1},
			wantErr: errInvalidMinStddev,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if err != tc.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
