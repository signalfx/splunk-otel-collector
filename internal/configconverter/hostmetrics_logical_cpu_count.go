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

package configconverter

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

const (
	hostMetricsReceiverType      = "host_metrics"
	hostMetricsReceiverAliasType = "hostmetrics"
	signalfxExporterType         = "signalfx"
	metricsPipelineType          = "metrics"
	cpuScraperName               = "cpu"
	logicalCPUCountMetricName    = "system.cpu.logical.count"
)

// IncludeHostMetricsLogicalCPUCount adds system.cpu.logical.count to SignalFx
// exporter includes when users explicitly enable it on host_metrics receivers.
func IncludeHostMetricsLogicalCPUCount(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return nil
	}

	out := cfgMap.ToStringMap()

	service, ok, err := getMap(out, "service")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	pipelines, ok, err := getMap(service, "pipelines")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	receivers, ok, err := getMap(out, "receivers")
	if err != nil || !ok {
		return err
	}

	exporters, _, err := getMap(out, "exporters")
	if err != nil {
		return err
	}

	targetExporters, err := signalfxExportersForEnabledLogicalCPUCount(pipelines, receivers, exporters)
	if err != nil {
		return err
	}

	changed := false
	for exporterID := range targetExporters {
		exporterChanged, err := includeLogicalCPUCountMetric(exporters, exporterID)
		if err != nil {
			return err
		}
		if exporterChanged {
			changed = true
		}
	}
	if !changed {
		return nil
	}

	*cfgMap = *confmap.NewFromStringMap(out)
	return nil
}

func signalfxExportersForEnabledLogicalCPUCount(
	pipelines, receivers, exporters map[string]any,
) (map[string]struct{}, error) {
	targetExporters := map[string]struct{}{}
	for pipelineID, rawPipeline := range pipelines {
		if !isMetricsPipeline(pipelineID) {
			continue
		}

		pipeline, ok := rawPipeline.(map[string]any)
		if !ok {
			if rawPipeline == nil {
				continue
			}
			return nil, fmt.Errorf("service::pipelines::%s is of unexpected form (%T): %v", pipelineID, rawPipeline, rawPipeline)
		}

		pipelineReceivers, err := componentIDs(pipeline, "receivers")
		if err != nil {
			return nil, fmt.Errorf("cannot determine service::pipelines::%s::receivers: %w", pipelineID, err)
		}
		if !usesHostMetricsLogicalCPUCount(pipelineReceivers, receivers) {
			continue
		}

		pipelineExporters, err := componentIDs(pipeline, "exporters")
		if err != nil {
			return nil, fmt.Errorf("cannot determine service::pipelines::%s::exporters: %w", pipelineID, err)
		}
		for _, exporterID := range pipelineExporters {
			if componentType(exporterID) == signalfxExporterType && signalfxDefaultTranslationRulesEnabled(exporterID, exporters) {
				targetExporters[exporterID] = struct{}{}
			}
		}
	}
	return targetExporters, nil
}

func getMap(parent map[string]any, key string) (map[string]any, bool, error) {
	raw, ok := parent[key]
	if !ok || raw == nil {
		return nil, false, nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s is of unexpected form (%T): %v", key, raw, raw)
	}
	return m, true, nil
}

func componentIDs(parent map[string]any, key string) ([]string, error) {
	raw, ok := parent[key]
	if !ok || raw == nil {
		return nil, nil
	}
	rawComponents, err := toAnySlice(raw)
	if err != nil {
		return nil, err
	}

	components := make([]string, 0, len(rawComponents))
	for _, rawComponent := range rawComponents {
		component, ok := rawComponent.(string)
		if !ok {
			continue
		}
		components = append(components, component)
	}
	return components, nil
}

func usesHostMetricsLogicalCPUCount(receiverIDs []string, receivers map[string]any) bool {
	for _, receiverID := range receiverIDs {
		if isHostMetricsReceiver(receiverID) && hostMetricsLogicalCPUCountEnabled(receivers, receiverID) {
			return true
		}
	}
	return false
}

func signalfxDefaultTranslationRulesEnabled(exporterID string, exporters map[string]any) bool {
	rawExporter, ok := exporters[exporterID]
	if !ok || rawExporter == nil {
		return true
	}

	exporter, ok := rawExporter.(map[string]any)
	if !ok {
		return true
	}

	disableDefaultTranslationRules, ok := exporter["disable_default_translation_rules"].(bool)
	return !ok || !disableDefaultTranslationRules
}

