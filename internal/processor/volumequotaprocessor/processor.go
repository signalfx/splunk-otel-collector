// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package volumequotaprocessor

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	conventions "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

type volumeQuotaProcessor struct {
	next           consumer.Traces
	config         *Config
	logger         *zap.Logger
	shutdownChan   chan struct{}
	currentEpoch   *epoch
	lookbackEpochs []epoch
	wg             sync.WaitGroup
	lock           sync.RWMutex
}

type epoch struct {
	servicesSpansSamplingRate  map[string]float64
	servicesTracesSamplingRate map[string]float64
	serviceSpansCount          map[string]int64
	serviceTracesCount         map[string]int64
	serviceTraceIdsTracker     map[string]map[string]struct{}
	traceIdsTracker            map[string]struct{}
	globalSpanSamplingRate     float64
	globalTracesSamplingRate   float64
	globalSpanCount            int64
	globalTracesCount          int64
}

func (p *volumeQuotaProcessor) Start(_ context.Context, _ component.Host) error {
	p.wg.Go(func() {
		ticker := time.NewTicker(p.config.Epoch)
		p.rotateEpoch()
		for {
			select {
			case <-ticker.C:
				p.rotateEpoch()
			case <-p.shutdownChan:
				return
			}
		}
	})
	return nil
}

func (p *volumeQuotaProcessor) Shutdown(_ context.Context) error {
	close(p.shutdownChan)
	p.wg.Wait()
	return nil
}

func newVolumeQuotaProcessor(_ context.Context, cfg *Config, set processor.Settings, next consumer.Traces) (processor.Traces, error) {
	p := &volumeQuotaProcessor{
		config:       cfg,
		logger:       set.Logger,
		next:         next,
		lock:         sync.RWMutex{},
		shutdownChan: make(chan struct{}),
	}

	return p, nil
}

func (*volumeQuotaProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *volumeQuotaProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		sms := rs.ScopeSpans()
		for j := 0; j < sms.Len(); j++ {
			sm := sms.At(j)
			ss := sm.Spans()
			for k := 0; k < ss.Len(); k++ {
				p.processSpan(ss.At(k))
			}
		}
	}
	return p.next.ConsumeTraces(ctx, td)
}

func (p *volumeQuotaProcessor) processSpan(s ptrace.Span) {
	p.lock.Lock()
	defer p.lock.Unlock()
	rateLimit := 1.0

	if p.config.GlobalLimits.Spans > 0 {
		rateLimit = p.currentEpoch.globalSpanSamplingRate
		p.currentEpoch.globalSpanCount++
		if p.currentEpoch.globalSpanCount > p.config.GlobalLimits.Spans {
			globalSpanLimit := 1 - (float64(p.currentEpoch.globalSpanCount) / float64(p.config.GlobalLimits.Spans))
			if globalSpanLimit < rateLimit {
				rateLimit = globalSpanLimit
			}
		}
	}

	if p.config.GlobalLimits.Traces > 0 {
		if p.currentEpoch.globalTracesSamplingRate < rateLimit {
			rateLimit = p.currentEpoch.globalTracesSamplingRate
		}
		traceIdStr := s.TraceID().String()
		if _, ok := p.currentEpoch.traceIdsTracker[traceIdStr]; !ok {
			p.currentEpoch.traceIdsTracker[traceIdStr] = struct{}{}
			p.currentEpoch.globalTracesCount++
		}
		if p.currentEpoch.globalTracesCount > p.config.GlobalLimits.Traces {
			globalTracesLimit := 1 - (float64(p.currentEpoch.globalTracesCount) / float64(p.config.GlobalLimits.Traces))
			if globalTracesLimit < rateLimit {
				rateLimit = globalTracesLimit
			}
		}
	}

	// update service counters
	if serviceNameValue, ok := s.Attributes().Get(string(conventions.ServiceNameKey)); ok {
		serviceName := serviceNameValue.Str()
		if v, ok := p.currentEpoch.serviceSpansCount[serviceName]; ok {
			if p.currentEpoch.servicesSpansSamplingRate[serviceName] < rateLimit {
				rateLimit = p.currentEpoch.servicesSpansSamplingRate[serviceName]
			}
			p.currentEpoch.serviceSpansCount[serviceName] = v + 1
			if v+1 > p.config.Limits.Spans[serviceName] {
				serviceSpanLimit := 1 - (float64(v+1) / float64(p.config.Limits.Spans[serviceName]))
				if serviceSpanLimit < rateLimit {
					rateLimit = serviceSpanLimit
				}
			}
		}
		if tracker, ok := p.currentEpoch.serviceTraceIdsTracker[serviceName]; ok {
			if p.currentEpoch.servicesTracesSamplingRate[serviceName] < rateLimit {
				rateLimit = p.currentEpoch.servicesTracesSamplingRate[serviceName]
			}
			traceIDstr := s.TraceID().String()
			if _, ok := tracker[traceIDstr]; !ok {
				tracker[traceIDstr] = struct{}{}
				p.currentEpoch.serviceTracesCount[serviceName]++
			}
			if p.currentEpoch.serviceTracesCount[serviceName] > p.config.Limits.Traces[serviceName] {
				serviceTraceLimit := 1 - (float64(p.currentEpoch.serviceTracesCount[serviceName]) / float64(p.config.Limits.Traces[serviceName]))
				if serviceTraceLimit < rateLimit {
					rateLimit = serviceTraceLimit
				}
			}
		}
	}
	if rateLimit < 0 {
		rateLimit = 0
	}
	s.Attributes().PutInt("sampling.priority", int64(rateLimit*100))
}

