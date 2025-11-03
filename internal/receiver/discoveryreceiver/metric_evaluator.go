// Copyright Splunk, Inc.
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

package discoveryreceiver

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver/statussources"
)

var _ consumer.Metrics = (*metricsConsumer)(nil)

var jsonMarshaler = &pmetric.JSONMarshaler{}

// metricsConsumer conforms to a consumer.Metrics to receive any metrics from
// receiver creator-created receivers and determine if they match any configured
// Status match rules. If so, they emit log records for the matching metric.
// It also passes the metrics to the next consumer in the pipeline.
type metricsConsumer struct {
	*evaluator
	nextConsumer consumer.Metrics
}

func newMetricsConsumer(logger *zap.Logger, cfg *Config, correlations *correlationStore, nextConsumer consumer.Metrics) *metricsConsumer {
	mc := &metricsConsumer{nextConsumer: nextConsumer}
	if correlations != nil {
		mc.evaluator = newEvaluator(logger, cfg, correlations,
			// TODO: provide more capable env w/ resource and metric attributes
			func(pattern string) map[string]any {
				return map[string]any{"name": pattern}
			})
	}
	return mc
}

func (m *metricsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (m *metricsConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.evaluateMetrics(md)
	if m.nextConsumer != nil {
		return m.nextConsumer.ConsumeMetrics(ctx, cleanupMetrics(md))
	}
	return nil
}

// evaluateMetrics parses the provided Metrics and returns plog.Logs with a single log record if it matches
// against the first applicable configured Status match rule.
func (m *metricsConsumer) evaluateMetrics(md pmetric.Metrics) {
	if m.evaluator == nil {
		return
	}
	if ce := m.logger.Check(zapcore.DebugLevel, "evaluating metrics"); ce != nil {
		if mbytes, err := jsonMarshaler.MarshalMetrics(md); err == nil {
			ce.Write(zap.ByteString("metrics", mbytes))
		} else {
			m.logger.Debug("failed json-marshaling metrics for logging", zap.Error(err))
		}
	}
	if md.MetricCount() == 0 {
		return
	}

	receiverID, endpointID := statussources.MetricsToReceiverIDs(md)
	if receiverID == discovery.NoType || endpointID == "" {
		m.logger.Debug("unable to evaluate metrics from receiver without corresponding name or Endpoint.ID", zap.Any("metrics", md))
		return
	}

	_, ok := m.config.Receivers[receiverID]
	if !ok {
		m.logger.Info("No matching configured receiver for metric status evaluation", zap.String("receiver", receiverID.String()))
		return
	}

	meta, hasMeta := receiverMetaMap[receiverID.String()]
	if !hasMeta || len(meta.Status.Metrics) == 0 {
		m.logger.Warn("No metadata found for receiver", zap.String("receiver", receiverID.String()))
		return
	}

	for _, match := range meta.Status.Metrics {
		res, matched := m.findMatchedMetric(md, match, receiverID, endpointID)
		if !matched {
			continue
		}

		corr := m.correlations.GetOrCreate(endpointID, receiverID)
		attrs := m.correlations.Attrs(endpointID)

		// If the status is already the same as desired, we don't need to update the entity state.
		if match.Status == discovery.StatusType(attrs[discovery.StatusAttr]) {
			return
		}

		res.Attributes().Range(func(k string, v pcommon.Value) bool {
			// skip endpoint ID attr since it's set in the entity ID
			if k == discovery.EndpointIDAttr {
				return true
			}
			attrs[k] = v.AsString()
			return true
		})
		m.correlateResourceAttributes(m.config, attrs, corr)

		attrs[discovery.MessageAttr] = match.Message
		attrs[discovery.StatusAttr] = string(match.Status)
		m.correlations.UpdateAttrs(endpointID, attrs)

		m.correlations.emitCh <- corr
		return
	}
}

// findMatchedMetric finds the metric that matches the provided match rule and return the resource where it's found.
func (m *metricsConsumer) findMatchedMetric(md pmetric.Metrics, match Match, receiverID component.ID, endpointID observer.EndpointID) (pcommon.Resource, bool) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				if shouldLog, err := m.evaluateMatch(match, metric.Name(), match.Status, receiverID, endpointID); err != nil {
					m.logger.Info(fmt.Sprintf("Error evaluating %s metric match", metric.Name()), zap.Error(err))
					continue
				} else if !shouldLog {
					continue
				}
				return rm.Resource(), true
			}
		}
	}
	return pcommon.NewResource(), false
}

// cleanupMetrics removes resource attributes used for status correlation.
func cleanupMetrics(md pmetric.Metrics) pmetric.Metrics {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		attrs := md.ResourceMetrics().At(i).Resource().Attributes()
		attrs.Remove(discovery.ReceiverTypeAttr)
		attrs.Remove(discovery.ReceiverNameAttr)
	}
	return md
}
