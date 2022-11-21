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
	"net/http"

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

type sparkService interface {
	getSparkMetricsForClusters(clusters []cluster) (map[cluster]spark.ClusterMetrics, error)
	getSparkMetricsForCluster(clusterID string) (spark.ClusterMetrics, error)
	getSparkExecutorInfoSliceByApp(clusterID string) (map[spark.Application][]spark.ExecutorInfo, error)
	getSparkJobInfoSliceByApp(clusterID string) (map[spark.Application][]spark.JobInfo, error)
	getSparkStageInfoSliceByApp(clusterID string) (map[spark.Application][]spark.StageInfo, error)
}

type sparkRestService struct {
	logger      *zap.Logger
	dbsvc       databricksService
	sparkClient spark.Client
}

func newSparkService(
	logger *zap.Logger,
	dbsvc databricksService,
	httpClient *http.Client,
	tok string,
	sparkAPIURL string,
	orgID string,
	sparkUIPort int,
) sparkService {
	return sparkRestService{
		logger:      logger,
		dbsvc:       dbsvc,
		sparkClient: spark.NewClient(httpClient, tok, sparkAPIURL, orgID, sparkUIPort),
	}
}

func (s sparkRestService) getSparkMetricsForClusters(clusters []cluster) (map[cluster]spark.ClusterMetrics, error) {
	out := map[cluster]spark.ClusterMetrics{}
	for _, clstr := range clusters {
		metrics, err := s.getSparkMetricsForCluster(clstr.ClusterID)
		if err != nil {
			if httpauth.IsForbidden(err) {
				s.logger.Warn(
					"not authorized to get metrics for cluster, skipping",
					zap.String("cluster name", clstr.ClusterName),
					zap.String("cluster id", clstr.ClusterID),
				)
				continue
			}
			return nil, fmt.Errorf("error getting spark metrics for cluster: %s: %w", clstr, err)
		}
		out[clstr] = metrics
	}
	return out, nil
}

func (s sparkRestService) getSparkMetricsForCluster(clusterID string) (spark.ClusterMetrics, error) {
	return s.sparkClient.Metrics(clusterID)
}

func (s sparkRestService) getSparkExecutorInfoSliceByApp(clusterID string) (map[spark.Application][]spark.ExecutorInfo, error) {
	out := map[spark.Application][]spark.ExecutorInfo{}
	apps, err := s.sparkClient.Applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		executors, err := s.sparkClient.AppExecutors(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get executors for app id: %s: %w", app.ID, err)
		}
		out[app] = executors
	}
	return out, nil
}

func (s sparkRestService) getSparkJobInfoSliceByApp(clusterID string) (map[spark.Application][]spark.JobInfo, error) {
	out := map[spark.Application][]spark.JobInfo{}
	apps, err := s.sparkClient.Applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		jobs, err := s.sparkClient.AppJobs(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.ID, err)
		}
		out[app] = jobs
	}
	return out, nil
}

func (s sparkRestService) getSparkStageInfoSliceByApp(clusterID string) (map[spark.Application][]spark.StageInfo, error) {
	out := map[spark.Application][]spark.StageInfo{}
	apps, err := s.sparkClient.Applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		stages, err := s.sparkClient.AppStages(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.ID, err)
		}
		out[app] = stages
	}
	return out, nil
}
