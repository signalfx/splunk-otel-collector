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

package metrics // import "github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension/internal/unmarshaler/metrics"

import (
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// metricsBuilder accumulates OCI metric records into pmetric.Metrics,
// grouping and merging them as records are added via unmarshalRecord.
type metricsBuilder struct {
	logger *zap.Logger

	allResourceMetrics map[resourceIdentity]pmetric.ResourceMetrics
	allMetrics         map[metricIdentity]map[string]pmetric.Metric
}

func newMetricsBuilder(logger *zap.Logger) *metricsBuilder {
	return &metricsBuilder{
		logger:             logger,
		allResourceMetrics: map[resourceIdentity]pmetric.ResourceMetrics{},
		allMetrics:         map[metricIdentity]map[string]pmetric.Metric{},
	}
}

// unmarshalRecord parses a single JSON OCI metric record and merges
// it into the builder's accumulated state. Records sharing the same
// compartment, namespace, resource group and resource ID are grouped into a
// single ResourceMetrics. Within a ResourceMetrics, records sharing the same
// metric name and unit are merged into a single Metric.
func (b *metricsBuilder) unmarshalRecord(jsonRecord []byte) {
	rec, err := b.getValidRecord(jsonRecord)
	if err != nil {
		b.logger.Warn("Skipping invalid OCI metric record", zap.Error(err))
		return
	}

	dataPoints := b.getDatapoints(rec)

	if dataPoints.Len() == 0 {
		b.logger.Warn("Skipping OCI metric record without valid datapoints",
			zap.Any("name", rec.Name),
			zap.Any("namespace", rec.Namespace),
			zap.Any("datapoints", rec.Datapoints))
		return
	}

	resourceID := extractResourceID(rec.Dimensions)
	resourceKey := resourceIdentity{
		compartmentID: rec.CompartmentID,
		namespace:     rec.Namespace,
		resourceGroup: rec.ResourceGroup,
		resourceID:    resourceID,
	}

	rm, found := b.allResourceMetrics[resourceKey]
	if !found {
		rm = pmetric.NewResourceMetrics()
		for k, v := range resourceAttributes(*rec, resourceID) {
			rm.Resource().Attributes().PutStr(k, v)
		}
		rm.ScopeMetrics().AppendEmpty().Scope().SetName(ScopeName)
		b.allResourceMetrics[resourceKey] = rm
	}

	metricKey := metricIdentity{resource: resourceKey, name: rec.Name}
	metricsByUnit, found := b.allMetrics[metricKey]
	if !found {
		metricsByUnit = map[string]pmetric.Metric{}
		b.allMetrics[metricKey] = metricsByUnit
	}

	m, found := metricsByUnit[rec.Metadata.Unit]
	if !found {
		if len(metricsByUnit) > 0 {
			b.logger.Warn(
				"Conflicting units for OCI metric records",
				zap.String("name", rec.Name),
				zap.String("namespace", rec.Namespace),
				zap.String("unit", rec.Metadata.Unit),
			)
		}

		m = rm.ScopeMetrics().At(0).Metrics().AppendEmpty()
		m.SetName(rec.Name)
		if rec.Metadata.Unit != "" {
			m.SetUnit(rec.Metadata.Unit)
		}
		// OCI Monitoring does not report an explicit metric type, and
		// metadata.unit is descriptive (e.g. "ms") so always use gauge.
		m.SetEmptyGauge()
		metricsByUnit[rec.Metadata.Unit] = m
	}

	// Description is not identifying. Prefer the longer description when
	// repeated records disagree, following the OTel producer recommendation.
	if len(rec.Metadata.DisplayName) > len(m.Description()) {
		m.SetDescription(rec.Metadata.DisplayName)
	}

	dataPoints.MoveAndAppendTo(m.Gauge().DataPoints())
}

// build returns the accumulated ResourceMetrics as a pmetric.Metrics, with
// each metric's datapoints sorted by timestamp, oldest first.
func (b *metricsBuilder) build() pmetric.Metrics {
	for _, metricsByUnit := range b.allMetrics {
		for _, m := range metricsByUnit {
			m.Gauge().DataPoints().Sort(func(a, b pmetric.NumberDataPoint) bool {
				return a.Timestamp() < b.Timestamp()
			})
		}
	}

	md := pmetric.NewMetrics()
	for _, rm := range b.allResourceMetrics {
		rm.MoveTo(md.ResourceMetrics().AppendEmpty())
	}
	return md
}

func (b *metricsBuilder) getValidRecord(jsonRecord []byte) (*ociMetricRecord, error) {
	var rec ociMetricRecord
	if err := json.Unmarshal(jsonRecord, &rec); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed for OCI metric record: %w", err)
	}

	if rec.Name == "" {
		return nil, fmt.Errorf(
			"no name set on OCI metric record (namespace=%q, compartmentId=%q)",
			rec.Namespace, rec.CompartmentID,
		)
	}

	if rec.CompartmentID == "" {
		return nil, fmt.Errorf(
			"no compartmentId set on OCI metric record (namespace=%q, name=%q)",
			rec.Namespace, rec.Name,
		)
	}

	if rec.Namespace == "" {
		return nil, fmt.Errorf(
			"no namespace set on OCI metric record (compartmentId=%q, name=%q)",
			rec.CompartmentID, rec.Name,
		)
	}

	return &rec, nil
}

func (b *metricsBuilder) getDatapoints(rec *ociMetricRecord) pmetric.NumberDataPointSlice {
	dataPoints := pmetric.NewNumberDataPointSlice()
	for _, point := range rec.Datapoints {
		if point.Timestamp == 0 {
			b.logger.Warn(
				"Skipping OCI metric datapoint with zero timestamp",
				zap.String("name", rec.Name),
				zap.String("namespace", rec.Namespace),
			)
			continue
		}

		timestamp := time.UnixMilli(point.Timestamp)

		dp := dataPoints.AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(timestamp))
		dp.SetDoubleValue(point.Value)
		if len(rec.Dimensions) > 0 {
			if err := dp.Attributes().FromRaw(rec.Dimensions); err != nil {
				b.logger.Warn(
					"Failed to set attributes from dimensions",
					zap.Any("dimensions", rec.Dimensions),
					zap.Error(err),
				)
			}
		}
	}
	return dataPoints
}
