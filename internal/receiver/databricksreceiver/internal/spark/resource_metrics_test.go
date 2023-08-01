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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/commontest"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

func TestSparkDbrMetrics_Append(t *testing.T) {
	outerRM := NewResourceMetrics()
	c := Cluster{ClusterID: "my-cluster-id", ClusterName: "my-cluster-name", State: "my-cluster-state"}

	rmSub1 := NewResourceMetrics()
	rmSub1.addCounter(c, "my-app-id", Counter{Count: 42}, sparkMetricBase{
		partialMetricName: "databricks.directorycommit.autovacuumcount",
		pipelineID:        "my-pipeline-id",
		pipelineName:      "my-pipeline-name",
	})
	outerRM.Append(rmSub1)

	rmSub2 := NewResourceMetrics()
	rmSub2.addCounter(c, "my-app-id", Counter{Count: 111}, sparkMetricBase{
		partialMetricName: "databricks.directorycommit.deletedfilesfiltered",
		pipelineID:        "my-pipeline-id",
		pipelineName:      "my-pipeline-name",
	})
	outerRM.Append(rmSub2)

	mb := commontest.NewTestMetricsBuilder()
	rb := metadata.NewResourceBuilder(metadata.DefaultResourceAttributesConfig())
	built := outerRM.Build(mb, rb, pcommon.NewTimestampFromTime(time.Now()), "my-app-id")
	allMetrics := pmetric.NewMetrics()
	for _, metrics := range built {
		metrics.ResourceMetrics().CopyTo(allMetrics.ResourceMetrics())
	}
	assert.Equal(t, 2, allMetrics.MetricCount())
}
