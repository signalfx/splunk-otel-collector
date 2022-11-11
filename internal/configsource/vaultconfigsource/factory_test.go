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

package vaultconfigsource

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestVaultFactory_CreateConfigSource(t *testing.T) {
	emptyStr := ""
	factory := NewFactory()
	assert.Equal(t, component.Type("vault"), factory.Type())
	createParams := configprovider.CreateParams{
		Logger: zap.NewNop(),
	}
	tests := []struct {
		config  *Config
		wantErr error
		name    string
	}{
		{
			name:    "missing_endpoint",
			config:  &Config{},
			wantErr: &errMissingEndpoint{},
		},
		{
			name: "invalid_endpoint",
			config: &Config{
				Endpoint: "some\bad/endpoint",
			},
			wantErr: &errInvalidEndpoint{},
		},
		{
			name: "missing_auth",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Path:     "some/path",
			},
			wantErr: &errMissingAuthentication{},
		},
		{
			name: "empty_auth",
			config: &Config{
				Endpoint:       "http://localhost:8200",
				Path:           "some/path",
				Authentication: &Authentication{},
			},
			wantErr: &errEmptyAuth{},
		},
		{
			name: "multiple_auth_methods",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Path:     "some/path",
				Authentication: &Authentication{
					Token:             &tokenStr,
					IAMAuthentication: &IAMAuthentication{},
					GCPAuthentication: &GCPAuthentication{},
				},
			},
			wantErr: &errMultipleAuthMethods{},
		},
		{
			name: "empty_token",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Path:     "some/path",
				Authentication: &Authentication{
					Token: &emptyStr,
				},
			},
			wantErr: &errEmptyToken{},
		},
		{
			name: "missing_path",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Authentication: &Authentication{
					Token: &tokenStr,
				},
			},
			wantErr: &errMissingPath{},
		},
		{
			name: "invalid_poll_interval",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Authentication: &Authentication{
					Token: &tokenStr,
				},
				Path: "some/path",
			},
			wantErr: &errNonPositivePollInterval{},
		},
		{
			name: "success",
			config: &Config{
				Endpoint: "http://localhost:8200",
				Authentication: &Authentication{
					Token: &tokenStr,
				},
				Path:         "some/path",
				PollInterval: 2 * time.Minute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := factory.CreateConfigSource(context.Background(), createParams, tt.config)
			require.IsType(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.NotNil(t, actual)
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}
