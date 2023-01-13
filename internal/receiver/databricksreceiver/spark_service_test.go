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

func newTestSuccessSparkService(dbrsvc databricksService) sparkRestService {
	return newTestSparkService(dbrsvc, "")
}

func newTestForbiddenSparkService(dbrsvc databricksService, forbiddenClusterID string) sparkRestService {
	return newTestSparkService(dbrsvc, forbiddenClusterID)
}

func newTestSparkService(dbrsvc databricksService, forbiddenClusterID string) sparkRestService {
	return sparkRestService{
		logger: zap.New(zapcore.NewNopCore()),
		dbrsvc: dbrsvc,
		sparkClient: spark.Client{RawClient: &testDataSparkRawClient{
			forbiddenClusterID: forbiddenClusterID,
		}},
	}
}

type testDataSparkRawClient struct {
	forbiddenClusterID string
}

func (c *testDataSparkRawClient) Metrics(clusterID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "metrics.json"))
}

func (c *testDataSparkRawClient) Applications(clusterID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "applications.json"))
}

func (c *testDataSparkRawClient) AppExecutors(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "executors.json"))
}

func (c *testDataSparkRawClient) AppJobs(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "jobs.json"))
}

func (c *testDataSparkRawClient) AppStages(clusterID, appID string) ([]byte, error) {
	if clusterID == c.forbiddenClusterID {
		return nil, httpauth.ForbiddenErr()
	}
	return os.ReadFile(filepath.Join("testdata", "spark", "stages.json"))
}
