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

package configconverter

import (
	"context"
	"fmt"
	"regexp"

	sfxpb "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/configconverter/dpfilters"
)

// Metrics deprecated by kubeletstats receiver that need to be disabled if not explicitly included in signalfx exporter.
const (
	k8sNodeCPUUtilization   = "k8s.node.cpu.utilization"
	k8sPodCPUUtilization    = "k8s.pod.cpu.utilization"
	containerCPUUtilization = "container.cpu.utilization"
)

// signalfxExporterConfig is the configuration for the signalfx exporter that contains the metrics filters.
type signalfxExporterConfig struct {
	ExcludeMetrics []dpfilters.MetricFilter `mapstructure:"exclude_metrics"`
	IncludeMetrics []dpfilters.MetricFilter `mapstructure:"include_metrics"`
}

// DisableKubeletUtilizationMetrics is a MapConverter that disables the following deprecated metrics:
// - `k8s.node.cpu.utilization`
// - `k8s.pod.cpu.utilization`
// - `container.cpu.utilization`
// The converter disables the metrics at the receiver level to avoid showing users a warning message because
// they are excluded in signalfx exporter by default.
// We don't disable them in case if users explicitly include them in signalfx exporter.
type DisableKubeletUtilizationMetrics struct{}

func (DisableKubeletUtilizationMetrics) Convert(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return fmt.Errorf("cannot DisableKubeletUtilizationMetrics on nil *confmap.Conf")
	}

	receivers, err := cfgMap.Sub("receivers")
	if err != nil {
		return nil // Ignore invalid config. Rely on the config validation to catch this.
	}
	kubeletReceiverConfigs := map[string]map[string]any{}
	for receiverName, receiverCfg := range receivers.ToStringMap() {
		if regexp.MustCompile(`kubeletstats(/\w+)?`).MatchString(receiverName) {
			if v, ok := receiverCfg.(map[string]any); ok {
				kubeletReceiverConfigs[receiverName] = v
			}
		}
	}

	exporters, err := cfgMap.Sub("exporters")
	if err != nil {
		return nil // Ignore invalid config. Rely on the config validation to catch this.
	}
	sfxExporterConfigs := map[string]map[string]any{}
	for exporterName, exporterCfg := range exporters.ToStringMap() {
		if regexp.MustCompile(`signalfx(/\w+)?`).MatchString(exporterName) {
			if v, ok := exporterCfg.(map[string]any); ok {
				sfxExporterConfigs[exporterName] = v
			}
		}
	}

	// If there is no signalfx exporter or kubeletstats receiver, there is nothing to do.
	if len(kubeletReceiverConfigs) == 0 || len(sfxExporterConfigs) == 0 {
		return nil
	}

	disableMetrics := map[string]bool{
		k8sNodeCPUUtilization:   true,
		k8sPodCPUUtilization:    true,
		containerCPUUtilization: true,
	}

	// Check if the metrics are explicitly included in signalfx exporter.
	// If they are not included, we will disable them in kubeletstats receiver.
	for _, cm := range sfxExporterConfigs {
		cfg := signalfxExporterConfig{}
		err = confmap.NewFromStringMap(cm).Unmarshal(&cfg, confmap.WithIgnoreUnused())
		if err != nil {
			return nil // Ignore invalid config. Rely on the config validation to catch this.
		}
		if len(cfg.ExcludeMetrics) == 0 {
			// Apply default excluded metrics if not explicitly set.
			// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/f2d8efe507083b0f38b6567f8dba3f37053bfa86/exporter/signalfxexporter/internal/translation/default_metrics.go#L133
			cfg.ExcludeMetrics = []dpfilters.MetricFilter{
				{MetricNames: []string{"/^(?i:(container)|(k8s\\.node)|(k8s\\.pod))\\.cpu\\.utilization$/"}},
			}
		}

		var filter *dpfilters.FilterSet
		filter, err = dpfilters.NewFilterSet(cfg.ExcludeMetrics, cfg.IncludeMetrics)
		if err != nil {
			return nil // Ignore invalid config. Rely on the config validation to catch this.
		}
		for metricName := range disableMetrics {
			if !filter.Matches(&sfxpb.DataPoint{Metric: metricName}) {
				disableMetrics[metricName] = false
			}
		}
	}

	// Disable the metrics in kubeletstats receiver.
	for receiverName, cfg := range kubeletReceiverConfigs {
		metricsCfg := map[string]any{}
		if cfg["metrics"] != nil {
			if v, ok := cfg["metrics"].(map[string]any); ok {
				metricsCfg = v
			}
		}
		if _, ok := metricsCfg[k8sNodeCPUUtilization]; !ok && disableMetrics[k8sNodeCPUUtilization] {
			metricsCfg[k8sNodeCPUUtilization] = map[string]any{"enabled": false}
		}
		if _, ok := metricsCfg[k8sPodCPUUtilization]; !ok && disableMetrics[k8sPodCPUUtilization] {
			metricsCfg[k8sPodCPUUtilization] = map[string]any{"enabled": false}
		}
		if _, ok := metricsCfg[containerCPUUtilization]; !ok && disableMetrics[containerCPUUtilization] {
			metricsCfg[containerCPUUtilization] = map[string]any{"enabled": false}
		}
		metricsCfgKey := fmt.Sprintf("receivers::%s::metrics", receiverName)
		if len(metricsCfg) > 0 {
			if err = cfgMap.Merge(confmap.NewFromStringMap(map[string]any{metricsCfgKey: metricsCfg})); err != nil {
				return err
			}
		}
	}

	return nil
}
