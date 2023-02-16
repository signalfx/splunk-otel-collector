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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
)

// more info on these endpoints here: https://spark.apache.org/docs/latest/monitoring.html#rest-api
const (
	metricsPath         = "/metrics/json"
	applicationsPath    = "/api/v1/applications"
	appExecutorsPathFmt = applicationsPath + "/%s/executors"
	appJobsPathFmt      = applicationsPath + "/%s/jobs"
	appStagesPathFmt    = applicationsPath + "/%s/stages"
)

// rawClient defines the methods that can be called against the Spark API. Its
// return values are byte arrays to be unmarshalled by the caller.
type rawClient interface {
	metrics(clusterID string) ([]byte, error)
	applications(clusterID string) ([]byte, error)
	appExecutors(clusterID, appID string) ([]byte, error)
	appJobs(clusterID, appID string) ([]byte, error)
	appStages(clusterID, appID string) ([]byte, error)
}

// rawHTTPClient implements rawClient.
type rawHTTPClient struct {
	authClient     httpauth.Client
	baseURLPattern string
}

func newRawHTTPClient(authClient httpauth.Client, sparkAPIURL string, orgID string, port int) *rawHTTPClient {
	return &rawHTTPClient{
		authClient:     authClient,
		baseURLPattern: fmt.Sprintf("%s/driver-proxy-api/o/%s/%%s/%d", sparkAPIURL, orgID, port),
	}
}

func (c rawHTTPClient) metrics(clusterID string) ([]byte, error) {
	return c.authClient.Get(c.baseURL(clusterID) + metricsPath)
}

func (c rawHTTPClient) applications(clusterID string) ([]byte, error) {
	return c.authClient.Get(c.baseURL(clusterID) + applicationsPath)
}

func (c rawHTTPClient) appExecutors(clusterID, appID string) ([]byte, error) {
	return c.authClient.Get(c.baseURL(clusterID) + fmt.Sprintf(appExecutorsPathFmt, appID))
}

func (c rawHTTPClient) appJobs(clusterID, appID string) ([]byte, error) {
	return c.authClient.Get(c.baseURL(clusterID) + fmt.Sprintf(appJobsPathFmt, appID))
}

func (c rawHTTPClient) appStages(clusterID, appID string) ([]byte, error) {
	return c.authClient.Get(c.baseURL(clusterID) + fmt.Sprintf(appStagesPathFmt, appID))
}

func (c rawHTTPClient) baseURL(clusterID string) string {
	return fmt.Sprintf(c.baseURLPattern, clusterID)
}
