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
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver/statussources"
)

var _ consumer.Metrics = (*metricEvaluator)(nil)

const (
	metricMatch = "metric.match"
)

var (
	jsonMarshaler = &pmetric.JSONMarshaler{}
)

// metricEvaluator conforms to a consumer.Metrics to receive any metrics from
// receiver creator-created receivers and determine if they match any configured
// Status match rules. If so, they emit log records for the matching metric.
type metricEvaluator struct {
	*evaluator
	pLogs chan plog.Logs
}

func newMetricEvaluator(logger *zap.Logger, cfg *Config, pLogs chan plog.Logs, correlations correlationStore) *metricEvaluator {
	return &metricEvaluator{
		pLogs: pLogs,
		evaluator: newEvaluator(logger, cfg, correlations,
			// TODO: provide more capable env w/ resource and metric attributes
			func(pattern string) map[string]any {
				return map[string]any{"name": pattern}
			},
		),
	}
}

func (m *metricEvaluator) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (m *metricEvaluator) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	if pLogs := m.evaluateMetrics(md); pLogs.LogRecordCount() > 0 {
		m.pLogs <- pLogs
	}
	return nil
}

// evaluateMetrics parses the provided Metrics and returns plog.Logs with a single log record if it matches
// against the first applicable configured Status match rule.
func (m *metricEvaluator) evaluateMetrics(md pmetric.Metrics) plog.Logs {
	if ce := m.logger.Check(zapcore.DebugLevel, "evaluating metrics"); ce != nil {
		if mbytes, err := jsonMarshaler.MarshalMetrics(md); err == nil {
			ce.Write(zap.ByteString("metrics", mbytes))
		} else {
			m.logger.Debug("failed json-marshaling metrics for logging", zap.Error(err))
		}
	}
	if md.MetricCount() == 0 {
		return plog.NewLogs()
	}

	receiverID, endpointID := statussources.MetricsToReceiverIDs(md)
	if receiverID == discovery.NoType || endpointID == "" {
		m.logger.Debug("unable to evaluate metrics from receiver without corresponding name or Endpoint.ID", zap.Any("metrics", md))
		return plog.NewLogs()
	}

	rEntry, ok := m.config.Receivers[receiverID]
	if !ok {
		m.logger.Info("No matching configured receiver for metric status evaluation", zap.String("receiver", receiverID.String()))
		return plog.NewLogs()
	}
	if rEntry.Status == nil || len(rEntry.Status.Metrics) == 0 {
		return plog.NewLogs()
	}

	for _, match := range rEntry.Status.Metrics {
		res, metric, matched := m.findMatchedMetric(md, match, receiverID, endpointID)
		if !matched {
			continue
		}

		entityEvents := experimentalmetricmetadata.NewEntityEventsSlice()
		entityEvent := entityEvents.AppendEmpty()
		entityEvent.ID().PutStr(discovery.EndpointIDAttr, string(endpointID))
		entityState := entityEvent.SetEntityState()

		res.Attributes().CopyTo(entityState.Attributes())
		corr := m.correlations.GetOrCreate(endpointID, receiverID)
		m.correlateResourceAttributes(m.config, entityState.Attributes(), corr)

		// Remove the endpoint ID from the attributes as it's set in the entity ID.
		entityState.Attributes().Remove(discovery.EndpointIDAttr)

		entityState.Attributes().PutStr(eventTypeAttr, metricMatch)
		entityState.Attributes().PutStr(receiverRuleAttr, rEntry.Rule.String())

		desiredRecord := match.Record
		if desiredRecord == nil {
			desiredRecord = &LogRecord{}
		}
		var desiredMsg string
		if desiredRecord.Body != "" {
			desiredMsg = desiredRecord.Body
		}
		entityState.Attributes().PutStr(discovery.MessageAttr, desiredMsg)
		for k, v := range desiredRecord.Attributes {
			entityState.Attributes().PutStr(k, v)
		}
		entityState.Attributes().PutStr(metricNameAttr, metric.Name())
		entityState.Attributes().PutStr(discovery.StatusAttr, string(match.Status))
		if ts := m.timestampFromMetric(metric); ts != nil {
			entityEvent.SetTimestamp(*ts)
		}

		return entityEvents.ConvertAndMoveToLogs()
	}
	return plog.NewLogs()
}

// findMatchedMetric finds the metric that matches the provided match rule and return it along with the resource if found.
func (m *metricEvaluator) findMatchedMetric(md pmetric.Metrics, match Match, receiverID component.ID, endpointID observer.EndpointID) (pcommon.Resource, pmetric.Metric, bool) {
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
				return rm.Resource(), metric, true
			}
		}
	}
	return pcommon.NewResource(), pmetric.NewMetric(), false
}

func (m *metricEvaluator) timestampFromMetric(metric pmetric.Metric) *pcommon.Timestamp {
	var ts *pcommon.Timestamp
	switch dt := metric.Type(); dt {
	case pmetric.MetricTypeGauge:
		dps := metric.Gauge().DataPoints()
		if dps.Len() > 0 {
			t := dps.At(0).Timestamp()
			ts = &t
		}
	case pmetric.MetricTypeSum:
		dps := metric.Sum().DataPoints()
		if dps.Len() > 0 {
			t := dps.At(0).Timestamp()
			ts = &t
		}
	case pmetric.MetricTypeHistogram:
		dps := metric.Histogram().DataPoints()
		if dps.Len() > 0 {
			t := dps.At(0).Timestamp()
			ts = &t
		}
	case pmetric.MetricTypeExponentialHistogram:
		dps := metric.ExponentialHistogram().DataPoints()
		if dps.Len() > 0 {
			t := dps.At(0).Timestamp()
			ts = &t
		}
	case pmetric.MetricTypeSummary:
		dps := metric.Summary().DataPoints()
		if dps.Len() > 0 {
			t := dps.At(0).Timestamp()
			ts = &t
		}
	default:
		m.logger.Debug("cannot get timestamp from data type", zap.String("data type", dt.String()))
	}
	return ts
}
