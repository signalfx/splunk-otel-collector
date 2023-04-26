// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testdata

import (
	"testing"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicNoMd(t *testing.T) {
	wqs := GetWriteRequestsOfAllTypesWithoutMetadata()
	require.NotNil(t, wqs)
	for _, wq := range wqs {
		for _, ts := range wq.Timeseries {
			require.NotNil(t, ts)
			assert.NotEmpty(t, ts.Labels)
		}
		require.Empty(t, wq.Metadata)
	}
}

func TestCoveringMd(t *testing.T) {
	wq := FlattenWriteRequests(GetWriteRequestsWithMetadata())
	require.NotNil(t, wq)
	for _, ts := range wq.Timeseries {
		require.NotNil(t, ts)
		assert.NotEmpty(t, ts.Labels)
	}
	require.NotEmpty(t, wq.Metadata)
	total := 0
	unknown := 0
	for _, ts := range wq.Metadata {
		total += 1
		require.NotNil(t, ts)
		assert.NotEmpty(t, ts.Type)
		if ts.Type == prompb.MetricMetadata_UNKNOWN {
			unknown += 1
		}
		assert.NotEmpty(t, ts.MetricFamilyName)
		assert.Equal(t, ts.Unit, "unit")
		assert.NotEmpty(t, ts.Help)
	}
	assert.Equal(t, total, total-unknown)
}
