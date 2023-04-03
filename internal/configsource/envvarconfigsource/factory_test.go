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

package envvarconfigsource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestEnvVarConfigSourceFactory_CreateConfigSource(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, component.Type("env"), factory.Type())
	tests := []struct {
		config *Config
		name   string
	}{
		{
			name:   "no_defaults",
			config: &Config{},
		},
		{
			name: "with_defaults",
			config: &Config{
				Defaults: map[string]any{
					"k0": "v0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := factory.CreateConfigSource(context.Background(), tt.config, zap.NewNop())
			assert.NoError(t, err)
			assert.NotNil(t, actual)
		})
	}
}
