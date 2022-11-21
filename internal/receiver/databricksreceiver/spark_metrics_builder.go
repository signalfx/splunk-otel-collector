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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

type sparkMetricsBuilder struct {
	ssvc sparkService
}

func (b sparkMetricsBuilder) buildExecutorMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusterIDs []string) error {
	for _, clusterID := range clusterIDs {
		execInfosByApp, err := b.ssvc.getSparkExecutorInfoSliceByApp(clusterID)
		if err != nil {
			return fmt.Errorf("failed to get executor info for cluster: %s: %w", clusterID, err)
		}
		for sparkApp, execInfos := range execInfosByApp {
			for _, execInfo := range execInfos {
				builder.RecordDatabricksSparkExecutorMemoryUsedDataPoint(now, int64(execInfo.MemoryUsed), clusterID, sparkApp.Id, execInfo.Id)
				builder.RecordDatabricksSparkExecutorDiskUsedDataPoint(now, int64(execInfo.DiskUsed), clusterID, sparkApp.Id, execInfo.Id)
				builder.RecordDatabricksSparkExecutorTotalInputBytesDataPoint(now, execInfo.TotalInputBytes, clusterID, sparkApp.Id, execInfo.Id)
				builder.RecordDatabricksSparkExecutorTotalShuffleReadDataPoint(now, int64(execInfo.TotalShuffleRead), clusterID, sparkApp.Id, execInfo.Id)
				builder.RecordDatabricksSparkExecutorTotalShuffleWriteDataPoint(now, int64(execInfo.TotalShuffleWrite), clusterID, sparkApp.Id, execInfo.Id)
				builder.RecordDatabricksSparkExecutorMaxMemoryDataPoint(now, execInfo.MaxMemory, clusterID, sparkApp.Id, execInfo.Id)
			}
		}
	}
	return nil
}

func (b sparkMetricsBuilder) buildJobMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusterIDs []string) error {
	for _, clusterID := range clusterIDs {
		jobInfosByApp, err := b.ssvc.getSparkJobInfoSliceByApp(clusterID)
		if err != nil {
			return fmt.Errorf("failed to get jobs for cluster: %s: %w", clusterID, err)
		}
		for sparkApp, jobInfos := range jobInfosByApp {
			for _, jobInfo := range jobInfos {
				builder.RecordDatabricksSparkJobNumTasksDataPoint(now, int64(jobInfo.NumTasks), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumActiveTasksDataPoint(now, int64(jobInfo.NumActiveTasks), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumCompletedTasksDataPoint(now, int64(jobInfo.NumCompletedTasks), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumSkippedTasksDataPoint(now, int64(jobInfo.NumSkippedTasks), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumFailedTasksDataPoint(now, int64(jobInfo.NumFailedTasks), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumActiveStagesDataPoint(now, int64(jobInfo.NumActiveStages), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumCompletedStagesDataPoint(now, int64(jobInfo.NumCompletedStages), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumSkippedStagesDataPoint(now, int64(jobInfo.NumSkippedStages), clusterID, sparkApp.Id, int64(jobInfo.JobId))
				builder.RecordDatabricksSparkJobNumFailedStagesDataPoint(now, int64(jobInfo.NumFailedStages), clusterID, sparkApp.Id, int64(jobInfo.JobId))
			}
		}
	}
	return nil
}

func (b sparkMetricsBuilder) buildStageMetrics(builder *metadata.MetricsBuilder, now pcommon.Timestamp, clusterIDs []string) error {
	for _, clusterID := range clusterIDs {
		stageInfosByApp, err := b.ssvc.getSparkStageInfoSliceByApp(clusterID)
		if err != nil {
			return fmt.Errorf("failed to get stages for cluster: %s: %w", clusterID, err)
		}
		for sparkApp, stageInfos := range stageInfosByApp {
			for _, stageInfo := range stageInfos {
				builder.RecordDatabricksSparkStageExecutorRunTimeDataPoint(now, int64(stageInfo.ExecutorRunTime), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageInputBytesDataPoint(now, int64(stageInfo.InputBytes), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageInputRecordsDataPoint(now, int64(stageInfo.InputRecords), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageOutputBytesDataPoint(now, int64(stageInfo.OutputBytes), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageOutputRecordsDataPoint(now, int64(stageInfo.OutputRecords), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageMemoryBytesSpilledDataPoint(now, int64(stageInfo.MemoryBytesSpilled), clusterID, sparkApp.Id, int64(stageInfo.StageId))
				builder.RecordDatabricksSparkStageDiskBytesSpilledDataPoint(now, int64(stageInfo.DiskBytesSpilled), clusterID, sparkApp.Id, int64(stageInfo.StageId))
			}
		}
	}
	return nil
}
