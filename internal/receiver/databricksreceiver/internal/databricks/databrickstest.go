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

package databricks

import (
	"fmt"
	"os"
	"path/filepath"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func NewTestDatabricksSingleClusterService(testdataDir string) Service {
	return NewDatabricksService(
		&testdataDBRawClient{testDataDir: testdataDir},
		25,
	)
}

func NewTestDatabricksMultiClusterService(testdataDir string) Service {
	return NewDatabricksService(&testdataDBRawClient{
		testDataDir:  testdataDir,
		multiCluster: true,
	}, 25)
}

// testdataDBRawClient implements DatabricksRawClient but is backed by json files in testdata.
type testdataDBRawClient struct {
	i            int
	testDataDir  string
	multiCluster bool
}

func (c *testdataDBRawClient) jobsList(limit int, offset int) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("jobs-list-%d.json", offset/limit)))
}

func (c *testdataDBRawClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("active-job-runs-%d.json", offset/limit)))
}

func (c *testdataDBRawClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	if jobID != TestdataJobID {
		return []byte("{}"), nil
	}
	if offset == 0 {
		c.i++
	}
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("completed-job-runs-%d-%d.json", c.i-1, offset/limit)))
}

func (c *testdataDBRawClient) clustersList() ([]byte, error) {
	if c.multiCluster {
		return os.ReadFile(filepath.Join(c.testDataDir, "clusters-list-multi.json"))
	}
	return os.ReadFile(filepath.Join(c.testDataDir, "clusters-list.json"))
}

func (c *testdataDBRawClient) pipelines() ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, "pipelines.json"))
}

func (c *testdataDBRawClient) pipeline(s string) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, "pipeline.json"))
}

func MetricsByName(pm pmetric.Metrics) map[string]pmetric.Metric {
	out := map[string]pmetric.Metric{}
	for i := 0; i < pm.ResourceMetrics().Len(); i++ {
		sms := pm.ResourceMetrics().At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			ms := sms.At(j).Metrics()
			for k := 0; k < ms.Len(); k++ {
				metric := ms.At(k)
				out[metric.Name()] = metric
			}
		}
	}
	return out
}

type fakeDatabricksRestService struct {
	runs []JobRun
	i    int
}

func (c *fakeDatabricksRestService) jobs() (out []Job, err error) {
	return nil, nil
}

func (c *fakeDatabricksRestService) activeJobRuns() (out []JobRun, err error) {
	return nil, nil
}

func (c *fakeDatabricksRestService) CompletedJobRuns(jobID int, _ int64) ([]JobRun, error) {
	c.addCompletedRun(jobID)
	return c.runs, nil
}

func (c *fakeDatabricksRestService) addCompletedRun(jobID int) {
	c.runs = append([]JobRun{{
		JobID:             jobID,
		StartTime:         1_600_000_000_000 + (1_000_000 * int64(c.i)),
		ExecutionDuration: 15_000 + (1000 * c.i),
	}}, c.runs...)
	c.i++
}

func (c *fakeDatabricksRestService) RunningClusters() ([]spark.Cluster, error) {
	return nil, nil
}

func (c *fakeDatabricksRestService) RunningPipelines() ([]spark.PipelineSummary, error) {
	return nil, nil
}
