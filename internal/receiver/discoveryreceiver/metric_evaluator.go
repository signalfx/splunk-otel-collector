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
	"time"

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

func (m *metricEvaluator) evaluateMetrics(md pmetric.Metrics) plog.Logs {
	if ce := m.logger.Check(zapcore.DebugLevel, "evaluating metrics"); ce != nil {
		if mbytes, err := jsonMarshaler.MarshalMetrics(md); err == nil {
			ce.Write(zap.ByteString("metrics", mbytes))
		} else {
			m.logger.Debug("failed json-marshaling metrics for logging", zap.Error(err))
		}
	}
	pLogs := plog.NewLogs()
	if md.MetricCount() == 0 {
		return pLogs
	}

	receiverID, endpointID := statussources.MetricsToReceiverIDs(md)
	if receiverID == discovery.NoType || endpointID == "" {
		m.logger.Debug("unable to evaluate metrics from receiver without corresponding name or Endpoint.ID", zap.Any("metrics", md))
		return pLogs
	}

	rEntry, ok := m.config.Receivers[receiverID]
	if !ok {
		m.logger.Info("No matching configured receiver for metric status evaluation", zap.String("receiver", receiverID.String()))
		return pLogs
	}
	if rEntry.Status == nil || len(rEntry.Status.Metrics) == 0 {
		return pLogs
	}
	var matchFound bool

	stagePLogs := plog.NewLogs()
	rLog := stagePLogs.ResourceLogs().AppendEmpty()
	rAttrs := rLog.Resource().Attributes()
	m.correlateResourceAttributes(
		md.ResourceMetrics().At(0).Resource().Attributes(), rAttrs,
		m.correlations.GetOrCreate(receiverID, endpointID),
	)
	rAttrs.PutStr(eventTypeAttr, metricMatch)
	rAttrs.PutStr(receiverRuleAttr, rEntry.Rule)

	logRecords := rLog.ScopeLogs().AppendEmpty().LogRecords()

	receiverMetrics := map[string][]pmetric.Metric{}
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				receiverMetrics[m.Name()] = append(receiverMetrics[m.Name()], m)
			}
		}
	}

	for status, matches := range rEntry.Status.Metrics {
		for _, match := range matches {
			for metricName, metrics := range receiverMetrics {
				for _, metric := range metrics {
					if shouldLog, err := m.evaluateMatch(match, metricName, status, receiverID, endpointID); err != nil {
						m.logger.Info(fmt.Sprintf("Error evaluating %s metric match", status), zap.Error(err))
						continue
					} else if !shouldLog {
						continue
					}
					matchFound = true
					logRecord := logRecords.AppendEmpty()
					desiredRecord := match.Record
					if desiredRecord == nil {
						desiredRecord = &LogRecord{}
					}
					var desiredBody string
					if desiredRecord.Body != "" {
						desiredBody = desiredRecord.Body
					}
					logRecord.Body().SetStr(desiredBody)
					for k, v := range desiredRecord.Attributes {
						logRecord.Attributes().PutStr(k, v)
					}
					severityText := desiredRecord.SeverityText
					if severityText == "" {
						severityText = "info"
					}
					logRecord.SetSeverityText(severityText)
					logRecord.Attributes().PutStr(metricNameAttr, metricName)
					logRecord.Attributes().PutStr(discovery.StatusAttr, string(status))
					if ts := m.timestampFromMetric(metric); ts != nil {
						logRecord.SetTimestamp(*ts)
					}
					logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
				}
			}
		}
	}
	if matchFound {
		pLogs = stagePLogs
	}
	return pLogs
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
