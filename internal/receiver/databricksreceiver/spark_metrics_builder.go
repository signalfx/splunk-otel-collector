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

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

type sparkExtraMetricsBuilder struct {
	ssvc   sparkService
	logger *zap.Logger
}

func (b sparkExtraMetricsBuilder) buildExecutorMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusters []cluster) error {
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
			return fmt.Errorf("failed to get spark executor info for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, execInfos := range execInfosByApp {
			for _, execInfo := range execInfos {
				builder.RecordDatabricksSparkExecutorMemoryUsedDataPoint(now, int64(execInfo.MemoryUsed), clstr.ClusterID, sparkApp.ID, execInfo.ID)
				builder.RecordDatabricksSparkExecutorDiskUsedDataPoint(now, int64(execInfo.DiskUsed), clstr.ClusterID, sparkApp.ID, execInfo.ID)
				builder.RecordDatabricksSparkExecutorTotalInputBytesDataPoint(now, execInfo.TotalInputBytes, clstr.ClusterID, sparkApp.ID, execInfo.ID)
				builder.RecordDatabricksSparkExecutorTotalShuffleReadDataPoint(now, int64(execInfo.TotalShuffleRead), clstr.ClusterID, sparkApp.ID, execInfo.ID)
				builder.RecordDatabricksSparkExecutorTotalShuffleWriteDataPoint(now, int64(execInfo.TotalShuffleWrite), clstr.ClusterID, sparkApp.ID, execInfo.ID)
				builder.RecordDatabricksSparkExecutorMaxMemoryDataPoint(now, execInfo.MaxMemory, clstr.ClusterID, sparkApp.ID, execInfo.ID)
			}
		}
	}
	return nil
}

func (b sparkExtraMetricsBuilder) buildJobMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusters []cluster) error {
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
			return fmt.Errorf("failed to get jobs for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, jobInfos := range jobInfosByApp {
			for _, jobInfo := range jobInfos {
				builder.RecordDatabricksSparkJobNumTasksDataPoint(now, int64(jobInfo.NumTasks), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumActiveTasksDataPoint(now, int64(jobInfo.NumActiveTasks), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumCompletedTasksDataPoint(now, int64(jobInfo.NumCompletedTasks), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumSkippedTasksDataPoint(now, int64(jobInfo.NumSkippedTasks), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumFailedTasksDataPoint(now, int64(jobInfo.NumFailedTasks), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumActiveStagesDataPoint(now, int64(jobInfo.NumActiveStages), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumCompletedStagesDataPoint(now, int64(jobInfo.NumCompletedStages), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumSkippedStagesDataPoint(now, int64(jobInfo.NumSkippedStages), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
				builder.RecordDatabricksSparkJobNumFailedStagesDataPoint(now, int64(jobInfo.NumFailedStages), clstr.ClusterID, sparkApp.ID, int64(jobInfo.JobID))
			}
		}
	}
	return nil
}

func (b sparkExtraMetricsBuilder) buildStageMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusters []cluster) error {
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
			return fmt.Errorf("failed to get stages for cluster: %s: %w", clstr.ClusterID, err)
		}
		for sparkApp, stageInfos := range stageInfosByApp {
			for _, stageInfo := range stageInfos {
				builder.RecordDatabricksSparkStageExecutorRunTimeDataPoint(now, int64(stageInfo.ExecutorRunTime), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageInputBytesDataPoint(now, int64(stageInfo.InputBytes), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageInputRecordsDataPoint(now, int64(stageInfo.InputRecords), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageOutputBytesDataPoint(now, int64(stageInfo.OutputBytes), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageOutputRecordsDataPoint(now, int64(stageInfo.OutputRecords), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageMemoryBytesSpilledDataPoint(now, int64(stageInfo.MemoryBytesSpilled), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
				builder.RecordDatabricksSparkStageDiskBytesSpilledDataPoint(now, int64(stageInfo.DiskBytesSpilled), clstr.ClusterID, sparkApp.ID, int64(stageInfo.StageID))
			}
		}
	}
	return nil
}