func (p *volumeQuotaProcessor) rotateEpoch() {
	p.lock.Lock()
	defer p.lock.Unlock()
	// record current epoch, drop latest epoch if needed
	if p.config.Lookback > 0 && p.currentEpoch != nil {
		pastLookbacks := p.lookbackEpochs
		if len(p.lookbackEpochs) == p.config.Lookback {
			pastLookbacks = pastLookbacks[1:]
		}
		p.lookbackEpochs = append(pastLookbacks, *p.currentEpoch)
	}
	serviceSpans := make(map[string]int64, len(p.config.Limits.Spans))
	for k := range p.config.Limits.Spans {
		serviceSpans[k] = 0
	}
	serviceTraces := make(map[string]int64, len(p.config.Limits.Traces))
	serviceTracesTracker := make(map[string]map[string]struct{}, len(p.config.Limits.Traces))
	for k := range p.config.Limits.Traces {
		serviceTraces[k] = 0
		serviceTracesTracker[k] = map[string]struct{}{}
	}
	newEpoch := &epoch{
		globalSpanSamplingRate:     1.0,
		globalTracesSamplingRate:   1.0,
		servicesSpansSamplingRate:  make(map[string]float64, len(p.config.Limits.Spans)),
		servicesTracesSamplingRate: make(map[string]float64, len(p.config.Limits.Traces)),
		globalSpanCount:            0,
		globalTracesCount:          0,
		serviceSpansCount:          serviceSpans,
		serviceTracesCount:         serviceTraces,
		serviceTraceIdsTracker:     serviceTracesTracker,
		traceIdsTracker:            map[string]struct{}{},
	}
	p.determineStartSamplingRate(newEpoch)
	p.currentEpoch = newEpoch
}

func (p *volumeQuotaProcessor) determineStartSamplingRate(newEpoch *epoch) {
	if p.config.GlobalLimits.Spans > 0 {
		var total int64
		for _, e := range p.lookbackEpochs {
			total += e.globalSpanCount
		}
		used := (float64(len(p.lookbackEpochs)) * float64(p.config.GlobalLimits.Spans)) / float64(total)
		if used > 1 || total == 0 {
			newEpoch.globalSpanSamplingRate = 1
		} else {
			newEpoch.globalSpanSamplingRate = used
		}

	}

	if p.config.GlobalLimits.Traces > 0 {
		var total int64
		for _, e := range p.lookbackEpochs {
			total += e.globalTracesCount
		}
		used := (float64(len(p.lookbackEpochs)) * float64(p.config.GlobalLimits.Traces)) / float64(total)
		if used > 1 || total == 0 {
			newEpoch.globalTracesSamplingRate = 1
		} else {
			newEpoch.globalTracesSamplingRate = used
		}
	}

	for serviceName, limit := range p.config.Limits.Spans {
		var total int64
		for _, e := range p.lookbackEpochs {
			total += e.serviceSpansCount[serviceName]
		}
		used := (float64(len(p.lookbackEpochs)) * float64(limit)) / float64(total)
		if used > 1 || total == 0 {
			newEpoch.servicesSpansSamplingRate[serviceName] = 1
		} else {
			newEpoch.servicesSpansSamplingRate[serviceName] = used
		}
	}

	for serviceName, limit := range p.config.Limits.Traces {
		var total int64
		for _, e := range p.lookbackEpochs {
			total += e.serviceTracesCount[serviceName]
		}
		used := (float64(len(p.lookbackEpochs)) * float64(limit)) / float64(total)
		if used > 1 || total == 0 {
			newEpoch.servicesTracesSamplingRate[serviceName] = 1
		} else {
			newEpoch.servicesTracesSamplingRate[serviceName] = used
		}
	}
}
