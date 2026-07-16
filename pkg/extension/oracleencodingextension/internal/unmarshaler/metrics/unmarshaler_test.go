// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

func TestUnmarshalMetrics(t *testing.T) {
	buf, err := os.ReadFile(filepath.Join("testdata", "metrics.jsonl"))
	require.NoError(t, err)

	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics(buf)
	require.NoError(t, err)

	// Two distinct resources: (namespace, compartmentId, resourceGroup, resourceId) tuples.
	require.Equal(t, 2, md.ResourceMetrics().Len())

	var firstRM, secondRM = md.ResourceMetrics().At(0), md.ResourceMetrics().At(1)
	if v, _ := firstRM.Resource().Attributes().Get(oracleCloudNamespaceKey); v.AsString() != "myFirstNamespace" {
		firstRM, secondRM = secondRM, firstRM
	}

	provider, ok := firstRM.Resource().Attributes().Get(string(conventions.CloudProviderKey))
	require.True(t, ok)
	require.Equal(t, conventions.CloudProviderOracleCloud.Value.AsString(), provider.AsString())

	compartmentID, ok := firstRM.Resource().Attributes().Get(oracleCloudCompartmentIDKey)
	require.True(t, ok)
	require.Equal(t, "ocid1.compartment.oc1..exampleuniqueID", compartmentID.AsString())

	resourceID, ok := firstRM.Resource().Attributes().Get(string(conventions.CloudResourceIDKey))
	require.True(t, ok)
	require.Equal(t, "ocid1.exampleresource.region1.phx.exampleuniqueID", resourceID.AsString())

	ns, ok := firstRM.Resource().Attributes().Get(oracleCloudNamespaceKey)
	require.True(t, ok)
	require.Equal(t, "myFirstNamespace", ns.AsString())
	rg, ok := firstRM.Resource().Attributes().Get(oracleCloudResourceGroupKey)
	require.True(t, ok)
	require.Equal(t, "myFirstResourceGroup", rg.AsString())
	realm, ok := firstRM.Resource().Attributes().Get(oracleCloudRealmKey)
	require.True(t, ok)
	require.Equal(t, "oc1", realm.AsString())

	scopeMetrics := firstRM.ScopeMetrics().At(0)
	require.Equal(t, ScopeName, scopeMetrics.Scope().Name())
	require.Equal(t, 2, scopeMetrics.Metrics().Len())

	successRate := scopeMetrics.Metrics().At(0)
	require.Equal(t, "successRate", successRate.Name())
	require.Equal(t, "percent", successRate.Unit())
	require.Equal(t, "MyAppA Success Rate", successRate.Description())
	require.Equal(t, pmetric.MetricTypeGauge, successRate.Type())
	require.Equal(t, 2, successRate.Gauge().DataPoints().Len())

	dp0 := successRate.Gauge().DataPoints().At(0)
	require.Equal(t, 83.0, dp0.DoubleValue())
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1784285626872)), dp0.Timestamp())

	// resourceId is promoted to the Resource as cloud.resource_id, but is
	// also kept as a datapoint attribute.
	resourceIDAttr, ok := dp0.Attributes().Get(dimensionResourceID)
	require.True(t, ok)
	require.Equal(t, "ocid1.exampleresource.region1.phx.exampleuniqueID", resourceIDAttr.AsString())
	appName, ok := dp0.Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myAppA", appName.AsString())

	// Resource without a resourceGroup.
	_, ok = secondRM.Resource().Attributes().Get(oracleCloudResourceGroupKey)
	require.False(t, ok)
	require.Equal(t, 1, secondRM.ScopeMetrics().At(0).Metrics().Len())
}

