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
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func newTestSuccessSparkService(dbsvc databricksService) sparkRestService {
	return newTestSparkService(dbsvc, false)
}

func newTestForbiddenSparkService(dbsvc databricksService) sparkRestService {
	return newTestSparkService(dbsvc, true)
}

func newTestSparkService(dbsvc databricksService, metricsForbidden bool) sparkRestService {
	return sparkRestService{
		logger: zap.New(zapcore.NewNopCore()),
		dbsvc:  dbsvc,
		sparkClient: spark.Client{RawClient: &testdataSparkRawClient{
			metricsForbidden: metricsForbidden,
		}},
	}
}

type testdataSparkRawClient struct {
	metricsForbidden        bool
	testMetricsAPICallCount int
}

func (c *testdataSparkRawClient) Metrics(clusterID string) ([]byte, error) {
	c.testMetricsAPICallCount++
	if c.metricsForbidden && c.testMetricsAPICallCount == 1 {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "metrics.json"))
}

func (c *testdataSparkRawClient) Applications(clusterID string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", "spark", "applications.json"))
}

func (c *testdataSparkRawClient) AppExecutors(clusterID, appID string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", "spark", "executors.json"))
}

func (c *testdataSparkRawClient) AppJobs(clusterID, appID string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", "spark", "jobs.json"))
}

func (c *testdataSparkRawClient) AppStages(clusterID, appID string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", "spark", "stages.json"))
}
