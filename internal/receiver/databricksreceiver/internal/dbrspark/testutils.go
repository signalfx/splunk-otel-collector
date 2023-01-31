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

package dbrspark

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/httpauth"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func NewTestSuccessSparkService(testdataDir string) Service {
	return newTestSparkService(testdataDir, "")
}

func NewTestForbiddenSparkService(testdataDir, forbiddenClusterID string) Service {
	return newTestSparkService(testdataDir, forbiddenClusterID)
}

func newTestSparkService(testdataDir, forbiddenClusterID string) restService {
	return restService{
		logger: zap.New(zapcore.NewNopCore()),
		sparkClient: spark.Client{RawClient: &testDataSparkRawClient{
			testdataDir:        testdataDir,
			forbiddenClusterID: forbiddenClusterID,
		}},
	}
}

type testDataSparkRawClient struct {
	testdataDir        string
	forbiddenClusterID string
}

func (c *testDataSparkRawClient) Metrics(clusterID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join(c.testdataDir, "spark", "metrics.json"))
}

func (c *testDataSparkRawClient) Applications(clusterID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join(c.testdataDir, "spark", "applications.json"))
}

func (c *testDataSparkRawClient) AppExecutors(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join(c.testdataDir, "spark", "executors.json"))
}

func (c *testDataSparkRawClient) AppJobs(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join(c.testdataDir, "spark", "jobs.json"))
}

func (c *testDataSparkRawClient) AppStages(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join(c.testdataDir, "spark", "stages.json"))
}
