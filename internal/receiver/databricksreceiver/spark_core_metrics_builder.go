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

package databricksreceiver

import (
	"fmt"
	"strings"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

type sparkClusterMetricsBuilder struct {
	ssvc sparkService
}

func (b sparkClusterMetricsBuilder) buildCoreMetrics(clusters []cluster, pipelines []pipelineSummary) (*sparkDbrMetrics, error) {
	out := newSparkDbrMetrics()

	sparkMetricsForCluster, err := b.ssvc.getSparkMetricsForClusters(clusters)
	if err != nil {
		return nil, fmt.Errorf("error getting spark metrics for clusters: %w", err)
	}

	pipelinesByClusterID := map[string]pipelineSummary{}
	for _, pipeline := range pipelines {
		pipelinesByClusterID[pipeline.clusterID] = pipeline
	}

	for clstr, sparkMetrics := range sparkMetricsForCluster {
		pipeline, ok := pipelinesByClusterID[clstr.ClusterID]
		var pipelineRef *pipelineSummary
		if ok {
			pipelineRef = &pipeline
		}

		cm := b.clusterMetrics(sparkMetrics, pipelineRef, clstr)
		out.append(cm)

		tm := b.timerMetrics(sparkMetrics, clstr, pipelineRef)
		out.append(tm)

		hm := b.histoMetrics(sparkMetrics, clstr, pipelineRef)
		out.append(hm)
	}
	return out, nil
}

func (b sparkClusterMetricsBuilder) histoMetrics(m spark.ClusterMetrics, clstr cluster, pipeline *pipelineSummary) *sparkDbrMetrics {
	out := newSparkDbrMetrics()
	for key, sparkHisto := range m.Histograms {
		appID, partialMetricName := stripSparkMetricKey(key)
		out.addHisto(clstr, appID, sparkHisto, newSparkMetricBase(partialMetricName, pipeline))
	}
	return out
}

func (b sparkClusterMetricsBuilder) clusterMetrics(m spark.ClusterMetrics, pipeline *pipelineSummary, clstr cluster) *sparkDbrMetrics {
	out := newSparkDbrMetrics()
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

func (b sparkClusterMetricsBuilder) timerMetrics(m spark.ClusterMetrics, cluster cluster, pipeline *pipelineSummary) *sparkDbrMetrics {
	out := newSparkDbrMetrics()
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
