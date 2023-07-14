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

package spark

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

type Service interface {
	getSparkMetricsForClusters(clusters []Cluster) (map[Cluster]ClusterMetrics, error)
	getSparkMetricsForCluster(clusterID string) (ClusterMetrics, error)
	getSparkExecutorInfoSliceByApp(clusterID string) (map[Application][]ExecutorInfo, error)
	getSparkJobInfoSliceByApp(clusterID string) (map[Application][]JobInfo, error)
	getSparkStageInfoSliceByApp(clusterID string) (map[Application][]StageInfo, error)
}

type restService struct {
	logger      *zap.Logger
	sparkClient client
}

func NewService(
	logger *zap.Logger,
	httpDoer httpauth.HTTPDoer,
	tok string,
	sparkEndpoint string,
	orgID string,
	sparkUIPort int,
) Service {
	return restService{
		logger:      logger,
		sparkClient: newClient(httpDoer, tok, sparkEndpoint, orgID, sparkUIPort),
	}
}

func (s restService) getSparkMetricsForClusters(clusters []Cluster) (map[Cluster]ClusterMetrics, error) {
	out := map[Cluster]ClusterMetrics{}
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

func (s restService) getSparkMetricsForCluster(clusterID string) (ClusterMetrics, error) {
	return s.sparkClient.metrics(clusterID)
}

func (s restService) getSparkExecutorInfoSliceByApp(clusterID string) (map[Application][]ExecutorInfo, error) {
	out := map[Application][]ExecutorInfo{}
	apps, err := s.sparkClient.applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		executors, err := s.sparkClient.appExecutors(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get executors for app id: %s: %w", app.ID, err)
		}
		out[app] = executors
	}
	return out, nil
}

func (s restService) getSparkJobInfoSliceByApp(clusterID string) (map[Application][]JobInfo, error) {
	out := map[Application][]JobInfo{}
	apps, err := s.sparkClient.applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		jobs, err := s.sparkClient.appJobs(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.ID, err)
		}
		out[app] = jobs
	}
	return out, nil
}

func (s restService) getSparkStageInfoSliceByApp(clusterID string) (map[Application][]StageInfo, error) {
	out := map[Application][]StageInfo{}
	apps, err := s.sparkClient.applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		stages, err := s.sparkClient.appStages(clusterID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.ID, err)
		}
		out[app] = stages
	}
	return out, nil
}