func hostMetricsLogicalCPUCountEnabled(receivers map[string]any, receiverID string) bool {
	rawReceiver, ok := receivers[receiverID]
	if !ok {
		return false
	}

	receiver := map[string]any{}
	if rawReceiver != nil {
		var receiverOK bool
		receiver, receiverOK = rawReceiver.(map[string]any)
		if !receiverOK {
			return false
		}
	}

	rawScrapers, ok := receiver["scrapers"]
	if !ok || rawScrapers == nil {
		return false
	}
	scrapers, ok := rawScrapers.(map[string]any)
	if !ok {
		return false
	}

	rawCPU, ok := scrapers[cpuScraperName]
	if !ok {
		return false
	}

	cpu := map[string]any{}
	if rawCPU != nil {
		var cpuOK bool
		cpu, cpuOK = rawCPU.(map[string]any)
		if !cpuOK {
			return false
		}
	}

	metrics, ok := nestedMap(cpu, "metrics")
	if !ok {
		return false
	}
	metric, ok := nestedMap(metrics, logicalCPUCountMetricName)
	if !ok {
		return false
	}

	enabled, ok := metric["enabled"].(bool)
	return ok && enabled
}

func includeLogicalCPUCountMetric(exporters map[string]any, exporterID string) (bool, error) {
	rawExporter, ok := exporters[exporterID]
	if !ok {
		return false, nil
	}

	exporter := map[string]any{}
	if rawExporter != nil {
		var exporterOK bool
		exporter, exporterOK = rawExporter.(map[string]any)
		if !exporterOK {
			return false, nil
		}
	}

	included, err := metricFilterListContains(exporter, "include_metrics")
	if err != nil {
		return false, fmt.Errorf("cannot determine exporters::%s::include_metrics: %w", exporterID, err)
	}
	excluded, err := metricFilterListContains(exporter, "exclude_metrics")
	if err != nil {
		return false, fmt.Errorf("cannot determine exporters::%s::exclude_metrics: %w", exporterID, err)
	}
	if included || excluded {
		return false, nil
	}

	includeMetrics, err := metricFilterList(exporter["include_metrics"])
	if err != nil {
		return false, fmt.Errorf("cannot determine exporters::%s::include_metrics: %w", exporterID, err)
	}
	includeMetrics = append(includeMetrics, map[string]any{"metric_name": logicalCPUCountMetricName})
	exporter["include_metrics"] = includeMetrics
	exporters[exporterID] = exporter
	return true, nil
}

func metricFilterListContains(parent map[string]any, key string) (bool, error) {
	filters, err := metricFilterList(parent[key])
	if err != nil {
		return false, err
	}
	for i, rawFilter := range filters {
		filter, ok := rawFilter.(map[string]any)
		if !ok {
			if rawFilter == nil {
				continue
			}
			return false, fmt.Errorf("%d is of unexpected form (%T): %v", i, rawFilter, rawFilter)
		}
		if filter["metric_name"] == logicalCPUCountMetricName {
			return true, nil
		}
		rawMetricNames, ok := filter["metric_names"]
		if !ok || rawMetricNames == nil {
			continue
		}
		metricNames, err := toAnySlice(rawMetricNames)
		if err != nil {
			return false, fmt.Errorf("%d::metric_names is of unexpected form: %w", i, err)
		}
		for _, rawMetricName := range metricNames {
			if rawMetricName == logicalCPUCountMetricName {
				return true, nil
			}
		}
	}
	return false, nil
}

func metricFilterList(raw any) ([]any, error) {
	if raw == nil {
		return nil, nil
	}
	return toAnySlice(raw)
}

func nestedMap(parent map[string]any, key string) (map[string]any, bool) {
	raw, ok := parent[key]
	if !ok || raw == nil {
		return map[string]any{}, true
	}

	m, ok := raw.(map[string]any)
	return m, ok
}

func isMetricsPipeline(id string) bool {
	return componentType(id) == metricsPipelineType
}

func isHostMetricsReceiver(id string) bool {
	switch componentType(id) {
	case hostMetricsReceiverType, hostMetricsReceiverAliasType:
		return true
	default:
		return false
	}
}

func componentType(id string) string {
	componentType, _, _ := strings.Cut(id, "/")
	return componentType
}
