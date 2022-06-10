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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.opentelemetry.io/collector/confmap"
)

func TestEnvVarConfigSource_Session(t *testing.T) {
	const testEnvVarName = "_TEST_ENV_VAR_CFG_SRC"
	const testEnvVarValue = "test_env_value"

	tests := []struct {
		defaults map[string]interface{}
		params   map[string]interface{}
		expected interface{}
		wantErr  error
		name     string
		selector string
	}{
		{
			name:     "simple",
			selector: testEnvVarName,
			expected: testEnvVarValue,
		},
		{
			name:     "missing_not_required",
			selector: "UNDEFINED_ENV_VAR",
			params: map[string]interface{}{
				"optional": true,
			},
			expected: nil,
		},
		{
			name: "invalid_param",
			params: map[string]interface{}{
				"unknow_params_field": true,
			},
			wantErr: &errInvalidRetrieveParams{},
		},
		{
			name:     "missing_required",
			selector: "UNDEFINED_ENV_VAR",
			wantErr:  &errMissingRequiredEnvVar{},
		},
		{
			name: "required_on_defaults",
			defaults: map[string]interface{}{
				"FALLBACK_ENV_VAR": "fallback_env_var",
			},
			selector: "FALLBACK_ENV_VAR",
			expected: "fallback_env_var",
		},
	}

	require.NoError(t, os.Setenv(testEnvVarName, testEnvVarValue))
	t.Cleanup(func() {
		assert.NoError(t, os.Unsetenv(testEnvVarName))
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := tt.defaults
			if defaults == nil {
				defaults = make(map[string]interface{})
			}

			source := &envVarConfigSource{
				defaults: defaults,
			}

			ctx := context.Background()
			defer func() {
				assert.NoError(t, source.Close(ctx))
			}()

			r, err := source.Retrieve(ctx, tt.selector, confmap.NewFromStringMap(tt.params))
			if tt.wantErr != nil {
				assert.Nil(t, r)
				require.IsType(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tt.expected, r.Value())
			_, ok := r.(configsource.Watchable)
			assert.False(t, ok)
		})
	}
}
