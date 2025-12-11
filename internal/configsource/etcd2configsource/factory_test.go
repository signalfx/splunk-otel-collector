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

package etcd2configsource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestEtcd2Factory_CreateConfigSource(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, component.MustNewType("etcd2"), factory.Type())
	tests := []struct {
		wantErr error
		config  *Config
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
				Endpoints: []string{"some\bad/endpoint"},
			},
			wantErr: &errInvalidEndpoint{},
		},
		{
			name: "missing_token",
			config: &Config{
				Endpoints: []string{"http://localhost:8200", "bad endpoint"},
			},
			wantErr: &errInvalidEndpoint{},
		},
		{
			name: "success",
			config: &Config{
				Endpoints: []string{"http://localhost:8200"},
				Authentication: &Authentication{
					Username: "user",
					Password: "pass",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := factory.CreateConfigSource(context.Background(), tt.config, zap.NewNop())
			if tt.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantErr == nil {
				assert.NotNil(t, actual)
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}
