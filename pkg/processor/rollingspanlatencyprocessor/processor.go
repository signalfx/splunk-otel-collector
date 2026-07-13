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

package rollingspanlatencyprocessor // import "github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor"

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor/internal/metadata"
)

const (
	attributeValueSlow     = "slow"
	attributeValueVerySlow = "very_slow"

	metricActiveBaselines = "rollingspanlatency_active_baselines"
	metricDroppedKeys     = "rollingspanlatency_dropped_keys_total"
)

type rollingSpanLatencyProcessor struct {
	config      Config
	logger      *zap.Logger
	next        consumer.Traces
	statsMu     sync.RWMutex
	statsMap    map[string]*spanStats // keyed by buildKey(resourceVals, spanName)
	nowFn       func() time.Time      // injectable for testing
	cancelEvict context.CancelFunc

	// droppedTotal counts keys dropped due to the max_baselines cap since the
	// last eviction sweep. Reset to 0 after each sweep logs the value.
	droppedTotal atomic.Int64
}

// buildKey returns a composite stats-map key from an ordered slice of resource
// attribute values and the span name. \x00 is the separator; it cannot appear
// in OTel attribute values in practice, so collisions are not possible.
func buildKey(resourceVals []string, spanName string) string {
	key := spanName
	for _, v := range resourceVals {
		key = v + "\x00" + key
	}
	return key
}

func newProcessor(cfg Config, telemetry component.TelemetrySettings, next consumer.Traces) (*rollingSpanLatencyProcessor, error) {
	p := &rollingSpanLatencyProcessor{
		config:   cfg,
		logger:   telemetry.Logger,
		next:     next,
		statsMap: make(map[string]*spanStats),
		nowFn:    time.Now,
	}
	if err := p.registerMetrics(telemetry.MeterProvider); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *rollingSpanLatencyProcessor) registerMetrics(mp metric.MeterProvider) error {
	meter := mp.Meter(metadata.ScopeName)

	_, err := meter.Int64ObservableGauge(
		metricActiveBaselines,
		metric.WithDescription("Number of span baseline entries currently held in memory."),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			p.statsMu.RLock()
			n := int64(len(p.statsMap))
			p.statsMu.RUnlock()
			o.Observe(n)
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableCounter(
		metricDroppedKeys,
		metric.WithDescription("Total number of new baseline keys dropped because max_baselines was reached."),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(p.droppedTotal.Load())
			return nil
		}),
	)
	return err
}

func (p *rollingSpanLatencyProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *rollingSpanLatencyProcessor) Start(_ context.Context, _ component.Host) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancelEvict = cancel
	go p.evictLoop(ctx)
	return nil
}

func (p *rollingSpanLatencyProcessor) Shutdown(_ context.Context) error {
	if p.cancelEvict != nil {
		p.cancelEvict()
	}
	return nil
}

func (p *rollingSpanLatencyProcessor) evictLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.EvictionInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.evict(p.nowFn())
		}
	}
}

func (p *rollingSpanLatencyProcessor) evict(now time.Time) {
	cutoff := now.Add(-p.config.IdleTimeout)

	p.statsMu.Lock()
	before := len(p.statsMap)
	for key, s := range p.statsMap {
		if s.idleSince().Before(cutoff) {
			delete(p.statsMap, key)
		}
	}
	after := len(p.statsMap)
	p.statsMu.Unlock()

	evicted := before - after
	dropped := p.droppedTotal.Load()

	if evicted > 0 {
		fields := []zap.Field{
			zap.Int("evicted", evicted),
			zap.Int("remaining", after),
			zap.Int64("dropped_since_last_sweep", dropped),
		}
		// Churn warning: evicted count exceeded the configured ratio of the
		// post-eviction map size. This indicates keys are turning over rapidly,
		// which often means span names contain high-cardinality values.
		if after > 0 && float64(evicted)/float64(after) > p.config.ChurnWarningRatio {
			p.logger.Warn("high baseline key churn detected — check for high-cardinality span names",
				fields...,
			)
		} else {
			p.logger.Debug("evicted stale span baselines", fields...)
		}
	}

	// Reset the per-interval drop counter now that we've reported it.
	p.droppedTotal.Store(0)
}

func (p *rollingSpanLatencyProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		resAttrs := rs.Resource().Attributes()
		resourceVals := make([]string, len(p.config.ResourceKeyAttributes))
		for idx, attrKey := range p.config.ResourceKeyAttributes {
			if v, ok := resAttrs.Get(attrKey); ok {
				resourceVals[idx] = v.Str()
			}
		}
		scopeSpans := rs.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			spans := scopeSpans.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				p.processSpan(spans.At(k), resourceVals)
			}
		}
	}
	return p.next.ConsumeTraces(ctx, td)
}

func (p *rollingSpanLatencyProcessor) processSpan(span ptrace.Span, resourceVals []string) {
	key := buildKey(resourceVals, span.Name())
	durationNs := float64(span.EndTimestamp() - span.StartTimestamp())
	if durationNs <= 0 {
		return
	}
	// Use the span's own end timestamp so spans within the same batch each
	// advance the EWMA clock correctly. A batch-shared wall-clock time would
	// give dt=0 for all but the first span, collapsing alpha to 0 and
	// leaving variance near-zero.
	now := time.Unix(0, int64(span.EndTimestamp()))

	stats := p.getOrCreateStats(key)
	if stats == nil {
		// Cap reached; this span has no baseline — skip attribute write.
		return
	}

	// Snapshot the baseline before updating so the current span is scored
	// against historical data only — prevents a single outlier from
	// inflating its own stddev and masking its own anomaly.
	preMean, preStddev, preCount := stats.snapshot()
	stats.update(durationNs, now, p.config.HalfLife)

	if preCount < int64(p.config.WarmupCount) {
		return
	}

	effectiveStddev := preStddev
	if effectiveStddev < p.config.MinStddev {
		effectiveStddev = p.config.MinStddev
	}

	deviations := (durationNs - preMean) / effectiveStddev
	switch {
	case deviations >= p.config.VerySlowThreshold:
		span.Attributes().PutStr(p.config.AttributeKey, attributeValueVerySlow)
	case deviations >= p.config.SlowThreshold:
		span.Attributes().PutStr(p.config.AttributeKey, attributeValueSlow)
	}
}

// getOrCreateStats returns the spanStats for key, creating it if absent.
// Returns nil when the max_baselines cap is reached and the key is new.
func (p *rollingSpanLatencyProcessor) getOrCreateStats(key string) *spanStats {
	p.statsMu.RLock()
	s, ok := p.statsMap[key]
	p.statsMu.RUnlock()
	if ok {
		return s
	}

	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	// double-checked locking
	if s, ok = p.statsMap[key]; ok {
		return s
	}

	if p.config.MaxBaselines > 0 && len(p.statsMap) >= p.config.MaxBaselines {
		p.droppedTotal.Add(1)
		p.logger.Warn("max_baselines cap reached; dropping new baseline key",
			zap.String("key", key),
			zap.Int("max_baselines", p.config.MaxBaselines),
		)
		return nil
	}

	s = &spanStats{}
	p.statsMap[key] = s
	return s
}
