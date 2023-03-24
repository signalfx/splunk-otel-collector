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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

const testdataJobID = 288

func NewTestSingleClusterService(testdataDir string) Service {
	return NewService(&testdataRawClient{
		testDataDir: testdataDir,
	}, 25)
}

func NewTestMultiClusterService(testdataDir string) Service {
	return NewService(&testdataRawClient{
		testDataDir:  testdataDir,
		multiCluster: true,
	}, 25)
}

// testdataRawClient implements RawClient but is backed by json files in testdata.
type testdataRawClient struct {
	testDataDir  string
	i            int
	multiCluster bool
}

func (c *testdataRawClient) jobsList(limit int, offset int) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("jobs-list-%d.json", offset/limit)))
}

func (c *testdataRawClient) activeJobRuns(limit int, offset int) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("active-job-runs-%d.json", offset/limit)))
}

func (c *testdataRawClient) completedJobRuns(jobID int, limit int, offset int) ([]byte, error) {
	if jobID != testdataJobID {
		return []byte("{}"), nil
	}
	if offset == 0 {
		c.i++
	}
	return os.ReadFile(filepath.Join(c.testDataDir, fmt.Sprintf("completed-job-runs-%d-%d.json", c.i-1, offset/limit)))
}

func (c *testdataRawClient) clustersList() ([]byte, error) {
	if c.multiCluster {
		return os.ReadFile(filepath.Join(c.testDataDir, "clusters-list-multi.json"))
	}
	return os.ReadFile(filepath.Join(c.testDataDir, "clusters-list.json"))
}

func (c *testdataRawClient) pipelines() ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, "pipelines.json"))
}

func (c *testdataRawClient) pipeline(_ string) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.testDataDir, "pipeline.json"))
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
