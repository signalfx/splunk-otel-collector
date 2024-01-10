// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ecs

import (
	"testing"

	"github.com/signalfx/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/signalfx-agent/pkg/core/common/ecs"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/neotest"
)

func TestDimensionToUpdate(t *testing.T) {
	ctr := ecs.Container{
		Name:     "some.container.name",
		DockerID: "some.container.id",
	}

	for _, test := range []struct {
		configValue      string
		expectedDimKey   string
		expectedDimValue string
		expectedError    string
	}{
		{
			configValue:      "container_name",
			expectedDimKey:   "container_name",
			expectedDimValue: "some.container.name",
		},
		{
			configValue:      "container_id",
			expectedDimKey:   "container_id",
			expectedDimValue: "some.container.id",
		},
		{
			configValue:      "default",
			expectedDimKey:   "container_id",
			expectedDimValue: "some.container.id",
		},
		{
			configValue:   "invalid",
			expectedError: "unsupported `dimensionToUpdate` \"invalid\". Must be one of \"container_name\" or \"container_id\"",
		},
	} {
		test := test
		t.Run(test.configValue, func(t *testing.T) {
			cfg := &Config{
				MetadataEndpoint: "http://localhost:0/not/real",
				MonitorConfig: config.MonitorConfig{
					IntervalSeconds: 1000,
				},
			}
			if test.configValue != "default" {
				cfg.DimensionToUpdate = test.configValue
			}
			require.NoError(t, defaults.Set(cfg))
			monitor := &Monitor{
				Output: neotest.NewTestOutput(),
			}

			err := monitor.Configure(cfg)
			t.Cleanup(func() {
				if monitor.cancel != nil {
					monitor.cancel()
				}
			})
			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
			key, value := monitor.dimensionToUpdate(ctr)
			assert.Equal(t, test.expectedDimKey, key)
			assert.Equal(t, test.expectedDimValue, value)
		})
	}
}
