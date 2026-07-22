// Copyright  Splunk, Inc.
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

package rollingspanlatencyprocessor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected *Config
	}{
		{
			id: component.MustNewID("rollingspanlatency"),
			expected: &Config{
				HalfLife:              2 * time.Hour,
				SlowThreshold:         3.0,
				VerySlowThreshold:     4.0,
				AttributeKey:          "latency.category",
				ResourceKeyAttributes: []string{"service.namespace", "service.name", "deployment.environment.name"},
				IdleTimeout:           8 * time.Hour,
				EvictionInterval:      10 * time.Minute,
				MaxBaselines:          1000,
				ChurnWarningRatio:     0.5,
				WarmupCount:           30,
				MinStddev:             1e6,
			},
		},
		{
			id: component.MustNewIDWithName("rollingspanlatency", "custom"),
			expected: &Config{
				HalfLife:              30 * time.Second,
				SlowThreshold:         2.0,
				VerySlowThreshold:     3.0,
				AttributeKey:          "span.latency_tier",
				ResourceKeyAttributes: []string{"service.name"},
				IdleTimeout:           5 * time.Minute,
				EvictionInterval:      time.Minute,
				MaxBaselines:          500,
				ChurnWarningRatio:     0.25,
				WarmupCount:           10,
				MinStddev:             5e6,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()
			sub, err := cm.Sub(tc.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))
			assert.Equal(t, tc.expected, cfg)
		})
	}
}
