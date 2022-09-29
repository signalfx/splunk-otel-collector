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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsProvider_Scrape(t *testing.T) {
	const ignored = 25
	c := newDatabricksClient(&testdataClient{}, ignored)
	scrpr := scraper{
		instanceName: "my-instance",
		mp:           newMetricsProvider(c),
		rmp:          newRunMetricsProvider(c),
	}
	metrics, err := scrpr.scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 6, metrics.MetricCount())
	attrs := metrics.ResourceMetrics().At(0).Resource().Attributes()
	v, _ := attrs.Get("databricks.instance.name")
	assert.Equal(t, "my-instance", v.Str())
}
