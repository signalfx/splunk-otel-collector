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
	"encoding/json"
	"fmt"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

// newClient creates a client, which is a REST client for a Spark API running
// behind a Databricks proxy. It gets JSON bytes from its rawClient and
// unmarshalls content into go types. Although it's focused on standalone Spark,
// this document has some additional information:
// https://spark.apache.org/docs/latest/monitoring.html
func newClient(httpDoer httpauth.HTTPDoer, tok, sparkEndpoint, orgID string, port int) client {
	return client{
		rawClient: newRawHTTPClient(httpauth.NewClient(httpDoer, tok, nil), sparkEndpoint, orgID, port),
	}
}

type client struct {
	rawClient rawClient
}

func (c client) metrics(clusterID string) (ClusterMetrics, error) {
	cm := ClusterMetrics{}
	bytes, err := c.rawClient.metrics(clusterID)
	if err != nil {
		return cm, fmt.Errorf("metrics failed to get metrics from spark: %w", err)
	}
	err = json.Unmarshal(bytes, &cm)
	if err != nil {
		return cm, fmt.Errorf("metricsfailed to unmarshal spark metrics: %w", err)
	}
	return cm, nil
}

func (c client) applications(clusterID string) ([]Application, error) {
	var apps []Application
	bytes, err := c.rawClient.applications(clusterID)
	if err != nil {
		return nil, fmt.Errorf("applications failed to get applications from spark: %w", err)
	}
	err = json.Unmarshal(bytes, &apps)
	if err != nil {
		return nil, fmt.Errorf("applications failed to unmarshal spark applications: %w", err)
	}
	return apps, nil
}

func (c client) appExecutors(clusterID, appID string) ([]ExecutorInfo, error) {
	bytes, err := c.rawClient.appExecutors(clusterID, appID)
	if err != nil {
		return nil, fmt.Errorf("appExecutors failed to get app executors from spark: %w", err)
	}
	var ei []ExecutorInfo
	err = json.Unmarshal(bytes, &ei)
	if err != nil {
		return nil, fmt.Errorf("appExecutors failed to unmarshal executor info: %w", err)
	}
	return ei, nil
}

func (c client) appJobs(clusterID, appID string) ([]JobInfo, error) {
	bytes, err := c.rawClient.appJobs(clusterID, appID)
	if err != nil {
		return nil, fmt.Errorf("appJobs failed to get jobs from spark: %w", err)
	}
	var jobs []JobInfo
	err = json.Unmarshal(bytes, &jobs)
	if err != nil {
		return nil, fmt.Errorf("appJobs failed to unmarshal job info: %w", err)
	}
	return jobs, nil
}

func (c client) appStages(clusterID, appID string) ([]StageInfo, error) {
	bytes, err := c.rawClient.appStages(clusterID, appID)
	if err != nil {
		return nil, fmt.Errorf("appStages failed to get jobs from spark: %w", err)
	}
	var stages []StageInfo
	err = json.Unmarshal(bytes, &stages)
	if err != nil {
		return nil, fmt.Errorf("appStages failed to unmarshal job info: %w", err)
	}
	return stages, nil
}
