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
	"errors"
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

// erroringReader returns a fixed line of data followed by a non-io.EOF error,
// simulating a failure in the underlying io.Reader (e.g. a broken pipe)
type erroringReader struct {
	err  error
	data []byte
	read bool
}

func (r *erroringReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, r.err
	}
	r.read = true
	n := copy(p, r.data)
	return n, nil
}

func TestUnmarshalMetrics(t *testing.T) {
	buf, err := os.ReadFile(filepath.Join("testdata", "metrics.jsonl"))
	require.NoError(t, err)

	u := NewResourceMetricsUnmarshaler(zap.NewNop())
	md, err := u.UnmarshalMetrics(buf)
	require.NoError(t, err)

	// Two distinct resources: (namespace, compartmentId, resourceGroup, resourceId) tuples.
	require.Equal(t, 2, md.ResourceMetrics().Len())

	firstRM, secondRM := md.ResourceMetrics().At(0), md.ResourceMetrics().At(1)
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
	require.InDelta(t, 83.0, dp0.DoubleValue(), .001)
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

func TestReadJSONLLines_ReturnsNonEOFReadError(t *testing.T) {
	wantErr := errors.New("broken pipe")
	r := &erroringReader{data: []byte(`{"name":"m"}` + "\n"), err: wantErr}

	var lines [][]byte
	err := readJSONLLines(r, func(line []byte) {
		lines = append(lines, append([]byte{}, line...))
	})

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, [][]byte{[]byte(`{"name":"m"}`)}, lines)
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
