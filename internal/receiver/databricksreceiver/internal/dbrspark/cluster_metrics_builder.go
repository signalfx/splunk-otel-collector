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

package dbrspark

import (
	"fmt"
	"strings"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

type ClusterMetricsBuilder struct {
	Ssvc Service
}

func (b ClusterMetricsBuilder) BuildCoreMetrics(clusters []spark.Cluster, pipelines []spark.PipelineSummary) (*ResourceMetrics, error) {
	out := NewResourceMetrics()

	sparkMetricsForCluster, err := b.Ssvc.getSparkMetricsForClusters(clusters)
	if err != nil {
		return nil, fmt.Errorf("error getting spark metrics for clusters: %w", err)
	}

	pipelinesByClusterID := map[string]spark.PipelineSummary{}
	for _, pipeline := range pipelines {
		pipelinesByClusterID[pipeline.ClusterID] = pipeline
	}

	for clstr, sparkMetrics := range sparkMetricsForCluster {
		pipeline, ok := pipelinesByClusterID[clstr.ClusterID]
		var pipelineRef *spark.PipelineSummary
		if ok {
			pipelineRef = &pipeline
		}

		cm := b.clusterMetrics(sparkMetrics, pipelineRef, clstr)
		out.Append(cm)

		tm := b.timerMetrics(sparkMetrics, clstr, pipelineRef)
		out.Append(tm)

		hm := b.histoMetrics(sparkMetrics, clstr, pipelineRef)
		out.Append(hm)
	}
	return out, nil
}

func (b ClusterMetricsBuilder) histoMetrics(m spark.ClusterMetrics, clstr spark.Cluster, pipeline *spark.PipelineSummary) *ResourceMetrics {
	out := NewResourceMetrics()
	for key, sparkHisto := range m.Histograms {
		appID, partialMetricName := stripSparkMetricKey(key)
		out.addHisto(clstr, appID, sparkHisto, newSparkMetricBase(partialMetricName, pipeline))
	}
	return out
}

func (b ClusterMetricsBuilder) clusterMetrics(m spark.ClusterMetrics, pipeline *spark.PipelineSummary, clstr spark.Cluster) *ResourceMetrics {
	out := NewResourceMetrics()
	for key, gauge := range m.Gauges {
		appID, partialMetricName := stripSparkMetricKey(key)
		out.addGauge(clstr, appID, gauge, newSparkMetricBase(partialMetricName, pipeline))
	}
	for key, counter := range m.Counters {
		appID, partialMetricName := stripSparkMetricKey(key)
		if appID == "" {
			continue
		}
		out.addCounter(clstr, appID, counter, newSparkMetricBase(partialMetricName, pipeline))
	}
	return out
}

func (b ClusterMetricsBuilder) timerMetrics(m spark.ClusterMetrics, cluster spark.Cluster, pipeline *spark.PipelineSummary) *ResourceMetrics {
	out := NewResourceMetrics()
	for key, timer := range m.Timers {
		appID, partialMetricName := stripSparkMetricKey(key)
		out.addTimer(cluster, appID, timer, newSparkMetricBase(partialMetricName, pipeline))
	}
	return out
}

func stripSparkMetricKey(s string) (string, string) {
	parts := strings.Split(s, ".")
	if len(parts) <= 2 || parts[1] != "driver" {
		return "", ""
	}
	metricParts := parts[2:]
	lastPart := metricParts[len(metricParts)-1]
	if strings.HasSuffix(lastPart, "_MB") {
		metricParts[len(metricParts)-1] = lastPart[:len(lastPart)-3]
	}
	joined := strings.Join(metricParts, ".")
	return parts[0], strings.ToLower(joined)
}
