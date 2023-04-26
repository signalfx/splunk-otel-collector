// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsinkexporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type client struct {
	spans   chan *model.Batch
	metrics chan pmetric.Metrics
	opts    options
	stopped bool
}

func newClient(opts options) *client {
	return &client{
		spans:   make(chan *model.Batch),
		metrics: make(chan pmetric.Metrics),
		opts:    opts,
	}
}

func (c *client) response(ctx context.Context) ([]byte, error) {
	// TODO: add support to filter spans by attributes
	defer func() {
		c.stopped = true
	}()

	results := []string{}
	received := 0

	done := make(chan struct{})
	for {
		select {
		case batch := <-c.spans:
			batch = c.filterSpans(batch)
			for _, span := range batch.Spans {
				json, err := spanMarshaler.MarshalToString(span)
				if err != nil {
					return nil, err
				}
				results = append(results, json)
				received++
				if received == c.opts.count {
					close(done)
				}
			}
		case md := <-c.metrics:
			md = c.filterMetrics(md)
			count := md.MetricCount()
			if count > 0 {
				json, err := metricMarshaler.MarshalMetrics(md)
				if err != nil {
					return nil, err
				}
				results = append(results, string(json))
				received += count
			}
			if received >= c.opts.count {
				close(done)
			}

		case <-time.After(c.opts.timeout):
			return nil, fmt.Errorf("timed out while waiting for results")

		case <-ctx.Done():
			return nil, fmt.Errorf("context deadline exceeded")

		case <-done:
			result := "[" + strings.Join(results, ",") + "]"
			return []byte(result), nil
		}
	}
}

func (c *client) filterSpans(b *model.Batch) *model.Batch {
	spans := []*model.Span{}

	for _, span := range b.Spans {
		match := c.filterByName(span.OperationName)
		if match {
			spans = append(spans, span)
		}
	}

	return &model.Batch{
		Process: b.Process,
		Spans:   spans,
	}
}

func (c *client) filterMetrics(md pmetric.Metrics) pmetric.Metrics {
	m := pmetric.NewMetrics()
	md.CopyTo(m)
	for i := 0; i < m.ResourceMetrics().Len(); i++ {
		rm := m.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			sm.Metrics().RemoveIf(func(metric pmetric.Metric) bool {
				matched := c.filterByName(metric.Name()) || c.filterMetricByAttrs(metric)
				return !matched
			})
		}
	}
	return m
}

func (c *client) filterByName(name string) bool {
	if len(c.opts.names) == 0 {
		return true
	}

	for _, n := range c.opts.names {
		if name == n {
			return true
		}
	}
	return false
}

func (c *client) filterMetricByAttrs(metric pmetric.Metric) bool {
	if len(c.opts.attrs) == 0 {
		return true
	}

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return filterMetricByAttr[pmetric.NumberDataPoint, pmetric.NumberDataPointSlice](metric.Gauge(), c.opts.attrs)

	case pmetric.MetricTypeSum:
		return filterMetricByAttr[pmetric.NumberDataPoint, pmetric.NumberDataPointSlice](metric.Sum(), c.opts.attrs)

	case pmetric.MetricTypeHistogram:
		return filterMetricByAttr[pmetric.HistogramDataPoint, pmetric.HistogramDataPointSlice](metric.Histogram(), c.opts.attrs)

	case pmetric.MetricTypeExponentialHistogram:
		return filterMetricByAttr[pmetric.ExponentialHistogramDataPoint, pmetric.ExponentialHistogramDataPointSlice](metric.ExponentialHistogram(), c.opts.attrs)

	case pmetric.MetricTypeSummary:
		return filterMetricByAttr[pmetric.SummaryDataPoint, pmetric.SummaryDataPointSlice](metric.Summary(), c.opts.attrs)
	}
	return false
}
