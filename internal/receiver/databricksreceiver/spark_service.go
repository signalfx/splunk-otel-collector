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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

type sparkService struct {
	logger             *zap.Logger
	dbsvc              databricksServiceIntf
	httpClient         *http.Client
	sparkAPIEndpoint   string
	sparkUIPort        int
	orgID              string
	tok                string
	unmarshalerFactory sparkUnmarshalerFactory
}

type sparkUnmarshalerFactory func(
	logger *zap.Logger,
	httpClient *http.Client,
	sparkProxyURL string,
	orgID string,
	port int,
	token string,
	clusterID string,
) spark.Unmarshaler

func newSparkService(
	logger *zap.Logger,
	dbsvc databricksServiceIntf,
	httpClient *http.Client,
	sparkAPIEndpoint string,
	sparkUIPort int,
	orgID string,
	tok string,
	factory sparkUnmarshalerFactory,
) sparkService {
	return sparkService{
		logger:             logger,
		dbsvc:              dbsvc,
		httpClient:         httpClient,
		sparkAPIEndpoint:   sparkAPIEndpoint,
		sparkUIPort:        sparkUIPort,
		orgID:              orgID,
		tok:                tok,
		unmarshalerFactory: factory,
	}
}

func (s sparkService) getSparkCoreMetricsForAllClusters() (map[cluster]spark.ClusterMetrics, error) {
	clusters, err := s.dbsvc.runningClusters()
	if err != nil {
		return nil, fmt.Errorf("error getting cluster IDs: %w", err)
	}
	out := map[cluster]spark.ClusterMetrics{}
	for _, clstr := range clusters {
		metrics, err := s.getSparkCoreMetricsForCluster(clstr.ClusterId)
		if err != nil {
			return nil, fmt.Errorf("error getting spark metrics for cluster: %s: %w", clstr, err)
		}
		out[clstr] = metrics
	}
	return out, nil
}

func (s sparkService) getSparkCoreMetricsForCluster(clusterID string) (spark.ClusterMetrics, error) {
	return s.newUnmarshaler(clusterID).Metrics()
}

func (s sparkService) getSparkExecutorInfoSliceByApp(clusterID string) (map[spark.Application][]spark.ExecutorInfo, error) {
	out := map[spark.Application][]spark.ExecutorInfo{}
	unm := s.newUnmarshaler(clusterID)
	apps, err := unm.Applications()
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		executors, err := unm.AppExecutors(app.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get executors for app id: %s: %w", app.Id, err)
		}
		out[app] = executors
	}
	return out, nil
}

func (s sparkService) getSparkJobInfoSliceByApp(clusterID string) (map[spark.Application][]spark.JobInfo, error) {
	out := map[spark.Application][]spark.JobInfo{}
	unm := s.newUnmarshaler(clusterID)
	apps, err := unm.Applications()
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		jobs, err := unm.AppJobs(app.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.Id, err)
		}
		out[app] = jobs
	}
	return out, nil
}

func (s sparkService) getSparkStageInfoSliceByApp(clusterID string) (map[spark.Application][]spark.StageInfo, error) {
	out := map[spark.Application][]spark.StageInfo{}
	unm := s.newUnmarshaler(clusterID)
	apps, err := unm.Applications()
	if err != nil {
		return nil, fmt.Errorf("failed to get applications from spark: %w", err)
	}
	for _, app := range apps {
		stages, err := unm.AppStages(app.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs for app id: %s: %w", app.Id, err)
		}
		out[app] = stages
	}
	return out, nil
}

func (s sparkService) newUnmarshaler(clusterID string) spark.Unmarshaler {
	return s.unmarshalerFactory(s.logger, s.httpClient, s.sparkAPIEndpoint, s.orgID, s.sparkUIPort, s.tok, clusterID)
}
