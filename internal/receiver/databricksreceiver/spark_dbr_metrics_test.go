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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

func TestSparkDbrMetrics_Append(t *testing.T) {
	outerSDM := newSparkDbrMetrics()
	c := cluster{ClusterID: "my-cluster-id", ClusterName: "my-cluster-name", State: "my-cluster-state"}

	sdmSub1 := newSparkDbrMetrics()
	sdmSub1.addCounter(c, "my-app-id", spark.Counter{Count: 42}, sparkMetricBase{
		partialMetricName: "databricks.directorycommit.autovacuumcount",
		pipelineID:        "my-pipeline-id",
		pipelineName:      "my-pipeline-name",
	})
	outerSDM.append(sdmSub1)

	sdmSub2 := newSparkDbrMetrics()
	sdmSub2.addCounter(c, "my-app-id", spark.Counter{Count: 111}, sparkMetricBase{
		partialMetricName: "databricks.directorycommit.deletedfilesfiltered",
		pipelineID:        "my-pipeline-id",
		pipelineName:      "my-pipeline-name",
	})
	outerSDM.append(sdmSub2)

	builder := newTestMetricsBuilder()
	emitted := outerSDM.build(builder, pcommon.NewTimestampFromTime(time.Now()))
	allMetrics := pmetric.NewMetrics()
	for _, metrics := range emitted {
		metrics.ResourceMetrics().CopyTo(allMetrics.ResourceMetrics())
	}
	assert.Equal(t, 2, allMetrics.MetricCount())
}
