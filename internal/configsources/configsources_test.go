// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package configsources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestConfigSourcesGet(t *testing.T) {
	tests := []struct {
		configSourceType config.Type
	}{
		{"env"},
		{"etcd2"},
		{"include"},
		{"vault"},
		{"zookeeper"},
	}

	defaultCfgSrcFactories := Get()
	require.Equal(t, len(tests), len(defaultCfgSrcFactories))

	cfgSrcFactoryMap := make(map[config.Type]struct{})
	for _, tt := range tests {
		t.Run(string(tt.configSourceType), func(t *testing.T) {
			var factory configprovider.Factory
			for _, f := range defaultCfgSrcFactories {
				if f.Type() == tt.configSourceType {
					// Ensure no duplicated factories.
					if _, ok := cfgSrcFactoryMap[tt.configSourceType]; ok {
						assert.Fail(t, "duplicated config source factory")
					}
					cfgSrcFactoryMap[f.Type()] = struct{}{}
					factory = f
					break
				}
			}

			require.NotNil(t, factory, "missing or nil config source factory")
		})
	}
}