func TestUnmarshalMetrics_GaugeUnit(t *testing.T) {
	for _, unit := range []string{"", "percent", "bytes", "milliseconds", "sum", "count", "Sum", "COUNT"} {
		t.Run(unit, func(t *testing.T) {
			input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"` + unit + `"},"datapoints":[{"timestamp":1673388760000,"value":42.0}]}
`
			u := NewResourceMetricsUnmarshaler(zap.NewNop())
			md, err := u.UnmarshalMetrics([]byte(input))
			require.NoError(t, err)
			require.Equal(t, 1, md.ResourceMetrics().Len())

			m := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
			require.Equal(t, pmetric.MetricTypeGauge, m.Type())
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, 42.0, m.Gauge().DataPoints().At(0).DoubleValue())
		})
	}
}

func TestUnmarshalMetrics_InvalidTimestamp(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"m","datapoints":[{"timestamp":"not-a-time","value":1.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_SkipsZeroTimestampDatapoint(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"m","datapoints":[{"timestamp":0,"value":1.0},{"timestamp":1673388760000,"value":2.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 1, md.ResourceMetrics().Len())

	dps := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints()
	require.Equal(t, 1, dps.Len())
	require.Equal(t, 2.0, dps.At(0).DoubleValue())
}

func TestUnmarshalMetrics_SkipsRecordWithOnlyZeroTimestampDatapoints(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"m","datapoints":[{"timestamp":0,"value":1.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_InvalidJSON(t *testing.T) {
	input := "not json\n"
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_MissingName(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","datapoints":[{"timestamp":"2023-01-10T22:19:20Z","value":1.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_MissingCompartmentID(t *testing.T) {
	input := `{"namespace":"ns","resourceGroup":"rg","name":"m","datapoints":[{"timestamp":"2023-01-10T22:19:20Z","value":1.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_MissingNamespace(t *testing.T) {
	input := `{"compartmentId":"c1","resourceGroup":"rg","name":"m","datapoints":[{"timestamp":"2023-01-10T22:19:20Z","value":1.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_NoDatapoints(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"m","datapoints":[]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestUnmarshalMetrics_SkipsResourceWithOnlyEmptyRecord(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"empty","datapoints":[]}
{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"valid","datapoints":[{"timestamp":1673388760000,"value":42.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 1, md.ResourceMetrics().Len())

	scopeMetrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	require.Equal(t, 1, scopeMetrics.Metrics().Len())
	require.Equal(t, "valid", scopeMetrics.Metrics().At(0).Name())
}

func TestUnmarshalMetrics_MergesSameMetricIdentity(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"successRate","dimensions":{"appName":"myAppA"},"metadata":{"unit":"percent","displayName":"Success rate"},"datapoints":[{"timestamp":1673388760000,"value":83.0}]}
{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"successRate","dimensions":{"appName":"myAppB"},"metadata":{"unit":"percent","displayName":"Application success rate"},"datapoints":[{"timestamp":1673388761000,"value":90.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 1, md.ResourceMetrics().Len())

	scopeMetrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0)
	require.Equal(t, 1, scopeMetrics.Metrics().Len())

	m := scopeMetrics.Metrics().At(0)
	require.Equal(t, "successRate", m.Name())
	require.Equal(t, "percent", m.Unit())
	require.Equal(t, "Application success rate", m.Description())
	require.Equal(t, 2, m.Gauge().DataPoints().Len())

	appName0, ok := m.Gauge().DataPoints().At(0).Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myAppA", appName0.AsString())
	appName1, ok := m.Gauge().DataPoints().At(1).Attributes().Get("appName")
	require.True(t, ok)
	require.Equal(t, "myAppB", appName1.AsString())
}

func TestUnmarshalMetrics_SortsDatapointsByTimestampAscending(t *testing.T) {
	// Datapoints are out of order both within a single record and across
	// separate records merged into the same metric.
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388762000,"value":3.0},{"timestamp":1673388760000,"value":1.0}]}
{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388761000,"value":2.0}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 1, md.ResourceMetrics().Len())

	m := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	dps := m.Gauge().DataPoints()
	require.Equal(t, 3, dps.Len())

	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388760000)), dps.At(0).Timestamp())
	require.Equal(t, 1.0, dps.At(0).DoubleValue())
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388761000)), dps.At(1).Timestamp())
	require.Equal(t, 2.0, dps.At(1).DoubleValue())
	require.Equal(t, pcommon.NewTimestampFromTime(time.UnixMilli(1673388762000)), dps.At(2).Timestamp())
	require.Equal(t, 3.0, dps.At(2).DoubleValue())
}

func TestUnmarshalMetrics_DoesNotMergeConflictingUnits(t *testing.T) {
	input := `{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"ms"},"datapoints":[{"timestamp":1673388760000,"value":42.0}]}
{"namespace":"ns","compartmentId":"c1","resourceGroup":"rg","name":"latency","metadata":{"unit":"s"},"datapoints":[{"timestamp":1673388761000,"value":0.043}]}
`
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(input))
	require.NoError(t, err)
	require.Equal(t, 1, md.ResourceMetrics().Len())

	metrics := md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	require.Equal(t, 2, metrics.Len())

	metricsByUnit := make(map[string]pmetric.Metric, metrics.Len())
	for i := 0; i < metrics.Len(); i++ {
		metricsByUnit[metrics.At(i).Unit()] = metrics.At(i)
	}
	require.Contains(t, metricsByUnit, "ms")
	require.Contains(t, metricsByUnit, "s")
	require.Equal(t, 42.0, metricsByUnit["ms"].Gauge().DataPoints().At(0).DoubleValue())
	require.Equal(t, 0.043, metricsByUnit["s"].Gauge().DataPoints().At(0).DoubleValue())
}

func TestUnmarshalMetrics_Empty(t *testing.T) {
	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics([]byte(""))
	require.NoError(t, err)
	require.Equal(t, 0, md.ResourceMetrics().Len())
}

func TestExtractRealm(t *testing.T) {
	tests := []struct {
		name string
		ocid string
		want string
	}{
		{name: "compartment OCID", ocid: "ocid1.compartment.oc1..exampleuniqueID", want: "oc1"},
		{name: "instance OCID with region", ocid: "ocid1.instance.oc1.phx.abuw4ljrlsfiqw6vzzxb43vyypt4pkodawglp3wqxjqofakrwvou52gb6s5a", want: "oc1"},
		{name: "government realm", ocid: "ocid1.tenancy.oc2..exampleuniqueID", want: "oc2"},
		{name: "empty", ocid: "", want: ""},
		{name: "not an OCID", ocid: "not-an-ocid", want: ""},
		{name: "future OCID version", ocid: "ocid2.compartment.oc1..exampleuniqueID", want: "oc1"},
		{name: "too few segments", ocid: "ocid1.compartment", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, extractRealm(tt.ocid))
		})
	}
}
