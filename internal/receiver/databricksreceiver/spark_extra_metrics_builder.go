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

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

type sparkExtraMetricsBuilder struct {
	ssvc   sparkService
	logger *zap.Logger
}

func (b sparkExtraMetricsBuilder) buildExecutorMetrics(clusters []cluster) (*sparkDbrMetrics, error) {
	out := newSparkDbrMetrics()
	for _, clstr := range clusters {
		execInfosByApp, err := b.ssvc.getSparkExecutorInfoSliceByApp(clstr.ClusterID)
		if err != nil {
			if httpauth.IsForbidden(err) {
				b.logger.Warn(
					"not authorized to get executor info for cluster, skipping",
					zap.String("cluster name", clstr.ClusterName),
					zap.String("cluster id", clstr.ClusterID),
				)
				continue
			}
			return nil, fmt.Errorf("failed to get spark executor info for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, execInfos := range execInfosByApp {
			for _, ei := range execInfos {
				out.addExecInfo(clstr, sparkApp.ID, ei)
			}
		}
	}
	return out, nil
}

func (b sparkExtraMetricsBuilder) buildJobMetrics(clusters []cluster) (*sparkDbrMetrics, error) {
	out := newSparkDbrMetrics()
	for _, clstr := range clusters {
		jobInfosByApp, err := b.ssvc.getSparkJobInfoSliceByApp(clstr.ClusterID)
		if err != nil {
			if httpauth.IsForbidden(err) {
				b.logger.Warn(
					"not authorized to get spark job info for cluster, skipping",
					zap.String("cluster name", clstr.ClusterName),
					zap.String("cluster id", clstr.ClusterID),
				)
				continue
			}
			return nil, fmt.Errorf("failed to get jobs for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, jobInfos := range jobInfosByApp {
			for _, ji := range jobInfos {
				out.addJobInfos(clstr, sparkApp.ID, ji)
			}
		}
	}
	return out, nil
}

func (b sparkExtraMetricsBuilder) buildStageMetrics(clusters []cluster) (*sparkDbrMetrics, error) {
	out := newSparkDbrMetrics()
	for _, clstr := range clusters {
		stageInfosByApp, err := b.ssvc.getSparkStageInfoSliceByApp(clstr.ClusterID)
		if err != nil {
			if httpauth.IsForbidden(err) {
				b.logger.Warn(
					"not authorized to get spark stage info for cluster, skipping",
					zap.String("cluster name", clstr.ClusterName),
					zap.String("cluster id", clstr.ClusterID),
				)
				continue
			}
			return nil, fmt.Errorf("failed to get stages for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, stageInfos := range stageInfosByApp {
			for _, si := range stageInfos {
				out.addStageInfo(clstr, sparkApp.ID, si)
			}
		}
	}
	return out, nil
}
