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

package splunksecretconfigsource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/modularinput"
)

func TestSplunkSecretFactory_CreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	assert.Equal(t, defaultTimeout, cfg.Timeout)
	assert.False(t, cfg.InsecureSkipVerify)
	assert.Empty(t, cfg.Endpoint)
	assert.Empty(t, cfg.SessionKey)
	assert.Equal(t, defaultApp, cfg.App)
	assert.Equal(t, defaultUser, cfg.User)
}

func TestSplunkSecretFactory_CreateDefaultConfig_FromModularInputEnv(t *testing.T) {
	t.Setenv(modularinput.EnvManagementURI, "https://127.0.0.1:8089")
	t.Setenv(modularinput.EnvSessionKey, "some_session_key")

	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	assert.Equal(t, "https://127.0.0.1:8089", cfg.Endpoint)
	assert.Equal(t, "some_session_key", cfg.SessionKey)
}

func TestSplunkSecretFactory_CreateConfigSource(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, component.MustNewType(typeStr), factory.Type())
	tests := []struct {
		wantErr error
		config  *Config
		name    string
	}{
		{
			name:    "missing_endpoint",
			config:  &Config{SessionKey: "key"},
			wantErr: &errMissingEndpoint{},
		},
		{
			name:    "invalid_endpoint",
			config:  &Config{Endpoint: "some\bad/endpoint", SessionKey: "key"},
			wantErr: &errInvalidEndpoint{},
		},
		{
			name:    "missing_session_key",
			config:  &Config{Endpoint: "https://localhost:8089"},
			wantErr: &errMissingSessionKey{},
		},
		{
			name:    "negative_timeout",
			config:  &Config{Endpoint: "https://localhost:8089", SessionKey: "key", Timeout: -1 * time.Second},
			wantErr: &errNonPositiveTimeout{},
		},
		{
			name:   "success",
			config: &Config{Endpoint: "https://localhost:8089", SessionKey: "key"},
		},
		{
			name:   "success_with_insecure_skip_verify",
			config: &Config{Endpoint: "https://localhost:8089", SessionKey: "key", InsecureSkipVerify: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := factory.CreateConfigSource(t.Context(), tt.config, zap.NewNop())
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, actual)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, actual)
			}
		})
	}
}
